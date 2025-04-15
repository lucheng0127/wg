package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/go-restful"
	_ "github.com/mattn/go-sqlite3"
	"k8s.io/klog/v2"
	"xorm.io/xorm"

	"github.com/lucheng0127/wg/pkg/apis/ping"
	"github.com/lucheng0127/wg/pkg/apis/subnet"
	"github.com/lucheng0127/wg/pkg/config"
	"github.com/lucheng0127/wg/pkg/core"
	"github.com/lucheng0127/wg/pkg/filters"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
)

type ApiServer struct {
	Config             *config.ApiserverConf
	DB                 *xorm.Engine
	Server             *http.Server
	handler            http.Handler
	ChangeSubnetQueue  chan string // Channel to get cahnged subent and sync to system
	AddSubnetQueue     chan string
	DeletedSubnetQueue chan string
}

func NewApiserver(cfg *config.ApiserverConf) (*ApiServer, error) {
	svc := new(ApiServer)
	svc.Config = cfg

	// Init db engine
	dbEng, err := xorm.NewEngine("sqlite3", filepath.Join(cfg.BaseDir, cfg.DB))
	if err != nil {
		return nil, err
	}
	svc.DB = dbEng

	// Parse x509 crt
	cer, err := tls.LoadX509KeyPair(filepath.Join(cfg.BaseDir, cfg.Crt), filepath.Join(cfg.BaseDir, cfg.Key))
	if err != nil {
		return nil, err
	}

	clientCaPool := x509.NewCertPool()
	caCrt, err := os.ReadFile(filepath.Join(cfg.BaseDir, cfg.Crt))
	if err != nil {
		return nil, err
	}

	ok := clientCaPool.AppendCertsFromPEM(caCrt)
	if !ok {
		return nil, errors.New("failed to add crt to client ca")
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCaPool,
	}
	httpSvc := new(http.Server)
	httpSvc.TLSConfig = tlsConf
	httpSvc.Addr = fmt.Sprintf(":%d", svc.Config.Port)
	svc.Server = httpSvc

	svc.ChangeSubnetQueue = make(chan string, 254)
	svc.AddSubnetQueue = make(chan string, 254)
	svc.DeletedSubnetQueue = make(chan string, 254)

	return svc, nil
}

func (svc *ApiServer) installApis() {
	container := restful.NewContainer()
	container.DoNotRecover(false)

	// Add api
	ping.AddToContainer(container)
	subnet.AddToContainer(container, svc.DB, svc.AddSubnetQueue, svc.ChangeSubnetQueue, svc.DeletedSubnetQueue, svc.Config.AceessIP, svc.Config.RRoutes)

	svc.handler = container
}

func (svc *ApiServer) initFilters() {
	svc.handler = filters.WithAccessLog(svc.handler)
	svc.handler = filters.WithOperationID(svc.handler)
}

func (svc *ApiServer) syncDb() error {
	if err := svc.DB.Sync(new(modelv1.Subnet)); err != nil {
		return err
	}

	if err := svc.DB.Sync(new(modelv1.Peer)); err != nil {
		return err
	}

	return nil
}

func (svc *ApiServer) syncWg() error {
	var subnets []modelv1.Subnet
	if err := svc.DB.Find(&subnets); err != nil {
		klog.Errorf("Failed to get subnets from db: %s", err.Error())
		return err
	}

	for _, subnet := range subnets {
		klog.V(4).Infof("Setup subnet: %s uuid: %s iface: %s", subnet.Name, subnet.Uuid, subnet.Iface)
		var peers []modelv1.Peer
		if err := svc.DB.Where("subnet = ?", subnet.Uuid).Find(&peers); err != nil {
			return fmt.Errorf("failed to get subnet %s uuid %s peers from db: %s", subnet.Name, subnet.Uuid, err.Error())
		}

		wg := &core.WG{
			Name:   subnet.Iface,
			CfgDir: svc.Config.BaseDir,
			Interface: &core.Iface{
				Address:    subnet.Addr,
				PrivateKey: subnet.PriKey,
				PublicKey:  subnet.PubKey,
				ListenPort: subnet.Port,
			},
			Peers: make([]*core.Peer, 0),
		}

		for _, peer := range peers {
			wg.Peers = append(wg.Peers, &core.Peer{
				Name:       peer.Name,
				PrivateKey: peer.PriKey,
				PublicKey:  peer.PubKey,
				AllowedIPs: svc.getPeerAllowdIps(peer.Addr),
			})
		}

		if err := wg.Up(); err != nil {
			return fmt.Errorf("failed to setup subnet: %s uuid: %s iface: %s: %s", subnet.Name, subnet.Uuid, subnet.Iface, err.Error())
		}
	}

	klog.Info("Sync wireguard finished")
	return nil
}

func (svc *ApiServer) PreRun(ctx context.Context) error {
	// Init db
	if err := svc.syncDb(); err != nil {
		return err
	}

	// Sync wireuard conf from db
	if err := svc.syncWg(); err != nil {
		return err
	}

	// Init http server
	svc.installApis()
	svc.initFilters()
	return nil
}

