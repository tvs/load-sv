/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package load

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pload "github.com/tvs/load-sv/pkg/cmd/load"
)

// NewCommand represents the version command
func NewCommand() *cobra.Command {
	o := &pload.LoadOptions{}

	c := &cobra.Command{
		Use:   "load",
		Short: "load a container into a vCenter Supervisor Cluster",
		Long:  `loads a container into each of the vCenter Supervisor Cluster's control plane VMs`,

		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Validate(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
			}

			if err := o.Run(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
			}
		},
	}

	o.BindFlags(c.Flags())

	return c
}
