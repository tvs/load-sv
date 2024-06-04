package load

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"

	"github.com/tvs/sshit"
	"github.com/tvs/ultravisor/pkg/config"
	"github.com/tvs/ultravisor/pkg/supervisor"
	"github.com/tvs/ultravisor/pkg/util/jumpbox"
)

func Load(ctx context.Context, container string) error {
	l := zerolog.Ctx(ctx)
	c := config.Ctx(ctx)

	l.Debug().Interface("config", c).Msg("beginning load")

	if err := supervisor.ValidateConfig(c); err != nil {
		l.Error().Err(err).Any("config", c).Msg("invalid config")
		return err
	}

	var j *sshit.Client
	if c.JumpboxConfig != nil {
		var (
			cleanup func()
			err     error
		)

		j, cleanup, err = jumpbox.JumpboxClient(ctx, c.JumpboxConfig)
		if err != nil {
			return err
		}

		defer cleanup()
	}

	// TODO(tvs): Ensure container file _is_ a container file

	// TODO(tvs): Preserve jumpbox session across commands so we can just
	// change our tunnel for each VM
	supervisorInfo, err := supervisor.InfoWithJumpbox(ctx, j)
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve Supervisor info")
		return fmt.Errorf("unable to retrieve Supervisor info: %w", err)
	}

	target := filepath.Join("/tmp", filepath.Base(container))
	for _, vm := range supervisorInfo.VMs {
		l.Debug().Str("address", vm).Str("file", container).Str("target", target).Msg("copying file to host")
		if err := copyToVM(ctx, c, vm, supervisorInfo.Password, j, container, target); err != nil {
			l.Error().Err(err).Str("address", vm).Str("file", container).Msg("error copying file to vm")
			return err
		}

		l.Debug().Str("address", vm).Str("file", container).Msg("load to container runtime")
		if err := loadToCtr(ctx, c, vm, supervisorInfo.Password, j, target); err != nil {
			l.Error().Err(err).Str("address", vm).Str("file", target).Msg("error loading file into ctr")
			return err
		}
	}

	return nil
}

func copyToVM(ctx context.Context, c *config.Config, host, password string, jumpbox *sshit.Client, source, target string) (err error) {
	l := zerolog.Ctx(ctx)

	// Start the tunnel if we need it
	var endpoint sshit.Endpoint
	if c.JumpboxConfig != nil {
		// TODO(tvs): Extract this to utility function
		tunnel := sshit.NewForwardTunnel(ctx,
			sshit.Endpoint{Host: "localhost", Port: 0},
			sshit.Endpoint{Host: host, Port: 22})

		if err = tunnel.Bind(jumpbox); err != nil {
			l.Error().Err(err).Str("host", host).Msg("unable to establish tunnel to Supervisor VM")
			return fmt.Errorf("unable to establish tunnel to Supervisor VM: %w", err)
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
		// TODO(tvs): Configurable ports for Supervisor VMs?
		endpoint = sshit.Endpoint{Host: host, Port: 22}
	}

	// TODO(tvs): Extract to utility functions
	var timeout time.Duration
	if c.VCenterConfig.SSH.Timeout == nil {
		timeout = 60 * time.Second
	} else {
		timeout = c.VCenterConfig.SSH.Timeout.Duration
	}

	cfg := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO(tvs): Separate timeout for Supervisor
		Timeout: timeout,
	}

	l.Debug().Any("endpoint", endpoint).Msg("copying file through scp")
	ssh := sshit.Client{
		Config: cfg,
		Server: endpoint,
	}

	if err := ssh.Connect(ctx); err != nil {
		l.Error().Err(err).Msg("unable to initiate SSH connection")
		return fmt.Errorf("unable to initiate SSH connection: %w", err)
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

	return ssh.Copy(source, target)
}

func loadToCtr(ctx context.Context, c *config.Config, host, password string, jumpbox *sshit.Client, file string) (err error) {
	l := zerolog.Ctx(ctx)

	// Start the tunnel if we need it
	var endpoint sshit.Endpoint
	if c.JumpboxConfig != nil {
		// TODO(tvs): Extract this to utility function
		tunnel := sshit.NewForwardTunnel(ctx,
			sshit.Endpoint{Host: "localhost", Port: 0},
			sshit.Endpoint{Host: host, Port: 22})

		if err = tunnel.Bind(jumpbox); err != nil {
			l.Error().Err(err).Str("host", host).Msg("unable to establish tunnel to Supervisor VM")
			return fmt.Errorf("unable to establish tunnel to Supervisor VM: %w", err)
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
		// TODO(tvs): Configurable ports for Supervisor VMs?
		endpoint = sshit.Endpoint{Host: host, Port: 22}
	}

	// TODO(tvs): Extract to utility functions
	var timeout time.Duration
	if c.VCenterConfig.SSH.Timeout == nil {
		timeout = 60 * time.Second
	} else {
		timeout = c.VCenterConfig.SSH.Timeout.Duration
	}

	cfg := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// TODO(tvs): Separate timeout for Supervisor
		Timeout: timeout,
	}

	l.Debug().Any("endpoint", endpoint).Msg("copying file through scp")
	ssh := sshit.Client{
		Config: cfg,
		Server: endpoint,
	}

	if err := ssh.Connect(ctx); err != nil {
		l.Error().Err(err).Msg("unable to initiate SSH connection")
		return fmt.Errorf("unable to initiate SSH connection: %w", err)
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

	_, stderr, err := ssh.Run(fmt.Sprintf("ctr -n k8s.io images import %s", file))
	if err != nil {
		l.Error().Err(err).Str("stderr", stderr).Msg("unable to load container")
		return err
	}

	return nil
}
