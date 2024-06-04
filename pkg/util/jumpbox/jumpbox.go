package jumpbox

import (
	"context"
	"fmt"
	"net/url"

	"github.com/tvs/sshit"
	"github.com/tvs/ultravisor/pkg/config"
)

func Validate(c *config.SSHConfig) error {
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

// JumpboxClient returns an sshit client, a cleanup function to call once
// finished with the client, and an error if there was an issue initiating the
// connection
func JumpboxClient(ctx context.Context, c *config.SSHConfig) (*sshit.Client, func(), error) {
	if err := Validate(c); err != nil {
		return nil, nil, err
	}

	cfg, err := c.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	ssh := &sshit.Client{
		Config: cfg,
		Server: sshit.Endpoint{
			Host: c.Host,
			Port: *c.Port,
		},
	}

	if err := ssh.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("unable to initiate jumpbox connection: %w", err)
	}

	cleanup := func() {
		if tErr := ssh.Close(); tErr != nil {
			if err == nil {
				err = fmt.Errorf("unable to close jumpbox session: %w", tErr)
			}
		}
		ssh.Close()
	}

	return ssh, cleanup, nil

}
