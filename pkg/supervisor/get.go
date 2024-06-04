package supervisor

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/rs/zerolog"
	"github.com/tvs/sshit"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/tvs/ultravisor/pkg/config"
)

type SupervisorInfo struct {
	ControlPlane string   `json:"controlPlane,omitempty" yaml:"controlPlane,omitempty"`
	VMs          []string `json:"vms,omitempty" yaml:"vms,omitempty"`
	Password     string   `json:"password,omitempty" yaml:"password,omitempty"`
}

func Info(ctx context.Context) (*SupervisorInfo, error) {
	l := zerolog.Ctx(ctx)
	c := config.Ctx(ctx)

	l.Debug().Interface("config", c).Msg("beginning info retrieval")

	if err := ValidateConfig(c); err != nil {
		l.Error().Err(err).Any("config", c).Msg("invalid config")
		return nil, err
	}

	var jumpbox *sshit.Client
	if c.JumpboxConfig != nil {
		cfg, err := c.JumpboxConfig.ClientConfig()
		if err != nil {
			return nil, err
		}
		jumpbox = &sshit.Client{
			Config: cfg,
			Server: sshit.Endpoint{
				Host: c.JumpboxConfig.Host,
				Port: *c.JumpboxConfig.Port,
			},
		}

		if err := jumpbox.Connect(ctx); err != nil {
			l.Error().Err(err).Msg("unable to initiate jumpbox connection")
			return nil, fmt.Errorf("unable to initiate jumpbox connection: %w", err)
		}

		defer func() {
			if tErr := jumpbox.Close(); tErr != nil {
				l.Error().Err(tErr).Msg("unable to close jumpbox session")
				if err == nil {
					err = fmt.Errorf("unable to close jumpbox session: %w", tErr)
				}
			}
			jumpbox.Close()
		}()
	}

	vms, err := getSupervisorVMs(ctx, c, jumpbox)
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve Supervisor VMs")
		return nil, fmt.Errorf("unable to retrieve Supervisor VMs: %w", err)
	}

	controlPlane, password, err := getSupervisorCredentials(ctx, c, jumpbox)
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve Supervisor credentials")
		return nil, fmt.Errorf("unable to retrieve Supervisor credentials: %w", err)
	}

	return &SupervisorInfo{
		ControlPlane: controlPlane,
		VMs:          vms,
		Password:     password,
	}, nil
}

// TODO(tvs): Figure out a nice way to send back and log each invalidity
// errors.Join() doesn't work well with default zerolog for readability
// and log.Errs() doesn't use the colorized error handler
func ValidateConfig(c *config.Config) error {
	if c.JumpboxConfig != nil {
		if err := validateSSHConfig(c.JumpboxConfig); err != nil {
			return err
		}
	}

	if err := validateVCenterConfig(c.VCenterConfig); err != nil {
		return err
	}

	return nil
}

func validateSSHConfig(c *config.SSHConfig) error {
	if c.Host == "" {
		return fmt.Errorf("host must be supplied")
	}

	if _, err := url.Parse(c.Host); err != nil {
		return fmt.Errorf("unable to parse server: %w", err)
	}

	if c.Port == nil {
		return fmt.Errorf("port must be supplied")
	}

	if c.User == "" {
		return fmt.Errorf("SSH user must be supplied")
	}

	if c.Key == nil && c.KeyPath == nil && c.Password == nil {
		return fmt.Errorf("one of key, keyPath, or password must be supplied")
	}

	return nil
}

// TODO(tvs): Validate the vCenter config: sso user, etc.
func validateVCenterConfig(c *config.VCenterConfig) error {
	if c == nil {
		return fmt.Errorf("vcenter config must be supplied")
	}

	return validateSSHConfig(c.SSH)
}

