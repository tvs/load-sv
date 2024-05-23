package load

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rs/zerolog"
	"github.com/tvs/ultravisor/pkg/config"
)

func Load(ctx context.Context, container string) error {
	l := zerolog.Ctx(ctx)
	c := config.Ctx(ctx)

	l.Debug().Interface("config", c).Msg("Beginning load")

	if err := validateConfig(c); err != nil {
		l.Error().Err(err).Any("config", c).Msg("Invalid config")
		return err
	}

	// TODO(tvs): Rebuild load correctly...
	return nil
}

// TODO(tvs): Figure out a nice way to send back and log each invalidity
// errors.Join() doesn't work well with default zerolog for readability
// and log.Errs() doesn't use the colorized error handler
func validateConfig(c *config.Config) error {
	if c.JumpboxConfig != nil {
		if err := validateSSHConfig(c.JumpboxConfig); err != nil {
			return err
		}
	}

	if err := validateVCenterConfig(&c.VCenterConfig); err != nil {
		return err
	}

	return nil
}

func validateSSHConfig(c *config.SSHConfig) error {
	if c.Server == "" {
		return fmt.Errorf("server must be supplied")
	}

	if _, err := url.Parse(c.Server); err != nil {
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

	if c.Timeout == nil {
		return fmt.Errorf("a timeout must be supplied")
	}

	return nil
}

// TODO(tvs): Validate the vCenter config: sso user, etc.
func validateVCenterConfig(c *config.VCenterConfig) error {
	return validateSSHConfig(&c.SSHConfig)
}
