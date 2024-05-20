/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package load

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pload "github.com/tvs/ultravisor/pkg/cmd/load"
	"github.com/tvs/ultravisor/pkg/util/log"
)

func BindFlags(o *pload.LoadOptions, flags *pflag.FlagSet) {
	flags.StringVarP(&o.Container, "container", "c", "", "Container tarball to load")
	cobra.MarkFlagRequired(flags, "container")

	// vCenter SSH Setup
	flags.StringVar(&o.Server, "vcenter.server", "", "Address for the vCenter server that contains the target Supervisor Cluster")
	cobra.MarkFlagRequired(flags, "vcenter.server")
	flags.StringVar(&o.User, "vcenter.user", "", "SSH user for the vCenter server")
	cobra.MarkFlagRequired(flags, "vcenter.user")
	flags.StringVar(&o.Password, "vcenter.password", "", "SSH password for the vCenter server")
	cobra.MarkFlagRequired(flags, "vcenter.password")

	// vCenter SSO Setup
	flags.StringVar(&o.SSOUser, "vcenter.sso_user", "", "SSO User for vCenter server")
	cobra.MarkFlagRequired(flags, "vcenter.sso_user")
	flags.StringVar(&o.SSOPassword, "vcenter.sso_password", "", "SSO password for the vCenter server")
	cobra.MarkFlagRequired(flags, "vcenter.sso_password")

	// Optional jumpbox setup
	flags.StringVar(&o.Jumpbox, "jumpbox.server", "", "Optional jumpbox server")
	flags.StringVar(&o.JumpboxUser, "jumpbox.user", "", "Optional jumpbox user. Required if using a jumpbox.")
	flags.StringVar(&o.JumpboxPassword, "jumpbox.password", "", "Optional jumpbox password. Required if using a jumpbox.")

	// General options
	flags.BoolVar(&o.Cleanup, "cleanup", true, "Clean up container files after load")
}

// NewCommand represents the version command
func NewCommand() *cobra.Command {
	o := &pload.LoadOptions{}

	c := &cobra.Command{
		Use:   "load",
		Short: "load a container into a vCenter Supervisor Cluster",
		Long:  `loads a container into each of the vCenter Supervisor Cluster's control plane VMs`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(); err != nil {
				return err
			}

			l := log.LoggerFromContext(cmd.Context())

			if err := o.Run(cmd.Context(), l); err != nil {
				l.Error("Unable to load image", "error", err)
				return err
			}

			return nil
		},
	}

	BindFlags(o, c.Flags())

	return c
}
