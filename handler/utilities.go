package handler

import (
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
)

func getUserIp(r *http.Request) net.IP {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = ""
			log.Error().Err(err).Msgf("unable to get user ip: %s", r.RemoteAddr)
		}
	}
	return net.ParseIP(ip)
}
