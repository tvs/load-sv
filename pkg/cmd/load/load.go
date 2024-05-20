package load

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/appleboy/easyssh-proxy"
	"github.com/elliotchance/sshtunnel"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/crypto/ssh"
)

// TODO(tvs): Clean this up holy bejeesus
// TODO(tvs): Convert prints to logger so we can squelch output and set
// verbosity
type LoadOptions struct {
	Container       string
	Jumpbox         string
	JumpboxUser     string
	JumpboxPassword string
	Server          string
	User            string
	Password        string
	SSOUser         string
	SSOPassword     string
	Cleanup         bool
}

func (o *LoadOptions) Validate() error {
	if o.Jumpbox != "" {
		if o.JumpboxUser == "" || o.JumpboxPassword == "" {
			return fmt.Errorf("If using a jumpbox, both the 'jumpbox.user' and 'jumpbox.password' must be set")
		}
	}

	return nil
}

func (o *LoadOptions) Run(ctx context.Context, l *slog.Logger) error {
	// If we have a jumpbox, we need to proxy our VMOMI requests to the vc
	var tunnel *sshtunnel.SSHTunnel

	if o.HasJumpbox() {
		tunnel = o.getTunnel()
		go tunnel.Start()
		l.Debug("Waiting for the tunnel to bind...")
		// Need to wait a moment to let the bind happen
		time.Sleep(100 * time.Millisecond)
	}

	l.Info("Fetching the control plane addresses")
	// Do the cheap stuff first - get the control plane addresses
	addrs, err := o.getSupervisorControlPlaneAddresses(tunnel)
	if err != nil {
		return fmt.Errorf("Unable to get Supervisor Control Plane VMs: %w", err)
	}

	l.Info("Retrieved Supervisor Control Plane Addresses", "addresses", addrs)

	// Close our tunnel now that we're done
	if tunnel != nil {
		l.Debug("Closing the tunnel...")
		go tunnel.Close()
	}

	// Set up jumpbox as an SSH proxy
	var proxy easyssh.DefaultConfig
	if o.HasJumpbox() {
		l.Debug("Setting up the jumpbox...")
		proxy = easyssh.DefaultConfig{
			User:     o.JumpboxUser,
			Server:   o.Jumpbox,
			Password: o.JumpboxPassword,
			Port:     "22",
		}
	}

	ssh := &easyssh.MakeConfig{
		User:     o.User,
		Server:   o.Server,
		Password: o.Password,
		Port:     "22",
		Timeout:  60 * time.Second,
		Proxy:    proxy,
	}

	// Get our Supervisor CP VM Password
	l.Info("Retrieving Supervisor Control Plane VM Password")
	stdout, _, _, err := ssh.Run("/usr/lib/vmware-wcp/decryptK8Pwd.py")
	if err != nil {
		return fmt.Errorf("Unable to retrieve Supervisor Cluster password: %w", err)
	}
	l.Debug(stdout)
	supervisorPassword, err := getSupervisorPassword(stdout)
	if err != nil {
		return fmt.Errorf("Unable to retrieve Supervisor Cluster password: %w", err)
	}
	l.Debug("Retrieved Supervisor Control Plane VM Password", "password", supervisorPassword)

	// On to the expensive stuff

	// Start copying the container file over to the vCenter server
	l.Info("Copying container to VC", "container", o.Container, "vc", o.Server)
	containerPath := filepath.Join("/tmp", filepath.Base(o.Container))
	err = ssh.Scp(o.Container, containerPath)
	if err != nil {
		return fmt.Errorf("Unable to run remote command: %w", err)
	}

	// Optional cleanup of files on the first remote server
	defer func() {
		if o.Cleanup {
			_, stderr, done, err := ssh.Run(fmt.Sprintf("rm %s", containerPath))
			if !done || err != nil {
				fmt.Fprintf(os.Stderr, "Unable to clean up container files: remote.stderr: %s\nerr: %s", stderr, err)
			}
		}
	}()

	// Now SCP everything to the CP VMs
	scpCmdArgs := "-q -o PubkeyAuthentication=no -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
	for _, addr := range addrs {
		l.Info("Copying container to Supervisor Control Plane VM", "container", o.Container, "vm", addr)
		scpCmd := fmt.Sprintf(`sshpass -p %q scp %s "%s" "%s@%s:%s"`, supervisorPassword, scpCmdArgs, containerPath, "root", addr, containerPath)
		_, _, _, err := ssh.Run(scpCmd)
		if err != nil {
			return fmt.Errorf("Unable to write container to Supervisor VM %s: %w", addr, err)
		}

		// Now load into the local registry
		l.Info("Loading container into Supervisor Control Plane's containerd registry", "container", o.Container, "vm", addr)
		ctrCmd := fmt.Sprintf(`sshpass -p %q ssh %s "%s@%s" "ctr -n k8s.io images import %s"`, supervisorPassword, scpCmdArgs, "root", addr, containerPath)
		_, _, _, err = ssh.Run(ctrCmd)
		if err != nil {
			return fmt.Errorf("Unable to write container to Supervisor VM %s: %w", addr, err)
		}
	}

	return nil
}

