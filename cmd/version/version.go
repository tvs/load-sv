package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/tvs/ultravisor/cmd/root"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  "Prints the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%+v\n", provenance())
	},
}

type provenanceInfo struct {
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

func provenance() provenanceInfo {
	b, _ := debug.ReadBuildInfo()

	revision := "unknown"
	for _, setting := range b.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value
			break
		}
	}
	return provenanceInfo{
		Version:   b.Main.Version,
		Revision:  revision,
		GoOs:      runtime.GOOS,
		GoArch:    runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
}

func init() {
	root.Cmd().AddCommand(versionCmd)
}
