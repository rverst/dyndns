package handler

import (
  "fmt"
  dnsh "github.com/miekg/dns"
  "github.com/rs/zerolog/log"
  "github.com/rverst/dyndns/config"
  "github.com/rverst/dyndns/dns"
  "net"
  "net/http"
  "strings"
)

func FritzboxHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set(acaoHeader, "*")
  if r.Method == http.MethodOptions {
    return
  }

  u, p, ok := r.BasicAuth()
  if !ok {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  log.Info().Str("user", u).Interface("query", r.URL.Query()).Msg(r.URL.Path)

  user, err := config.GetUserBasic(u, p)
  if err == config.ErrAuth {
    log.Warn().Err(err).Str("username", u).Interface("ip", getUserIp(r)).
      Msg("authentication failed")
    w.WriteHeader(http.StatusUnauthorized)
    return
  } else if err != nil {
    log.Error().Err(err).Str("username", u).Interface("ip", getUserIp(r)).
      Msg("authentication error")
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  ip4 := net.ParseIP(r.URL.Query().Get("ip4"))
  ip6 := net.ParseIP(r.URL.Query().Get("ip6"))
  if ip4 == nil && ip6 == nil {
    log.Warn().Err(err).Str("username", u).Interface("ip", getUserIp(r)).
      Msg("no ip")
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  domains := strings.Split(r.URL.Query().Get("domain"), ",")

  zones := make(map[string][]string)
  c := 0
  for _, zone := range user.Zones {
    z := dnsh.Fqdn(zone)
    dd := make([]string, 0)
    for _, d := range domains {
      if dnsh.IsSubDomain(z, dnsh.Fqdn(d)) {
        dd = append(dd, dnsh.Fqdn(d))
        c++
      }
    }
    zones[z] = dd
  }
  if c == 0 {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  ers := make([]error, 0)
  for k, v := range zones {
    if len(v) == 0 {
      continue
    }
    if e := dns.UpdateZone(k, v, ip4, ip6); e != nil {
      ers = append(ers, e)
    }

  }

  if len(ers) > 0 {
    w.WriteHeader(http.StatusInternalServerError)
    for _, e := range ers {
      _, _ = w.Write([]byte(fmt.Sprintf("%s\n", e)))
    }
    return
  }

  w.WriteHeader(http.StatusOK)
}
