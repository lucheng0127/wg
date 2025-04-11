package v1alpha1

import (
	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/wg/pkg/utils/runtime"
	"xorm.io/xorm"
)

const (
	group   = "subnet"
	version = "v1alpha1"
)

func AddToContainer(container *restful.Container, db *xorm.Engine) {
	ws := runtime.NewApiWebService(group, version)
	ws.Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	handler := newHandler(db)

	ws.Route(ws.POST("/subnets").To(handler.subnetCreate))
	ws.Route(ws.GET("/subnets").To(handler.subnetList))
	ws.Route(ws.DELETE("/subnets/{uuid}").To(handler.subnetDelete).Param(ws.PathParameter("uuid", "uuid of subnet").DataType("string").Required(true)))
	ws.Route(ws.GET("/subnets/{subnet}/peers").To(handler.subnetPeerList).Param(ws.PathParameter("subnet", "uuid of subnet").DataType("string").Required(true)))
	ws.Route(ws.POST("/subnets/{subnet}/peers").To(handler.subnetPeerCreate).Param(ws.PathParameter("subnet", "uuid of subnet").DataType("string").Required(true)))
	ws.Route(ws.DELETE("/subnets/{subnet}/peers/{peer}").
		To(handler.subnetPeerDelete).
		Param(ws.PathParameter("subnet", "uuid of subnet").DataType("string").Required(true)).
		Param(ws.PathParameter("peer", "uuid of peer").DataType("string").Required(true)))
	ws.Route(ws.POST("/subnets/{subnet}/peers/{peer}/enable").
		To(handler.subnetPeerEnable).
		Param(ws.PathParameter("subnet", "uuid of subnet").DataType("string").Required(true)).
		Param(ws.PathParameter("peer", "uuid of peer").DataType("string").Required(true)))

	container.Add(ws)
}
