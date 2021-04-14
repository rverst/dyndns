package dns

import (
  "fmt"
  "github.com/rs/zerolog/log"
  "net"
)

var providers []*Provider

type Provider interface {
  Name() string
  UpdateZone(zone string, locations []string, ip4, ip6 net.IP) error
}

func AddProvider(p Provider) {
  if providers == nil {
    providers = make([]*Provider, 0)
  }

  providers = append(providers, &p)

}

func UpdateZone(zone string, locations []string, ip4, ip6 net.IP) error {

  if providers == nil {
    return fmt.Errorf("no providers")
  }
  var err error
  for _, p := range providers {
    err = (*p).UpdateZone(zone, locations, ip4, ip6)
    if err != nil {
      log.Error().Err(err).Str("zone", zone).Strs("locations", locations).Str("provider", (*p).Name()).
        Interface("ip4", ip4).Interface("ip6", ip6).Msg("update zone")
    } else {
      log.Info().Str("zone", zone).Strs("locations", locations).Str("provider", (*p).Name()).
        Interface("ip4", ip4).Interface("ip6", ip6).Msg("update zone")
    }
  }
  return err
}
