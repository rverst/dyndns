package gandi

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strings"
)

type Gandi struct {
	apiKey string
}

func New() (*Gandi, error) {
	k := os.Getenv("DYNDNS_GANDI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("no api key provided")
	}
	return &Gandi{apiKey: k}, nil
}

func (g Gandi) Name() string {
	return "gandi/v5"
}

func (g Gandi) UpdateZone(zone string, locations []string, ip4, ip6 net.IP) error {
  zone = strings.TrimSuffix(zone, ".")
  split := strings.Split(zone, ".")
  if len(split) >= 2 {
    zone = fmt.Sprintf("%s.%s", split[len(split)-2], split[len(split)-1])
  }

  for i := range locations {
		locations[i] = strings.TrimSuffix(strings.TrimSuffix(dns.Fqdn(locations[i]), dns.Fqdn(zone)), ".")
	}
	fmt.Printf("UpdateZone: %s | %s \n", zone, locations)

	var err error
	for _, location := range locations {
		var recs []*rrset
		recs, err = getDomainRecords(zone, location, g.apiKey)
		if err != nil {
			fmt.Println(err)
			log.Error().Err(err).Msg("error fetching domain records")
		} else {
			if len(recs) == 0 {
				if ip4 != nil {
					err = addDomainRecord(zone, location, g.apiKey, rrset{
						RrsetType:   "A",
						RrsetTTL:    defaultTtl,
						RrsetName:   "",
						RrsetValues: []string{ip4.String()},
					})
					if err != nil {
						log.Error().Err(err).Str("zone", zone).Str("location", location).
							IPAddr("ip4", ip4).Msg("addDomainRecords")
					}
				}
				if ip6 != nil {
					err = addDomainRecord(zone, location, g.apiKey, rrset{
						RrsetType:   "AAAA",
						RrsetTTL:    defaultTtl,
						RrsetName:   "",
						RrsetValues: []string{ip6.String()},
					})
					if err != nil {
						log.Error().Err(err).Str("zone", zone).Str("location", location).
							IPAddr("ip6", ip6).Msg("addDomainRecords")
					}
				}

				continue
			}

			needUpdate := false
			for _, rec := range recs {
				rec.RrsetHref = ""
				rec.RrsetName = ""
				if rec.RrsetType == "A" {
					if len(rec.RrsetValues) > 0 {
						ip := net.ParseIP(rec.RrsetValues[0])
						if ip4 != nil && !ip4.Equal(ip) {
							needUpdate = true
							rec.RrsetValues[0] = ip4.String()
						}
					} else if ip4 != nil {
						needUpdate = true
						rec.RrsetValues = []string{ip4.String()}
					}
				} else if rec.RrsetType == "AAAA" {
					if len(rec.RrsetValues) > 0 {
						ip := net.ParseIP(rec.RrsetValues[0])
						if ip6 != nil && !ip6.Equal(ip) {
							needUpdate = true
							rec.RrsetValues[0] = ip6.String()
						}
					} else if ip6 != nil {
						needUpdate = true
						rec.RrsetValues = []string{ip6.String()}
					}
				}
			}
			if needUpdate {
				err = updateDomainRecords(zone, location, g.apiKey, recs)
				if err != nil {
					log.Error().Err(err).Str("zone", zone).Str("location", location).
						Interface("records", recs).IPAddr("ip4", ip4).IPAddr("ip6", ip6).
						Msg("updateDomainRecords")
				}
			} else {
			  log.Debug().Str("zone", zone).Str("location", location).
          Interface("records", recs).IPAddr("ip4", ip4).IPAddr("ip6", ip6).
          Msg("records up to date")
      }
		}
	}

	return err
}
