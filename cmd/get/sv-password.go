package get

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/tvs/ultravisor/cmd/root"
	"github.com/tvs/ultravisor/pkg/supervisor"
)

var getSupervisorCmd = &cobra.Command{
	Use:     "supervisor",
	Aliases: []string{"sv"},
	Long:    `fetch the supervisor cluster information for the vSphere IaaS Control Plane`,
	Example: "  load supervisor\n" +
		"  get supervisor --password\n" +
		"  get supervisor --control-plane\n" +
		"  get supervisor --vms\n" +
		"  get supervisor --all",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		l := zerolog.Ctx(cmd.Context())

		info, err := supervisor.Info(cmd.Context())
		if err != nil {
			l.Error().Err(err).Msg("unable to retrieve supervisor info")
			root.SetExitCode(1)
			return
		}

		if getSupervisorCmdArgs.Flags.ControlPlane {
			info.VMs = []string{}
			info.Password = ""
		}

		if getSupervisorCmdArgs.Flags.VMs {
			info.ControlPlane = ""
			info.Password = ""
		}

		if getSupervisorCmdArgs.Flags.Password {
			info.ControlPlane = ""
			info.VMs = []string{}
		}

		b, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			l.Error().Err(err).Msg("unable to marshal supervisor info into json")
		}

		fmt.Println(string(b))
	},
}

var getSupervisorCmdArgs struct {
	Flags struct {
		ControlPlane bool
		Password     bool
		VMs          bool
		All          bool
	}
}

func init() {
	getSupervisorCmd.Flags().BoolVar(&getSupervisorCmdArgs.Flags.Password, "password", false, "fetch only the supervisor cluster password")
	getSupervisorCmd.Flags().BoolVar(&getSupervisorCmdArgs.Flags.ControlPlane, "control-plane", false, "fetch only the control plane address")
	getSupervisorCmd.Flags().BoolVar(&getSupervisorCmdArgs.Flags.VMs, "vms", false, "fetch only the supervisor control plane VMs")
	getSupervisorCmd.Flags().BoolVar(&getSupervisorCmdArgs.Flags.All, "all", true, "fetch all information about the supervisor (default)")
	getSupervisorCmd.MarkFlagsMutuallyExclusive("password", "control-plane", "vms", "all")

	getCmd.AddCommand(getSupervisorCmd)
}
