package filters

import (
	"net/http"

	"k8s.io/klog/v2"
)

func WithAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("%s [%s] %s", r.Header.Get(HeaderOperationID), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
