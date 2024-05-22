package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Config is struct representing an application config.
// This is functionally a profile that can be used when executing commands.
type Config struct {
	// JumpboxConfig represents an optional set of configuration for accessing
	// a jumpbox on the management network.
	JumpboxConfig *SSHConfig `yaml:"jumpbox,omitempty"`
	// VCenterConfig represents a required set of configuration for accessing
	// the vCenter server.
	VCenterConfig VCenterConfig `yaml:"vcenter,omitempty"`
}

// SSHConfig represents the data needed to establish an SSH connection with a
// remote server.
// Note: one of Key, KeyPath, or Password must be supplied.
type SSHConfig struct {
	// User is the username for the SSH connection.
	User string `yaml:"user,omitempty"`
	// Server is the address of the server to SSH to.
	Server string `yaml:"server,omitempty"`
	// Port is the port to utilize when SSHing. Defaults to 22.
	Port string `yaml:"port,omitempty"`
	// Key takes an optional encrypted private key.
	Key *string `yaml:"key,omitempty"`
	// KeyPath takes an optional path to a private key on the local machine.
	KeyPath *string `yaml:"keyPath,omitempty"`
	// Password takes a plaintext string of the user's password.
	Password *string `yaml:"password,omitempty"`
	// Timeout is timeout duration for SSH commands.
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

// VCenterConfig represents the data needed to access the vCenter server.
type VCenterConfig struct {
	SSHConfig
	// SSOUser is the username used to access the vCenter APIs.
	SSOUser string `yaml:"ssoUser,omitempty"`
	// SSOPassword is the password used to access the vCenter APIs.
	SSOPassword string `yaml:"ssoPassword,omitempty"`
}

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
)

func File() (string, error) {
	base, err := configBaseDir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "ultravisor.yaml"), nil
}
