package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		r.Header.Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r)
	})
}
