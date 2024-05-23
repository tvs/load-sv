package configmanager

import (
	"fmt"
	"os"

	"github.com/tvs/ultravisor/pkg/config"
	"gopkg.in/yaml.v3"
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

// Save saves the config to the file defined by the profile
// TODO(tvs): Use an embedded template with comments explaining
// the config and ensure that when we save we persist the comments
func Save(c *config.Config) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal config to yaml: %w", err)
	}

	f, err := config.File()
	if err != nil {
		return fmt.Errorf("unable to retrieve file path for profile: %w", err)
	}

	return os.WriteFile(f, b, 0644)
}
