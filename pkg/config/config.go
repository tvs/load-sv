package config

import (
	"time"

	"github.com/tvs/ultravisor/pkg/util/duration"
	"github.com/tvs/ultravisor/pkg/util/ptr"
)

var profile = Profile{Name: "default"}

// Profile repres
type Profile struct {
	Name string
}

func SetProfile(name string) {
	profile = Profile{Name: name}
}

func CurrentProfile() Profile { return profile }

// NOTE: Even though YAML formats are used for the config files, JSON tags are
// necessary for structured logging.

// Config represents an application config. This is, functionally, the set of
// all suppliable configuration for all executable commands. Each command may
// only utilize or require a subset of this configuration and thus must perform
// validation within the context of that command.
type Config struct {
	// JumpboxConfig represents an optional set of configuration for accessing
	// a jumpbox on the management network.
	JumpboxConfig *SSHConfig `json:"jumpbox,omitempty" yaml:"jumpbox,omitempty"`
	// VCenterConfig represents a required set of configuration for accessing
	// the vCenter server.
	VCenterConfig VCenterConfig `json:"vcenter,omitempty" yaml:"vcenter,omitempty"`
}

// SetDefaults sets default values for configuration that might have been
// omitted.
func (c *Config) SetDefaults() {
	if c.JumpboxConfig != nil {
		c.JumpboxConfig.SetDefaults()
	}

	c.VCenterConfig.SetDefaults()
}

// SSHConfig represents the configuration needed to SSH to a server. Each
// command may only utilize or require a subset of this configuration and thus
// must perform validation within the context of that command.
type SSHConfig struct {
	// User is the username for the SSH connection.
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	// Server is the address of the server to SSH to.
	Server string `json:"server,omitempty" yaml:"server,omitempty"`
	// Port is the port to utilize when SSHing. Defaults to 22.
	Port *int `json:"port,omitempty" yaml:"port,omitempty"`
	// Key takes an optional encrypted private key.
	Key *string `json:"key,omitempty" yaml:"key,omitempty"`
	// KeyPath takes an optional path to a private key on the local machine.
	KeyPath *string `json:"keyPath,omitempty" yaml:"keyPath,omitempty"`
	// Password takes a plaintext string of the user's password.
	Password *string `json:"password,omitempty" yaml:"password,omitempty"`
	// Timeout is timeout duration for SSH commands. Defaults to 60 seconds. 0
	// indicates no timeout.
	Timeout *duration.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// SetDefaults sets default values for configuration that might have been
// omitted.
func (c *SSHConfig) SetDefaults() {
	if c.Port == nil {
		c.Port = ptr.To(22)
	}

	if c.Timeout == nil {
		c.Timeout = &duration.Duration{Duration: 60 * time.Second}
	}
}

// VCenterConfig represents the data needed to access the vCenter Server and/or
// its API. Each command may only utilize or require a subset of this
// configuration and thus must perform validation within the context of that
// command.
type VCenterConfig struct {
	SSHConfig
	// SSOUser is the username used to access the vCenter APIs.
	SSOUser string `json:"ssoUser,omitempty" yaml:"ssoUser,omitempty"`
	// SSOPassword is the password used to access the vCenter APIs.
	SSOPassword string `json:"ssoPassword,omitempty" yaml:"ssoPassword,omitempty"`
}

// SetDefaults sets default values for configuration that might have been
// omitted.
func (c *VCenterConfig) SetDefaults() {
	c.SSHConfig.SetDefaults()
}
