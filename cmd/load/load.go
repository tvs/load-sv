package load

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/tvs/ultravisor/cmd/root"
	pload "github.com/tvs/ultravisor/pkg/load"
)

var loadCmd = &cobra.Command{
	Use:   "load [container]",
	Short: "load a container into the vSphere IaaS Control Plane",
	Long:  `loads a container into each of the vSphere IaaS Control Plane's control plane VMs`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		l := zerolog.Ctx(cmd.Context())
		container := args[0]

		if err := pload.Load(cmd.Context(), container); err != nil {
			l.Error().Err(err).Msg("Unable to load container to vSphere IaaS Control Plane VMs")
			root.SetExitCode(1)
		}
	},
}

func init() {
	root.Cmd().AddCommand(loadCmd)
}
