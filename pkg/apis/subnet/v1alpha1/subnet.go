package v1alpha1

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	"github.com/lucheng0127/wg/pkg/core"
	modelv1 "github.com/lucheng0127/wg/pkg/models/v1alpha1"
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

	// TODO(shawnlu): Teardown wireguard interface

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

	// TODO(shawnlu): Set wireguard interface up

	rsp.WriteEntity(map[string]string{
		"name":  subnet.Name,
		"uuid":  subnet.Uuid,
		"iface": subnet.Iface,
		"addr":  subnet.Addr,
	})
}
