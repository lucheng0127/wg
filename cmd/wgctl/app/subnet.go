package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
	"github.com/lucheng0127/wg/pkg/request"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	subnetResFile string
)

func subnetList(cmd *cobra.Command, args []string) error {
	req := &request.ReqMgr{
		Host: "https://127.0.0.1:5443",
	}

	rsp, err := req.SubnetList()
	if err != nil {
		return err
	}

	var subnets []modelv1.Subnet

	if err := json.Unmarshal(rsp, &subnets); err != nil {
		return err
	}

	tbl := table.New("UUID", "Name", "Address", "Public Key")

	for _, subnet := range subnets {
		tbl.AddRow(subnet.Uuid, subnet.Name, subnet.Addr, subnet.PubKey)
	}
	tbl.Print()

	return nil
}

type SubentReqRes struct {
	Name  string `yaml:"name" json:"name"`
	Iface string `yaml:"iface" json:"iface"`
	Addr  string `yaml:"addr" json:"addr"`
	Port  int    `yaml:"port" json:"port"`
}

func subnetAdd(cmd *cobra.Command, args []string) error {
	if subnetResFile == "" {
		return errors.New("subnet yaml file is required")
	}

	f, err := os.ReadFile(subnetResFile)
	if err != nil {
		return err
	}

	subRes := new(SubentReqRes)
	if err := yaml.Unmarshal(f, subRes); err != nil {
		return err
	}

	raw, err := json.Marshal(subRes)
	if err != nil {
		return err
	}

	req := &request.ReqMgr{
		Host: "https://127.0.0.1:5443",
	}
	rsp, err := req.SubnetAdd(raw)
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))
	return nil
}

func subnetDel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("uuid of subnet required")
	}

	if err := uuid.Validate(args[0]); err != nil {
		return err
	}

	req := &request.ReqMgr{
		Host: "https://127.0.0.1:5443",
	}

	rsp, err := req.SubnetDel(args[0])
	if err != nil {
		return err
	}

	fmt.Println(string(rsp))

	return nil
}

func newSubnetAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add subnet",
		RunE:  subnetAdd,
	}

	cmd.PersistentFlags().StringVarP(&subnetResFile, "file", "f", "", "resource file")
	return cmd
}

func newSubnetDelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del",
		Short: "del subnet, args: subnet uuid",
		RunE:  subnetDel,
	}

	return cmd
}

func newSubnetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "subnet",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list subnet",
		RunE:  subnetList,
	})
	cmd.AddCommand(newSubnetAddCommand())
	cmd.AddCommand(newSubnetDelCommand())

	return cmd
}
