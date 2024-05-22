package configmanager

import (
	"fmt"
	"os"

	"github.com/tvs/ultravisor/pkg/config"
	"gopkg.in/yaml.v2"
)

func LoadFrom(f string) (*config.Config, error) {
	c := &config.Config{}
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("could not load config from file: %w", err)
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, fmt.Errorf("could not load config from file: %w", err)
	}

	return c, nil
}

// Load loads the config. An error is only returned if there was an issue
// retrieving an existing config file. An empty config is returned if no
// file exists.
func Load() (*config.Config, error) {
	f, err := config.File()
	if err != nil {
		return nil, fmt.Errorf("config path cannot be retrieved: %w", err)
	}

	if _, err := os.Stat(f); err != nil {
		// Config file doesn't exist, but that's ok because we can save defaults
		// later
		return &config.Config{}, nil
	}

	return LoadFrom(f)
}
