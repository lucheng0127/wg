package runtime

import (
	"fmt"

	"github.com/emicklei/go-restful"
)

const (
	ApiPrefix = "/api"
)

func NewApiWebService(group, version string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(fmt.Sprintf("%s/%s/%s", ApiPrefix, group, version))
	return ws
}
