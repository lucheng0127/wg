package ping

import (
	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/wg/pkg/apis/ping/v1aplha1"
)

func AddToContainer(container *restful.Container) {
	v1aplha1.AddToContainer(container)
}