// HasJumpbox returns true if the Jumpbox is configured
func (o *LoadOptions) HasJumpbox() bool {
	return o.Jumpbox != ""
}

// TODO(tvs): Replace SSHTunnel lib
func (o *LoadOptions) getTunnel() *sshtunnel.SSHTunnel {
	return &sshtunnel.SSHTunnel{
		Config: &ssh.ClientConfig{
			User: o.JumpboxUser,
			Auth: []ssh.AuthMethod{ssh.Password(o.JumpboxPassword)},
			HostKeyCallback: func(_ string, _ net.Addr, _ ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local: &sshtunnel.Endpoint{
			Host: "localhost",
			Port: 8443,
		},
		Server: &sshtunnel.Endpoint{
			Host: o.Jumpbox,
			Port: 22,
			User: o.JumpboxUser,
		},
		Remote: &sshtunnel.Endpoint{
			Host: o.Server,
			Port: 443,
		},
	}
}

func (o *LoadOptions) getVCClient(ctx context.Context, tunnel *sshtunnel.SSHTunnel) (*vim25.Client, error) {
	ip := o.Server
	port := 443

	if tunnel != nil {
		ip = tunnel.Local.Host
		port = tunnel.Local.Port
	}

	u, err := url.Parse(fmt.Sprintf("https://%s:%d/sdk", ip, port))
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(o.SSOUser, o.SSOPassword)

	// TODO(tvs): Log out of session
	s := &cache.Session{
		URL:      u,
		Insecure: true,
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to log in with VC client: %w", err)
	}

	return c, nil
}

func (o *LoadOptions) getSupervisorControlPlaneAddresses(tunnel *sshtunnel.SSHTunnel) ([]string, error) {
	ctx := context.Background()

	c, err := o.getVCClient(ctx, tunnel)
	if err != nil {
		return nil, err
	}

	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)

	objs, err := v.Find(ctx, []string{"VirtualMachine"}, property.Match{"name": "SupervisorControlPlaneVM*"})
	if err != nil {
		return nil, err
	}
	if len(objs) < 1 {
		return nil, fmt.Errorf("No Supervisor Control Plane VMs were found")
	}

	p := property.DefaultCollector(c)

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

	return ipAddrs, nil
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

func getSupervisorPassword(s string) (string, error) {
	r := regexp.MustCompile(`(?m:^PWD: (\S+)$)`)

	matches := r.FindAllStringSubmatch(s, -1)
	if len(matches) != 1 {
		return "", fmt.Errorf("Obtained %d password matches; regex needs to be changed", len(matches))
	}

	return matches[0][1], nil
}
