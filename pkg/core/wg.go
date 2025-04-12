package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucheng0127/wg/pkg/utils/cmd"
)

type WG struct {
	Name       string
	CfgDir     string
	ExternalIP string

	Interface *Iface
	Peers     []*Peer
}

func (w *WG) CfgFilePath() string {
	return filepath.Join(w.CfgDir, fmt.Sprintf("%s.conf", w.Name))
}

func (w *WG) GenerateClientConf(peer *Peer) (string, error) {
	if w.ExternalIP == "" {
		return "", errors.New("wireguard server access ip address required")
	}

	port := 51820
	if w.Interface.ListenPort != 0 {
		port = w.Interface.ListenPort
	}
	keepalive := 10
	if peer.PersistentKeepalive != 0 {
		keepalive = peer.PersistentKeepalive
	}

	if w.Interface.PublicKey == "" {
		return "", errors.New("wireguard server public key required")
	}

	if peer.PrivateKey == "" {
		return "", errors.New("wireguard client private key required")
	}

	if peer.Address == "" {
		return "", errors.New("wireguard client address required")
	}

	if peer.AllowedIPs == "" {
		return "", errors.New("wireguard client allowed ips required")
	}

	var buf bytes.Buffer
	if _, err := buf.WriteString(fmt.Sprintln("[Interface]")); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("PrivateKey = %s\n", peer.PrivateKey)); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("Address = %s\n", peer.Address)); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintln("")); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintln("[Peer]")); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("PublicKey = %s\n", w.Interface.PublicKey)); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs)); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", w.ExternalIP, port)); err != nil {
		return "", err
	}
	if _, err := buf.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", keepalive)); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (w *WG) Down() error {
	// Pre down
	if w.Interface.PreDown != "" {
		subcmd := strings.Split(w.Interface.PreDown, " ")
		mgr, err := cmd.NewCmdMgr(subcmd[0])
		if err != nil {
			return err
		}

		if _, err := mgr.Execute(subcmd[1:]...); err != nil {
			return fmt.Errorf("failed to run pre down [%s]: %s", w.Interface.PreDown, err.Error())
		}
	}

	ipMgr, err := cmd.NewCmdMgr("ip")
	if err != nil {
		return err
	}
	if _, err := ipMgr.Execute(strings.Split(fmt.Sprintf("link del %s", w.Name), " ")...); err != nil {
		return fmt.Errorf("failed to delete wireguard [%s] link: %s", w.Name, err.Error())
	}

	// Post down
	if w.Interface.PostDown != "" {
		subcmd := strings.Split(w.Interface.PostDown, " ")
		mgr, err := cmd.NewCmdMgr(subcmd[0])
		if err != nil {
			return err
		}

		if _, err := mgr.Execute(subcmd[1:]...); err != nil {
			return fmt.Errorf("failed to run post down [%s]: %s", w.Interface.PostDown, err.Error())
		}
	}

	return nil
}

func (w *WG) Update() error {
	wgMgr, err := cmd.NewCmdMgr("wg")
	if err != nil {
		return err
	}

	// Generate conf
	if err := w.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate peer config file %s for %s: %s", w.CfgFilePath(), w.Name, err.Error())
	}

	// Sync conf
	if _, err := wgMgr.Execute(strings.Split(fmt.Sprintf("syncconf %s %s", w.Name, w.CfgFilePath()), " ")...); err != nil {
		return fmt.Errorf("failed to sync config for %s with %s: %s", w.Name, w.CfgFilePath(), err.Error())
	}

	return nil
}

func (w *WG) Up() error {
	// Pre up
	if w.Interface.PreUp != "" {
		subcmd := strings.Split(w.Interface.PreUp, " ")
		mgr, err := cmd.NewCmdMgr(subcmd[0])
		if err != nil {
			return err
		}

		if _, err := mgr.Execute(subcmd[1:]...); err != nil {
			return fmt.Errorf("failed to run pre up [%s]: %s", w.Interface.PreUp, err.Error())
		}
	}

	ipMgr, err := cmd.NewCmdMgr("ip")
	if err != nil {
		return err
	}
	wgMgr, err := cmd.NewCmdMgr("wg")
	if err != nil {
		return err
	}

	// Create link
	if _, err := ipMgr.Execute(strings.Split(fmt.Sprintf("link add dev %s type wireguard", w.Name), " ")...); err != nil {
		return fmt.Errorf("failed to create wireguard [%s] link: %s", w.Name, err.Error())
	}

	// Add address
	if _, err := ipMgr.Execute(strings.Split(fmt.Sprintf("address add %s dev %s", w.Interface.Address, w.Name), " ")...); err != nil {
		return fmt.Errorf("failed to assign address %s to %s: %s", w.Interface.Address, w.Name, err.Error())
	}

	// Set mtu
	if w.Interface.MTU != 0 {
		if _, err := ipMgr.Execute(strings.Split(fmt.Sprintf("link set %s mtu %d", w.Name, w.Interface.MTU), " ")...); err != nil {
			return fmt.Errorf("failed to set %s mtu to %d: %s", w.Name, w.Interface.MTU, err.Error())
		}
	}

	// Set up
	if _, err := ipMgr.Execute(strings.Split(fmt.Sprintf("link set %s up", w.Name), " ")...); err != nil {
		return fmt.Errorf("failed to set %s up: %s", w.Name, err.Error())
	}

	// Generate config
	if err := w.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate peer config file %s for %s: %s", w.CfgFilePath(), w.Name, err.Error())
	}

	// Config key and peers
	// When update peers use syncconf to sync peers config
	if _, err := wgMgr.Execute(strings.Split(fmt.Sprintf("setconf %s %s", w.Name, w.CfgFilePath()), " ")...); err != nil {
		return fmt.Errorf("failed to set config for %s with %s: %s", w.Name, w.CfgFilePath(), err.Error())
	}

	// Post up
	if w.Interface.PostUp != "" {
		subcmd := strings.Split(w.Interface.PostUp, " ")
		mgr, err := cmd.NewCmdMgr(subcmd[0])
		if err != nil {
			return err
		}

		if _, err := mgr.Execute(subcmd[1:]...); err != nil {
			return fmt.Errorf("failed to run post up [%s]: %s", w.Interface.PostUp, err.Error())
		}
	}

	return nil
}

func (w *WG) generateConfig() error {
	f, err := os.Create(w.CfgFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	fw := bufio.NewWriter(f)

	if _, err := fw.WriteString("[Interface]\n"); err != nil {
		return err
	}
	if _, err := fw.WriteString(fmt.Sprintf("PrivateKey = %s\n", w.Interface.PrivateKey)); err != nil {
		return err
	}
	listenPort := 51820
	if w.Interface.ListenPort != 0 {
		listenPort = w.Interface.ListenPort
	}
	if _, err := fw.WriteString(fmt.Sprintf("ListenPort = %d\n", listenPort)); err != nil {
		return err
	}

	for _, p := range w.Peers {
		if _, err := fw.WriteString(fmt.Sprintf("\n# User: %s\n", p.Name)); err != nil {
			return err
		}
		if _, err := fw.WriteString("[Peer]\n"); err != nil {
			return err
		}
		if _, err := fw.WriteString(fmt.Sprintf("PublicKey = %s\n", p.PublicKey)); err != nil {
			return err
		}
		if _, err := fw.WriteString(fmt.Sprintf("AllowedIPs = %s\n", p.AllowedIPs)); err != nil {
			return err
		}
	}

	return fw.Flush()
}