func (svc *ApiServer) Teardwon() {
	var subnets []modelv1.Subnet
	if err := svc.DB.Find(&subnets); err != nil {
		klog.Error(err)
		return
	}

	for _, subnet := range subnets {
		klog.V(4).Infof("Teardown subnet: %s uuid: %s iface: %s", subnet.Name, subnet.Uuid, subnet.Iface)
		wg := &core.WG{
			Name:      subnet.Iface,
			Interface: new(core.Iface),
		}

		if err := wg.Down(); err != nil {
			klog.Errorf("Teardown wireguard %s: %s", wg.Name, err.Error())
			continue
		}
	}
}

func (svc *ApiServer) getPeerAllowdIps(peerIP string) string {
	return fmt.Sprintf("%s/32", strings.Split(peerIP, "/")[0])
}

func (svc *ApiServer) processDeletedSubnet() {
	for {
		iface := <-svc.DeletedSubnetQueue
		klog.V(4).Infof("Start to process delete subent %s", iface)

		wg := &core.WG{
			Name:      iface,
			Interface: new(core.Iface),
		}

		if err := wg.Down(); err != nil {
			klog.Errorf("Failed to delete wireguard interface %s: %s", wg.Name, err.Error())
			continue
		}
	}
}

func (svc *ApiServer) processAddSubnet() {
	for {
		uid := <-svc.AddSubnetQueue
		klog.V(4).Infof("Start to process add subent %s", uid)

		subnet := new(modelv1.Subnet)
		ok, err := svc.DB.Where("uuid = ?", uid).Get(subnet)
		if err != nil {
			klog.Errorf("Failed to retrive subnet from db: %s", err.Error())
			continue
		}

		if !ok {
			klog.Warningf("Subnet %s not found", uid)
			continue
		}

		wg := &core.WG{
			Name:   subnet.Iface,
			CfgDir: svc.Config.BaseDir,
			Interface: &core.Iface{
				Address:    subnet.Addr,
				PrivateKey: subnet.PriKey,
				PublicKey:  subnet.PubKey,
				ListenPort: subnet.Port,
			},
			Peers: make([]*core.Peer, 0),
		}

		if err := wg.Up(); err != nil {
			klog.Errorf("Failed to setup subnet: %s uuid: %s iface: %s: %s", subnet.Name, subnet.Uuid, subnet.Iface, err.Error())
			continue
		}
	}
}

func (svc *ApiServer) processChangedSubnet() {
	for {
		uid := <-svc.ChangeSubnetQueue
		klog.V(4).Infof("Start to process change subent %s", uid)

		subnet := new(modelv1.Subnet)

		ok, err := svc.DB.Where("uuid = ?", uid).Get(subnet)
		if err != nil {
			klog.Errorf("Failed to retrive subnet from db: %s", err.Error())
			continue
		}

		if !ok {
			// Subent has beeen deleted, delete from system
			wg := &core.WG{
				Name:      subnet.Iface,
				CfgDir:    svc.Config.BaseDir,
				Interface: new(core.Iface),
			}

			if err := wg.Down(); err != nil {
				klog.Errorf("Failed to delete wireguard interface %s for subnet %s uuid %s: %s", subnet.Iface, subnet.Name, subnet.Uuid, err.Error())
				continue
			}

			continue
		}

		// Sync change
		var peers []modelv1.Peer
		if err := svc.DB.Where("subnet = ?", subnet.Uuid).Find(&peers); err != nil {
			klog.Errorf("Failed to get subnet %s uuid %s peers from db: %s", subnet.Name, subnet.Uuid, err.Error())
			continue
		}

		wg := &core.WG{
			Name:   subnet.Iface,
			CfgDir: svc.Config.BaseDir,
			Interface: &core.Iface{
				PrivateKey: subnet.PriKey,
				ListenPort: subnet.Port,
			},
			Peers: make([]*core.Peer, 0),
		}

		for _, peer := range peers {
			if !peer.Enable {
				continue
			}

			wg.Peers = append(wg.Peers, &core.Peer{
				Name:       peer.Name,
				PublicKey:  peer.PubKey,
				AllowedIPs: svc.getPeerAllowdIps(peer.Addr),
			})
		}

		if err := wg.Update(); err != nil {
			klog.Errorf("failed to sync config for interface %s with config %s", subnet.Iface, wg.CfgFilePath())
			continue
		}
	}
}

func (svc *ApiServer) Run(ctx context.Context) error {
	klog.Infof("Start listen on port %d ...", svc.Config.Port)

	// Goruntine to handle signal to teardown wireuard
	go func() {
		<-ctx.Done()
		klog.Info("Term signal received, teardown wireguard server, please wait ...\nCtrl+C to exit force")
		svc.Teardwon()
		klog.Info("Teardwon finished, to exist")
		os.Exit(0)
	}()

	// Precess next changed subnet
	go svc.processAddSubnet()
	go svc.processChangedSubnet()
	go svc.processDeletedSubnet()

	svc.Server.Handler = svc.handler
	if err := svc.Server.ListenAndServeTLS("", ""); err != nil {
		return err
	}

	return nil
}
