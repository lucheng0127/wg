package request

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lucheng0127/wg/pkg/config"
)

var (
	URL_SUBNETS     = "/api/subnet/v1alpha1/subnets"
	URL_SUBNET      = "/api/subnet/v1alpha1/subnets/%s"
	URL_PEERS       = "/api/subnet/v1alpha1/subnets/%s/peers"
	URL_PEER        = "/api/subnet/v1alpha1/subnets/%s/peers/%s"
	URL_PEER_ENABLE = "/api/subnet/v1alpha1/subnets/%s/peers/%s/enable"
	URL_PEER_CONFIG = "/api/subnet/v1alpha1/subnets/%s/peers/%s/config"
)

type ReqMgr struct {
	Host  string
	Cofig *config.WgCtlConf
}

func (r *ReqMgr) getCerts() (caCrt, cliCrt, cliKey []byte, err error) {
	caCrt, err = base64.StdEncoding.DecodeString(r.Cofig.CaCrt)
	if err != nil {
		return
	}

	cliCrt, err = base64.StdEncoding.DecodeString(r.Cofig.ClientCrt)
	if err != nil {
		return
	}

	cliKey, err = base64.StdEncoding.DecodeString(r.Cofig.ClientKey)
	if err != nil {
		return
	}

	return
}

func (r *ReqMgr) getTransport() (*http.Transport, error) {
	if r.Cofig == nil {

		tlsCfg := &tls.Config{
			InsecureSkipVerify: true,
		}
		transport := &http.Transport{
			TLSClientConfig: tlsCfg,
		}

		return transport, nil
	}

	// Load key
	svcCrt, cliCrt, cliKey, err := r.getCerts()
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(svcCrt)
	if !ok {
		return nil, errors.New("failed to add ca crt to pool")
	}

	cert, err := tls.X509KeyPair(cliCrt, cliKey)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
		},
	}

	return transport, nil
}

func (r *ReqMgr) SubnetPeerList(uid string) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	rsp, err := client.Get(r.Host + fmt.Sprintf(URL_PEERS, uid))
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) SubnetList() ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	rsp, err := client.Get(r.Host + URL_SUBNETS)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) SubnetDel(uid string) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	req, err := http.NewRequest("DELETE", r.Host+fmt.Sprintf(URL_SUBNET, uid), nil)
	if err != nil {
		return make([]byte, 0), err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) SubnetAdd(payload []byte) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	buf := bytes.NewBuffer(payload)
	rsp, err := client.Post(r.Host+URL_SUBNETS, "application/json", buf)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) PeerDel(sid, pid string) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	req, err := http.NewRequest("DELETE", r.Host+fmt.Sprintf(URL_PEER, sid, pid), nil)
	if err != nil {
		return make([]byte, 0), err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) PeerConf(sid, pid string) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	rsp, err := client.Get(r.Host + fmt.Sprintf(URL_PEER_CONFIG, sid, pid))
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}

func (r *ReqMgr) PeerSet(sid, pid string, payload []byte) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	buf := bytes.NewBuffer(payload)
	rsp, err := client.Post(r.Host+fmt.Sprintf(URL_PEER_ENABLE, sid, pid), "application/json", buf)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil

}

func (r *ReqMgr) PeerAdd(payload []byte, sid string) ([]byte, error) {
	client := new(http.Client)
	tran, err := r.getTransport()
	if err != nil {
		return make([]byte, 0), err
	}
	client.Transport = tran

	buf := bytes.NewBuffer(payload)
	rsp, err := client.Post(r.Host+fmt.Sprintf(URL_PEERS, sid), "application/json", buf)
	if err != nil {
		return make([]byte, 0), err
	}

	rst, err := io.ReadAll(rsp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return rst, nil
}
