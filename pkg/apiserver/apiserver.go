package apiserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/emicklei/go-restful"
	_ "github.com/mattn/go-sqlite3"
	"k8s.io/klog/v2"
	"xorm.io/xorm"

	"github.com/lucheng0127/wg/pkg/apis/ping"
	"github.com/lucheng0127/wg/pkg/apis/subnet"
	"github.com/lucheng0127/wg/pkg/config"
	"github.com/lucheng0127/wg/pkg/filters"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
)

type ApiServer struct {
	Config  *config.ApiserverConf
	DB      *xorm.Engine
	Server  *http.Server
	handler http.Handler
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

	tlsConf := &tls.Config{Certificates: []tls.Certificate{cer}}
	httpSvc := new(http.Server)
	httpSvc.TLSConfig = tlsConf
	httpSvc.Addr = fmt.Sprintf(":%d", svc.Config.Port)
	svc.Server = httpSvc

	return svc, nil
}

func (svc *ApiServer) installApis() {
	container := restful.NewContainer()
	container.DoNotRecover(false)

	// TODO(shawnlu): Add api
	ping.AddToContainer(container)
	subnet.AddToContainer(container, svc.DB)

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

func (svc *ApiServer) PreRun(ctx context.Context) error {
	// Init db
	if err := svc.syncDb(); err != nil {
		return err
	}

	// TODO(shawnlu): Sync wireuard conf from db

	// Init http server
	svc.installApis()
	svc.initFilters()
	return nil
}

func (svc *ApiServer) Run(ctx context.Context) error {
	klog.Infof("Start listen on port %d ...", svc.Config.Port)

	// TODO(shwnlu): Goruntine to handle signal to teardown wireuard

	svc.Server.Handler = svc.handler
	if err := svc.Server.ListenAndServeTLS("", ""); err != nil {
		return err
	}

	return nil
}
