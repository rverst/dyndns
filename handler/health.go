package handler

import (
	"net/http"
)

const acaoHeader = "Access-Control-Allow-Origin"

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(acaoHeader, "*")
	if r.Method == http.MethodOptions {
		return
	}
	w.WriteHeader(http.StatusOK)
}
