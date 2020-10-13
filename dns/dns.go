package dns

import (
	"fmt"
	dnsh "github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"github.com/rverst/dyndns/config"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var dataP = ""

func dataPath() string {
	if dataP == "" {
		dataP = os.Getenv("DYNDNS_DATA")
		if dataP == "" {
			if d, err := os.Getwd(); err == nil {
				dataP = filepath.Join(d, "data")
			} else {
				dataP = "./data"
			}
		}
	}
	return dataP
}

func zoneFilePath(zone string) string {
	return filepath.Join(dataPath(), fmt.Sprintf("db.%s", strings.Trim(zone, ".")))
}

func UpdateZone(zone string, locations []string, ip4, ip6 net.IP) error {

	for i := range locations {
		locations[i] = strings.TrimSuffix(strings.TrimSuffix(dnsh.Fqdn(locations[i]), dnsh.Fqdn(zone)), ".")
	}

	fn := zoneFilePath(zone)
	_, err := os.Stat(fn)
	if err != nil {
		return writeZoneFile(fn, zone, DefaultSerial(), locations, ip4, ip6)
	}

	f, _ := os.Open(fn)
	zp := dnsh.NewZoneParser(f, "", "")
	serial := int64(-1)
	oldLoc := make([]string, 0)
	var oldIp4, oldIp6 net.IP

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		if err := zp.Err(); err != nil {
			return err
		}
		if serial < 0 && rr.Header().Rrtype == dnsh.TypeSOA {
			if s, ok := rr.(*dnsh.SOA); ok {
				serial = int64(s.Serial)
			}
		}
		if oldIp4 == nil && rr.Header().Rrtype == dnsh.TypeA {
			if a, ok := rr.(*dnsh.A); ok {
				oldIp4 = a.A
			}
		}
		if oldIp6 == nil && rr.Header().Rrtype == dnsh.TypeAAAA {
			if a, ok := rr.(*dnsh.AAAA); ok {
				oldIp6 = a.AAAA
			}
		}
		added := false
		ll := strings.TrimSuffix(strings.TrimSuffix(dnsh.Fqdn(rr.Header().Name), dnsh.Fqdn(zone)), ".")
		for _, z := range oldLoc {
			if z == ll {
				added = true
				break
			}
		}
		if !added {
			oldLoc = append(oldLoc, ll)
		}
	}
	if serial < 0 {
		serial = int64(DefaultSerial())
	}

	needUpdate := false
	sl := 0
	for _, l := range locations {
		for _, ol := range oldLoc {
			if l == ol {
				sl++
				break
			}
		}
	}
	if sl != len(locations) || sl != len(oldLoc) {
		needUpdate = true
	}
	if !ip4.Equal(oldIp4) {
		needUpdate = true
	}
	if !ip6.Equal(oldIp6) {
		needUpdate = true
	}

	if needUpdate {
		newSer, err := IncrementSerial(uint32(serial))
		if err != nil {
			newSer = DefaultSerial()
		}
		err = writeZoneFile(fn, zone, newSer, locations, ip4, ip6)
		if err != nil {
			return err
		}

	}

	log.Info().Str("zone", zone).Strs("locations", locations).
		Interface("ip4", ip4).Interface("ip6", ip6).Msg("update zone")
	return nil
}

func writeZoneFile(path string, zone string, serial uint32, locations []string, ip4 net.IP, ip6 net.IP) error {

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("$ORIGIN %s\n\n", zone))

	soa := dnsh.SOA{
		Hdr: dnsh.RR_Header{
			Rrtype: dnsh.TypeSOA,
			Class:  dnsh.ClassINET,
			Ttl:    180,
		},
		Ns:      dnsh.Fqdn(config.GetNameserver()),
		Mbox:    fmt.Sprintf(config.GetMailbox()),
		Serial:  serial,
		Refresh: 360,
		Retry:   180,
		Expire:  1800,
		Minttl:  180,
	}
	sb.WriteString(fmt.Sprintf("@%s\n", soa.String()))

	for _, l := range locations {
		addRec(&sb, l, ip4, ip6)
	}
	err := ioutil.WriteFile(path, []byte(sb.String()), 0664)
	return err
}

func addRec(sb *strings.Builder, l string, ip4 net.IP, ip6 net.IP) {
	if ip4 != nil {
		a := dnsh.A{
			Hdr: dnsh.RR_Header{
				Name:   l,
				Rrtype: dnsh.TypeA,
				Class:  dnsh.ClassINET,
				Ttl:    180,
			},
			A: ip4,
		}
		sb.WriteString(fmt.Sprintf("%s\n", a.String()))
	}
	if ip6 != nil {
		aaaa := dnsh.AAAA{
			Hdr: dnsh.RR_Header{
				Name:   l,
				Rrtype: dnsh.TypeAAAA,
				Class:  dnsh.ClassINET,
				Ttl:    180,
			},
			AAAA: ip6,
		}
		sb.WriteString(fmt.Sprintf("%s\n", aaaa.String()))
	}
}
