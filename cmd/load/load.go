package load

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/tvs/ultravisor/cmd/root"
	"github.com/tvs/ultravisor/pkg/config/configmanager"
	pload "github.com/tvs/ultravisor/pkg/load"
)

var loadCmd = &cobra.Command{
	Use:   "load [container]",
	Short: "load a container into the vSphere IaaS Control Plane",
	Long:  `loads a container into each of the vSphere IaaS Control Plane's control plane VMs`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		l := zerolog.Ctx(ctx)
		container := args[0]

		cfg, err := configmanager.Load()
		if err != nil {
			l.Error().Err(err).Msg("Unable to load config")
			root.SetExitCode(1)
			return
		}
		cfg.SetDefaults()

		if err := pload.Load(cmd.Context(), cfg, container); err != nil {
			l.Error().Err(err).Msg("Unable to load container to vSphere IaaS Control Plane VMs")
			root.SetExitCode(1)
		}
	},
}

func init() {
	root.Cmd().AddCommand(loadCmd)
}
