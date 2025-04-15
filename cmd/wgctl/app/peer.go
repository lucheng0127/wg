package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
	"github.com/rodaine/table"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	peerResFile string
	peerSubnet  string
)

func peerList(cmd *cobra.Command, args []string) error {
	if peerSubnet == "" {
		return errors.New("subnet needed")
	}

	req, err := getRequest(cfgFile)
	if err != nil {
		return err
	}

	rsp, err := req.SubnetPeerList(peerSubnet)
	if err != nil {
		return err
	}

	var peers []modelv1.Peer

	if err := json.Unmarshal(rsp, &peers); err != nil {
		return err
	}

	tbl := table.New("UUID", "User", "Address", "Public Key", "Enable")

	for _, peer := range peers {
		tbl.AddRow(peer.Uuid, peer.Name, peer.Addr, peer.PubKey, peer.Enable)
	}
	tbl.Print()

	return nil
}

type PeerReqRes struct {
	Name   string `yaml:"name" json:"name"`
	Addr   string `yaml:"addr" json:"addr"`
	Subnet string `yaml:"subnet" json:"subnet"`
}

func peerAdd(cmd *cobra.Command, args []string) error {
	if peerResFile == "" {
		return errors.New("peer yaml file is required")
	}

	f, err := os.ReadFile(peerResFile)
	if err != nil {
		return err
	}

	peerRes := new(PeerReqRes)
	if err := yaml.Unmarshal(f, peerRes); err != nil {
		return err
	}

	raw, err := json.Marshal(peerRes)
	if err != nil {
		return err
	}

	req, err := getRequest(cfgFile)
	if err != nil {
		return err
	}

	rsp, err := req.PeerAdd(raw, peerRes.Subnet)
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))
	return nil
}

func peerDel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("uuid of peer required")
	}

	if err := uuid.Validate(args[0]); err != nil {
		return err
	}

	req, err := getRequest(cfgFile)
	if err != nil {
		return err
	}

	rsp, err := req.PeerDel(peerSubnet, args[0])
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))

	return nil
}

type PeerConfRsq struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

func peerConf(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("uuid of peer required")
	}

	if err := uuid.Validate(args[0]); err != nil {
		return err
	}

	req, err := getRequest(cfgFile)
	if err != nil {
		return err
	}

	rsp, err := req.PeerConf(peerSubnet, args[0])
	if err != nil {
		return err
	}

	pcRsq := new(PeerConfRsq)
	if err := json.Unmarshal(rsp, pcRsq); err != nil {
		return err
	}

	qr, err := qrcode.New(pcRsq.Config, qrcode.Low)
	if err != nil {
		return err
	}

	fmt.Println(qr.ToSmallString(false))

	return nil
}

type PeerSetReq struct {
	Enable bool `json:"enable"`
}

func peerSet(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("set <peer uuid> <enable/disable>")
	}

	if args[1] != "disable" && args[1] != "enable" {
		return errors.New("set <peer uuid> <enable/disable>")
	}
	enable := false
	if args[1] == "enable" {
		enable = true
	}

	raw, err := json.Marshal(&PeerSetReq{Enable: enable})
	if err != nil {
		return err
	}

	if err := uuid.Validate(args[0]); err != nil {
		return err
	}

	req, err := getRequest(cfgFile)
	if err != nil {
		return err
	}

	rsp, err := req.PeerSet(peerSubnet, args[0], raw)
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))

	return nil
}

func newPeerAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add peer",
		RunE:  peerAdd,
	}

	cmd.PersistentFlags().StringVarP(&peerResFile, "file", "f", "", "resource file")
	return cmd
}

func newPeerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list peer",
		RunE:  peerList,
	}

	cmd.PersistentFlags().StringVarP(&peerSubnet, "subnet", "s", "", "subent uuid")
	return cmd
}

func newPeerDelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del",
		Short: "delete subnet peer",
		RunE:  peerDel,
	}

	cmd.PersistentFlags().StringVarP(&peerSubnet, "subnet", "s", "", "subent uuid")
	return cmd
}

func newPeerConfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conf",
		Short: "config of subnet peer",
		RunE:  peerConf,
	}

	cmd.PersistentFlags().StringVarP(&peerSubnet, "subnet", "s", "", "subent uuid")
	return cmd
}

func newPeerSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set peer enable or disable",
		RunE:  peerSet,
	}

	cmd.PersistentFlags().StringVarP(&peerSubnet, "subnet", "s", "", "subent uuid")
	return cmd
}

func newPeerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "peer",
	}

	cmd.AddCommand(newPeerListCommand())
	cmd.AddCommand(newPeerAddCommand())
	cmd.AddCommand(newPeerDelCommand())
	cmd.AddCommand(newPeerConfCommand())
	cmd.AddCommand(newPeerSetCommand())

	return cmd
}
