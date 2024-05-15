/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Variables set at build time using ldflags
var (
	// During build and release this will be set to the release tag
	version = "v0.0.0-dev"
	// The latest git commit SHA
	revision = "HEAD"
	// Build time in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildTime string
)

type Provenance struct {
	// Version of the binary.
	Version string `json:"version,omitempty"`
	// Revision is a git commit.
	Revision string `json:"revision,omitempty"`
	//BuildTime is the time of the build.
	BuildTime string `json:"buildTime,omitempty"`
	// GoOs holds the name of the OS the binary was built on.
	GoOs string `json:"goOs,omitempty"`
	// GoArch holds the name of the architecture the binary was built on.
	GoArch string `json:"goArch,omitempty"`
	// GoVersion holds the version of Go used to build the binary.
	GoVersion string `json:"goVersion,omitempty"`
}

func GetProvenance() Provenance {
	return Provenance{
		Version:   version,
		Revision:  revision,
		BuildTime: buildTime,
		GoOs:      runtime.GOOS,
		GoArch:    runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
}

// NewCommand represents the version command
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Prints the version",
		Long:  "Prints the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%+v\n", GetProvenance())
		},
	}

	return c
}
