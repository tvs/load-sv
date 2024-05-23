package load

import (
	"context"
	"errors"
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

func validateConfig(c *config.Config) error {
	var errs []error
	if c.JumpboxConfig != nil {
		if err := validateSSHConfig(c.JumpboxConfig); err != nil {
			errs = append(errs, fmt.Errorf("unable to validate jumpbox config: %w", err))
		}
	}

	if err := validateVCenterConfig(&c.VCenterConfig); err != nil {
		errs = append(errs, fmt.Errorf("unable to validate vcenter config: %w", err))
	}

	return errors.Join(errs...)
}

// TODO(tvs): Validate SSH config: server format, port value, etc. etc.
func validateSSHConfig(c *config.SSHConfig) error {
	var errs []error
	if c.Server == "" {
		errs = append(errs, fmt.Errorf("server must be supplied"))
	}

	if _, err := url.Parse(c.Server); err != nil {
		errs = append(errs, fmt.Errorf("unable to parse server: %w", err))
	}

	if c.Port == nil {
		errs = append(errs, fmt.Errorf("port must be supplied"))
	}

	if c.User == "" {
		errs = append(errs, fmt.Errorf("SSH user must be supplied"))
	}

	if c.Key == nil && c.KeyPath == nil && c.Password == nil {
		errs = append(errs, fmt.Errorf("one of key, keyPath, or password must be supplied"))
	}

	if c.Timeout == nil {
		errs = append(errs, fmt.Errorf("a timeout must be supplied"))
	}

	return errors.Join(errs...)
}

// TODO(tvs): Validate the vCenter config: sso user, etc.
func validateVCenterConfig(c *config.VCenterConfig) error {
	var errs []error
	errs = append(errs, validateSSHConfig(&c.SSHConfig))

	return errors.Join(errs...)
}
