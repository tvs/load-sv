package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type requiredDir struct {
	once sync.Once
	dir  func() (string, error)
}

func (r *requiredDir) Dir() (string, error) {
	dir, err := r.dir()
	if err != nil {
		return "", fmt.Errorf("cannot fetch required directory: %w", err)
	}

	r.once.Do(func() {
		if err = os.MkdirAll(dir, 0755); err != nil {
			err = fmt.Errorf("cannot make required directory: %w", err)
			return
		}
	})

	return dir, err
}

var (
	configBaseDir = requiredDir{
		dir: func() (string, error) {
			if runtime.GOOS != "darwin" {
				dir := os.Getenv("XDG_CONFIG_HOME")
				if dir != "" {
					return filepath.Join(dir, "ultravisor"), nil
				}
			}

			dir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dir = filepath.Join(dir, ".config")
			_, err = os.Stat(dir)
			if err == nil || runtime.GOOS == "darwin" {
				// It's ok if we can't find the directory, we'll create it later
				return filepath.Join(dir, "ultravisor"), nil
			}

			dir, err = os.UserConfigDir()
			if err != nil {
				return "", err
			}

			return dir, nil
		},
	}

	configDir = requiredDir{
		dir: func() (string, error) {
			dir, err := configBaseDir.Dir()
			if err != nil {
				return "", nil
			}

			return filepath.Join(dir, profile.Name), nil
		},
	}
)

// File gets the path for the config file based on the current profile name
func File() (string, error) {
	base, err := configDir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "ultravisor.yaml"), nil
}
