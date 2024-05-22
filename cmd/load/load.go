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
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zerolog.Ctx(cmd.Context())
		container := args[0]

		cfg, err := configmanager.Load()
		if err != nil {
			return err
		}

		l.Debug().Any("cfg", cfg).Msg("Config loaded")

		return pload.Load(cmd.Context(), cfg, container)
	},
}

func init() {
	root.Cmd().AddCommand(loadCmd)
}
