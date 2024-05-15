package load

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type LoadOptions struct {
	Container  string
	Supervisor string
	Password   string
}

func (o *LoadOptions) BindFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.Container, "container", "c", "", "Name of the container to load")
	cobra.MarkFlagRequired(flags, "container")
	flags.StringVarP(&o.Supervisor, "vcenter", "s", "", "Address for the vCenter server that contains the target Supervisor Cluster")
	cobra.MarkFlagRequired(flags, "supervisor")
	flags.StringVarP(&o.Password, "password", "p", "", "SSH password for the vCenter server")
	cobra.MarkFlagRequired(flags, "password")
}

func (o *LoadOptions) Validate(c *cobra.Command, args []string) error {
	return nil
}

func (o *LoadOptions) Run(c *cobra.Command) error {
	fmt.Println("load called")
	return nil
}
