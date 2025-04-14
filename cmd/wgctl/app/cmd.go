package app

import (
	"fmt"

	"github.com/lucheng0127/wg/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wgctl",
		Short: "Command line tool for manage vpn server",
	}
	cmd.PersistentFlags().StringVarP(&cfgFile, "conf", "c", "./client.yaml", "config file default ./client.yaml")

	cmd.AddCommand(newSubnetCommand())
	cmd.AddCommand(newPeerCommand())

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version of apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Get().Version)
		},
	}
	cmd.AddCommand(versionCmd)

	return cmd
}
