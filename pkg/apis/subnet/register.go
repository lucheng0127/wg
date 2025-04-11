package subnet

import (
	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/wg/pkg/apis/subnet/v1alpha1"
	"xorm.io/xorm"
)

func AddToContainer(container *restful.Container, db *xorm.Engine) {
	v1alpha1.AddToContainer(container, db)
}
