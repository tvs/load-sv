package load

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/tvs/ultravisor/pkg/config"
	"github.com/tvs/ultravisor/pkg/supervisor"
)

func Load(ctx context.Context, container string) error {
	l := zerolog.Ctx(ctx)
	c := config.Ctx(ctx)

	l.Debug().Interface("config", c).Msg("beginning load")

	if err := supervisor.ValidateConfig(c); err != nil {
		l.Error().Err(err).Any("config", c).Msg("invalid config")
		return err
	}

	// TODO(tvs): Rebuild load correctly...
	return nil
}
