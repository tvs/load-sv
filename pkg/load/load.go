package load

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/tvs/ultravisor/pkg/config"
)

func Load(ctx context.Context, c *config.Config, container string) error {
	l := zerolog.Ctx(ctx)
	l.Info().Any("bang", "crash").Msg("Hi, I'm a loader!")
	l.Warn().Any("wait", "holup").Msg("something is off here...")
	err := fmt.Errorf("oopsie doodles")
	l.Error().Err(err).Any("well", "darn").Msg("I forgot to do things!")

	// TODO(tvs): Rebuild load correctly...
	return nil
}
