package middleware

import (
	"encoding/json"
	"github.com/devenairevo/todoapi/handlers"
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	const apiKey = "super-secret-api-key"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		providedKey := r.Header.Get("X-API-Key")
		if providedKey != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized: Missing or invalid API Key"})
			if err != nil {
				handlers.HandleError(w, err)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}
