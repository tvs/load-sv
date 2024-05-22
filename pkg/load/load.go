package load

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/tvs/ultravisor/pkg/config"
)

func Load(ctx context.Context, c *config.Config, container string) error {
	l := zerolog.Ctx(ctx)
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

	// TODO(tvs): URL parsing test
	if _, err := strconv.Atoi(c.Port); err != nil {
		errs = append(errs, fmt.Errorf("port must be a valid number: %w", err))
	}

	// TODO(tvs): Ensure one of Key, KeyPath, or Password are present

	return errors.Join(errs...)
}

// TODO(tvs): Validate the vCenter config: sso user, etc.
func validateVCenterConfig(c *config.VCenterConfig) error {
	var errs []error
	errs = append(errs, validateSSHConfig(&c.SSHConfig))

	return errors.Join(errs...)
}
