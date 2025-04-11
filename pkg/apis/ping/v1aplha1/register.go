package v1aplha1

import "github.com/emicklei/go-restful"

func AddToContainer(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/ping").To(ping).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON))

	container.Add(ws)
}
