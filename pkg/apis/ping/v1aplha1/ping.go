package v1aplha1

import "github.com/emicklei/go-restful"

func ping(req *restful.Request, rsp *restful.Response) {
	rsp.WriteEntity("pong")
}
