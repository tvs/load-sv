/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

type Provenance struct {
	// Version of the binary.
	Version string `json:"version,omitempty"`
	// Revision is a git commit.
	Revision string `json:"revision,omitempty"`
	// GoOs holds the name of the OS the binary was built on.
	GoOs string `json:"goOs,omitempty"`
	// GoArch holds the name of the architecture the binary was built on.
	GoArch string `json:"goArch,omitempty"`
	// GoVersion holds the version of Go used to build the binary.
	GoVersion string `json:"goVersion,omitempty"`
}

func GetProvenance() Provenance {
	b, _ := debug.ReadBuildInfo()

	revision := "unknown"
	for _, setting := range b.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value
			break
		}
	}
	return Provenance{
		Version:   b.Main.Version,
		Revision:  revision,
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