func getSupervisorVMs(ctx context.Context, c *config.Config, jumpbox *sshit.Client) (_ []string, err error) {
	l := zerolog.Ctx(ctx)

	// Start tunnel if we need it
	var endpoint sshit.Endpoint
	if c.JumpboxConfig != nil {
		tunnel := sshit.NewForwardTunnel(ctx,
			sshit.Endpoint{Host: "localhost", Port: 0},
			sshit.Endpoint{Host: c.VCenterConfig.SSH.Host, Port: 443})

		if err = tunnel.Bind(jumpbox); err != nil {
			l.Error().Err(err).Msg("unable to establish tunnel to vCenter")
			return nil, fmt.Errorf("unable to establish tunnel to vCenter: %w", err)
		}

		defer func() {
			if tErr := tunnel.Close(); tErr != nil {
				l.Error().Errs("err", tErr).Msg("unable to close tunnel to vCenter")
				if err == nil {
					err = errors.Join(tErr...)
				}
			}
		}()

		endpoint = sshit.Endpoint{Host: tunnel.Local().Host, Port: tunnel.Local().Port}
	} else {
		endpoint = sshit.Endpoint{Host: c.VCenterConfig.SSH.Host, Port: 443}
	}

	// Set up a session so we can log out once we're done
	l.Debug().
		Str("address", endpoint.Address()).
		Str("user", c.VCenterConfig.SSO.User).
		Str("password", c.VCenterConfig.SSO.Password).
		Msg("establishing vCenter session")
	session, err := vcSession(endpoint, c.VCenterConfig.SSO.User, c.VCenterConfig.SSO.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to create a VC session: %w", err)
	}

	client := new(vim25.Client)
	err = session.Login(ctx, client, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a vim client: %w", err)
	}

	defer func() {
		if tErr := session.Logout(ctx, client); err != nil {
			l.Error().Err(tErr).Msg("unable to log out of VC session")
			if err == nil {
				err = fmt.Errorf("unable to log out of VC session: %w", err)
			}
		}
	}()

	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)

	objs, err := v.Find(ctx, []string{"VirtualMachine"}, property.Match{"name": "SupervisorControlPlaneVM*"})
	if err != nil {
		return nil, err
	}

	if len(objs) < 1 {
		return nil, fmt.Errorf("no supervisor control plane VMs were found")
	}

	p := property.DefaultCollector(client)

	var ipAddrs []string
	for _, o := range objs {
		filter := new(property.WaitFilter)
		filter.Add(o, o.Type, []string{"guest.net"})
		req := types.RetrieveProperties{
			SpecSet: []types.PropertyFilterSpec{filter.Spec},
		}

		res, err := p.RetrieveProperties(ctx, req)
		if err != nil {
			return nil, err
		}
		content := res.Returnval
		if len(content) != 1 {
			return nil, fmt.Errorf("%d objects match", len(content))
		}
		obj, err := mo.ObjectContentToType(content[0])
		if err != nil {
			return nil, err
		}

		addr := getVMNetworkPreferredIP(obj.(mo.VirtualMachine))
		if addr == "" {
			return nil, fmt.Errorf("object did not have a VM Network preferred IP address")
		}

		ipAddrs = append(ipAddrs, addr)
	}

	return ipAddrs, err
}

func getVMNetworkPreferredIP(vm mo.VirtualMachine) string {
	for _, nic := range vm.Guest.Net {
		if nic.Network == "VM Network" {
			for _, addr := range nic.IpConfig.IpAddress {
				if addr.State == "preferred" {
					return addr.IpAddress
				}
			}
		}
	}

	return ""
}

func vcSession(endpoint sshit.Endpoint, username, password string) (*cache.Session, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s/sdk", endpoint.Address()))
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(username, password)

	// TODO(tvs): Secure sessions
	return &cache.Session{
		URL:      u,
		Insecure: true,
	}, nil
}

func getSupervisorCredentials(ctx context.Context, c *config.Config, jumpbox *sshit.Client) (controlPlane, password string, err error) {
	l := zerolog.Ctx(ctx)

	// TODO(tvs): Reuse jumpbox ssh client
	// Start tunnel if we need it
	var endpoint sshit.Endpoint
	if c.JumpboxConfig != nil {
		tunnel := sshit.NewForwardTunnel(ctx,
			sshit.Endpoint{Host: "localhost", Port: 0},
			sshit.Endpoint{Host: c.VCenterConfig.SSH.Host, Port: 22})

		if err = tunnel.Bind(jumpbox); err != nil {
			l.Error().Err(err).Msg("unable to establish tunnel to vCenter")
			return "", "", fmt.Errorf("unable to establish tunnel to vCenter: %w", err)
		}

		defer func() {
			if tErr := tunnel.Close(); tErr != nil {
				l.Error().Errs("err", tErr).Msg("unable to close tunnel to vCenter")
				if err == nil {
					err = errors.Join(tErr...)
				}
			}
		}()

		endpoint = sshit.Endpoint{Host: tunnel.Local().Host, Port: tunnel.Local().Port}
	} else {
		endpoint = sshit.Endpoint{Host: c.VCenterConfig.SSH.Host, Port: *c.VCenterConfig.SSH.Port}
	}

	cfg, err := c.VCenterConfig.SSH.ClientConfig()
	if err != nil {
		return "", "", err
	}

	ssh := sshit.Client{
		Config: cfg,
		Server: endpoint,
	}

	if err := ssh.Connect(ctx); err != nil {
		l.Error().Err(err).Msg("unable to initiate SSH connection")
		return "", "", fmt.Errorf("unable to initiate SSH connection: %w", err)
	}

	defer func() {
		if tErr := ssh.Close(); tErr != nil {
			l.Error().Err(tErr).Msg("unable to close SSH session")
			if err == nil {
				err = fmt.Errorf("unable to close SSH session: %w", tErr)
			}
		}
		ssh.Close()
	}()

	stdout, stderr, err := ssh.Run("/usr/lib/vmware-wcp/decryptK8Pwd.py")
	if err != nil {
		l.Error().Err(err).Str("stderr", stderr).Msg("unable to execute decryptK8Pwd")
		return "", "", fmt.Errorf("unable to execute decryptK8Pwd: %w", err)
	}

	return parseDecryptK8Pwd(stdout)
}

func parseDecryptK8Pwd(s string) (ip string, password string, err error) {
	ip, err = getMatch(s, regexp.MustCompile(`(?m:^IP: (\S+)$)`))
	if err != nil {
		return "", "", err
	}

	password, err = getMatch(s, regexp.MustCompile(`(?m:^PWD: (\S+)$)`))
	if err != nil {
		return "", "", err
	}

	return ip, password, nil
}

func getMatch(s string, r *regexp.Regexp) (string, error) {
	matches := r.FindAllStringSubmatch(s, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("obtained %d matches; regex %q does not match output", len(matches), r.String())
	}
	return matches[0][1], nil
}
