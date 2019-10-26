package handler

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery is a handler for handling panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("panic handling request", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
