package v1alpha1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	"github.com/lucheng0127/wg/pkg/core"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
	"github.com/lucheng0127/wg/pkg/utils/validate"
	"github.com/vishvananda/netlink"
	"xorm.io/xorm"
)

type handler struct {
	DB *xorm.Engine
}

func newHandler(db *xorm.Engine) *handler {
	return &handler{db}
}

func (h *handler) subnetList(req *restful.Request, rsp *restful.Response) {
	var subnets []modelv1.Subnet
	if err := h.DB.Find(&subnets); err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	rst := make([]map[string]string, 0)
	for _, s := range subnets {
		rst = append(rst, map[string]string{
			"name":   s.Name,
			"uuid":   s.Uuid,
			"iface":  s.Iface,
			"addr":   s.Addr,
			"pubkey": s.PubKey,
		})
	}

	rsp.WriteEntity(rst)
}

func (h *handler) subnetDelete(req *restful.Request, rsp *restful.Response) {
	uid := req.PathParameter("uuid")
	if _, err := h.DB.Delete(&modelv1.Peer{Subnet: uid}); err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := h.DB.Delete(&modelv1.Subnet{Uuid: uid}); err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO(shawnlu): Send subnet uuid to channel to sync conf

	rsp.WriteEntity(true)
}

func (h *handler) subnetCreate(req *restful.Request, rsp *restful.Response) {
	subnet := new(modelv1.Subnet)
	if err := req.ReadEntity(subnet); err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	if subnet.Name == "" || subnet.Iface == "" || subnet.Addr == "" {
		rsp.WriteErrorString(http.StatusBadRequest, "name, iface, addr needed")
		return
	}

	if strings.Contains(subnet.Iface, " ") {
		rsp.WriteErrorString(http.StatusBadRequest, "invalid interface name")
		return
	}

	addr, err := netlink.ParseAddr(subnet.Addr)
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if addr.IP.To4() == nil {
		rsp.WriteErrorString(http.StatusBadRequest, "only ipv4 address supported")
		return
	}

	key, err := core.NewRandomKey()
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	subnet.Uuid = uid.String()
	subnet.PubKey = key.Pub
	subnet.PriKey = key.Priv
	if _, err := h.DB.Insert(subnet); err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	// TODO(shawnlu): Send subnet uuid to channel to sync conf

	rsp.WriteEntity(map[string]string{
		"name":  subnet.Name,
		"uuid":  subnet.Uuid,
		"iface": subnet.Iface,
		"addr":  subnet.Addr,
	})
}

func (h *handler) subnetPeerList(req *restful.Request, rsp *restful.Response) {
	subnet := req.PathParameter("subnet")

	var peers []modelv1.Peer
	if err := h.DB.Where("subnet = ?", subnet).Find(&peers); err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	rst := make([]map[string]string, 0)
	for _, p := range peers {
		enable := "False"
		if p.Enable {
			enable = "True"
		}
		rst = append(rst, map[string]string{
			"name":   p.Name,
			"uuid":   p.Uuid,
			"subnet": p.Subnet,
			"addr":   p.Addr,
			"pubkey": p.PubKey,
			"enable": enable,
		})
	}

	// TODO(shawnlu): Send subnet uuid to channel to sync conf

	rsp.WriteEntity(rst)
}

func (h *handler) subnetPeerCreate(req *restful.Request, rsp *restful.Response) {
	subnetId := req.PathParameter("subnet")
	peer := new(modelv1.Peer)
	subnet := new(modelv1.Subnet)
	if err := req.ReadEntity(peer); err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	if peer.Name == "" || peer.Addr == "" {
		rsp.WriteErrorString(http.StatusBadRequest, "name, addr needed")
		return
	}

	ok, err := h.DB.Where("uuid = ?", subnetId).Get(subnet)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		rsp.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("subnet %s not exist", subnetId))
		return
	}

	existPeer := new(modelv1.Peer)
	existPeer.Subnet = subnetId
	existPeer.Name = peer.Name
	ok, err = h.DB.Exist(existPeer)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	if ok {
		rsp.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("peer %s exist in subnet %s", peer.Name, subnetId))
		return
	}

	if !validate.ValidatePeerAddr(subnet.Addr, peer.Addr) {
		rsp.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("invalidate peer address %s", peer.Addr))
		return
	}

	key, err := core.NewRandomKey()
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	peer.Uuid = uid.String()
	peer.Subnet = subnet.Uuid
	peer.PubKey = key.Pub
	peer.PriKey = key.Priv
	peer.Enable = true
	if _, err := h.DB.Insert(peer); err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	// TODO(shawnlu): Send subnet uuid to channel to sync conf

	rsp.WriteEntity(map[string]string{
		"name":   peer.Name,
		"uuid":   peer.Uuid,
		"addr":   peer.Addr,
		"subnet": peer.Subnet,
	})
}

func (h *handler) subnetPeerDelete(req *restful.Request, rsp *restful.Response) {
	subnetId := req.PathParameter("subnet")
	peerId := req.PathParameter("peer")

	if _, err := h.DB.Delete(&modelv1.Peer{Subnet: subnetId, Uuid: peerId}); err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	rsp.WriteEntity(true)
}

type EnableReq struct {
	Enable bool `json:"enable"`
}

func (h *handler) subnetPeerEnable(req *restful.Request, rsp *restful.Response) {
	subnetId := req.PathParameter("subnet")
	peerId := req.PathParameter("peer")
	ePars := new(EnableReq)
	if err := req.ReadEntity(ePars); err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	peer := new(modelv1.Peer)
	peer.Subnet = subnetId
	peer.Uuid = peerId
	ok, err := h.DB.Get(peer)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		rsp.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("peer %s in subnet %s not exist", peerId, subnetId))
		return
	}

	needUpdate := false
	if peer.Enable != ePars.Enable {
		peer.Enable = ePars.Enable
		needUpdate = true
	}

	if needUpdate {
		if _, err := h.DB.Where("uuid = ?", peer.Uuid).Cols("enable").Update(peer); err != nil {
			rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
			return
		}
	}

	rsp.WriteEntity(true)
}
