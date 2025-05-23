package app

import (
	"fmt"
	"os"
	"path"

	"github.com/lucheng0127/wg/pkg/config"
	"github.com/lucheng0127/wg/pkg/request"
	"github.com/lucheng0127/wg/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

func getRequest(c string) (*request.ReqMgr, error) {
	req := &request.ReqMgr{
		Host: "https://127.0.0.1:5443",
	}

	cfg, err := config.LoadWgCliConf(c)
	if err != nil {
		return nil, err
	}
	if cfg.CaCrt == "" || cfg.ClientCrt == "" || cfg.ClientKey == "" {
		req.Cofig = nil
	} else {
		req.Cofig = cfg
	}

	return req, nil
}

func NewCliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wgctl",
		Short: "Command line tool for manage vpn server",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultCfg := path.Join(homeDir, ".wg/client.yaml")
	cmd.PersistentFlags().StringVarP(&cfgFile, "conf", "c", defaultCfg, fmt.Sprintf("config file default: %s", defaultCfg))

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
