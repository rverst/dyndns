package handler

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

func LogMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Trace().Interface("ip", getUserIp(r)).Str("method", r.Method).
				Str("path", r.URL.Path).Send()
			next.ServeHTTP(w, r)
		})
	}
}
