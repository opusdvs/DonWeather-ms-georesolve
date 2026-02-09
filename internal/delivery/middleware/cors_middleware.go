package middleware

import (
	"net/http"
	"os"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
