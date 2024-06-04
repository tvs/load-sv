package config

import "context"

type ctxKey struct{}

// WithContext returns a copy of ctx with the receiver attached.
func (c *Config) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// Ctx returns the Config associated with the ctx. If no config is associated a
// new, default config is returned.
func Ctx(ctx context.Context) *Config {
	if c, ok := ctx.Value(ctxKey{}).(*Config); ok {
		return c
	}

	c := &Config{}
	return c
}
