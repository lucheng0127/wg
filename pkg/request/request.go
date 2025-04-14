package request

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
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
	Host string
}

func skipVerifyTran() *http.Transport {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return transport
}

func (r *ReqMgr) SubnetPeerList(uid string) ([]byte, error) {
	client := new(http.Client)
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
	client.Transport = skipVerifyTran()

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
