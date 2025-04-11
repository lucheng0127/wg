package filters

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	HeaderOperationID = "X-Operation-ID"
)

func WithOperationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderOperationID)

		uid, _ := uuid.NewUUID()
		if id == "" {
			r.Header.Set(HeaderOperationID, uid.String())
		}

		next.ServeHTTP(w, r)
	})
}
