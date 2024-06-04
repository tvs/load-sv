package config

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/tvs/ultravisor/pkg/util/duration"
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
	VCenterConfig *VCenterConfig `json:"vcenter,omitempty" yaml:"vcenter,omitempty"`
}

// SSHConfig represents the configuration needed to SSH to a server. Each
// command may only utilize or require a subset of this configuration and thus
// must perform validation within the context of that command.
type SSHConfig struct {
	// User is the username for the SSH connection.
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	// Host is the address of the server to SSH to.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
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

// ClientConfig creates an SSH ClientConfig for use when establishing SSH
// connections or tunnels
func (c *SSHConfig) ClientConfig() (*ssh.ClientConfig, error) {
	auth, err := c.auth()
	if err != nil {
		return nil, err
	}

	var timeout time.Duration
	if c.Timeout == nil {
		timeout = 60 * time.Second
	} else {
		timeout = c.Timeout.Duration
	}

	return &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{auth},
		// TODO(tvs): Establish trusts and use this as a fallback option...
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}, nil
}

// Address returns a string containing the host and port for an SSH connection.
func (c *SSHConfig) Address() string {
	port := 22
	if c.Port != nil {
		port = *c.Port
	}

	return fmt.Sprintf("%s:%d", c.Host, port)
}

func (c *SSHConfig) auth() (ssh.AuthMethod, error) {
	if c.Password != nil {
		return ssh.Password(*c.Password), nil
	}

	var key []byte
	if c.KeyPath != nil {
		var err error
		key, err = os.ReadFile(*c.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve key file: %w", err)
		}
	}

	if c.Key != nil {
		key = []byte(*c.Key)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// VCenterConfig represents the data needed to access the vCenter Server and/or
// its API. Each command may only utilize or require a subset of this
// configuration and thus must perform validation within the context of that
// command.
type VCenterConfig struct {
	SSH *SSHConfig `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	SSO *SSOConfig `json:"sso,omitempty" yaml:"sso,omitempty"`
}

// SSOConfig represents the data needed to access the vCenter Server's APIs.
// Each command may only utilize or require a subset of this configuration and
// thus must perform validation within the context of that command.
type SSOConfig struct {
	// User is the username used to access the vCenter APIs.
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	// Password is the password used to access the vCenter APIs.
	Password string `json:"password,omitempty" yaml:"password,omitempty"`

	// TODO(tvs): Should the host and port be separately configurable from SSH?
}
