package handler

import (
	"net"
	"net/http"
)

func FritzboxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(acaoHeader, "*")
	if r.Method == http.MethodOptions {
		return
	}
	w.WriteHeader(http.StatusOK)

	ip4 := net.ParseIP(r.URL.Query().Get("ip4"))
	ip6 := net.ParseIP(r.URL.Query().Get("ip6"))
	loc := r.URL.Query().Get("domain")



}
