package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lucheng0127/wg/pkg/apiserver"
	"github.com/lucheng0127/wg/pkg/config"
	"github.com/lucheng0127/wg/pkg/utils/cmd"
	"github.com/lucheng0127/wg/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/sample-controller/pkg/signals"
)

var (
	cfgDir string
)

func NewApiserverCmd() *cobra.Command {
	var fs flag.FlagSet
	klog.InitFlags(&fs)

	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "VPN server base on wireguard",
		Long:  "VPN server base on wireguard",
		RunE:  launch,
	}

	klog.InitFlags(flag.CommandLine)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.PersistentFlags().StringVarP(&cfgDir, "config-dir", "d", "/opt/wg", "config directory (config file named apiserver.yaml)")

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

func precheck() error {
	if os.Geteuid() != 0 {
		return errors.New("Please run as root")
	}

	_, err := cmd.NewCmdMgr("wg")
	if err != nil {
		return err
	}

	mgr, err := cmd.NewCmdMgr("cat")
	if err != nil {
		return err
	}

	out, err := mgr.Execute("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return err
	}

	if !strings.Contains(out, "1") {
		return fmt.Errorf("ip_forward enable is required")

	}
	return nil
}

func launch(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadApiserverConf(cfgDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %s", err.Error())
	}

	// Check wg and ip_forward
	if err := precheck(); err != nil {
		return fmt.Errorf("pre-check failed: %s", err.Error())
	}

	// Setup signal handler
	ctx := signals.SetupSignalHandler()

	// New apiserver
	svc, err := apiserver.NewApiserver(cfg)
	if err != nil {
		return fmt.Errorf("failed to create apiserver: %s", err.Error())
	}

	// PreRun apiserver
	if err := svc.PreRun(ctx); err != nil {
		return fmt.Errorf("failed to pre-run apiserver: %s", err.Error())
	}

	// Run apiserver
	if err := svc.Run(ctx); err != nil {
		return fmt.Errorf("failed to run apiserver: %s", err.Error())
	}
	return nil
}
