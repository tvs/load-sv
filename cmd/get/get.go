package get

import (
	"github.com/spf13/cobra"

	"github.com/tvs/ultravisor/cmd/root"
)

var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "fetch a resource from the vSphere IaaS Control Plane",
	Long:  `fetch a named resource from the vSphere IaaS Control Plane`,
}

func init() {
	root.Cmd().AddCommand(getCmd)
}
