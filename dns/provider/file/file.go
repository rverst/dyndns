package file

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/rverst/dyndns/config"
	dns2 "github.com/rverst/dyndns/dns"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	dataPath string
}

func New() *File {
	p := os.Getenv("DYNDNS_FILE_DATA")
	if p == "" {
		if d, err := os.Getwd(); err == nil {
			p = filepath.Join(d, "data")
		} else {
			p = "./data"
		}
	}
	f := &File{
		dataPath: p,
	}
	return f
}

func (f File)zoneFilePath(zone string) string {
	return filepath.Join(f.dataPath, fmt.Sprintf("db.%s", strings.Trim(zone, ".")))
}

func (f File) Name() string {
	return "file"
}

func (f File) UpdateZone(zone string, locations []string, ip4, ip6 net.IP) error {

	for i := range locations {
		locations[i] = strings.TrimSuffix(strings.TrimSuffix(dns.Fqdn(locations[i]), dns.Fqdn(zone)), ".")
	}

	fn := f.zoneFilePath(zone)
	_, err := os.Stat(fn)
	if err != nil {
		return writeZoneFile(fn, zone, dns2.DefaultSerial(), locations, ip4, ip6)
	}

	file, _ := os.Open(fn)
	defer file.Close()
	zp := dns.NewZoneParser(file, "", "")
	serial := int64(-1)
	oldLoc := make([]string, 0)
	var oldIp4, oldIp6 net.IP

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		if err := zp.Err(); err != nil {
			return err
		}
		if serial < 0 && rr.Header().Rrtype == dns.TypeSOA {
			if s, ok := rr.(*dns.SOA); ok {
				serial = int64(s.Serial)
			}
		}
		if oldIp4 == nil && rr.Header().Rrtype == dns.TypeA {
			if a, ok := rr.(*dns.A); ok {
				oldIp4 = a.A
			}
		}
		if oldIp6 == nil && rr.Header().Rrtype == dns.TypeAAAA {
			if a, ok := rr.(*dns.AAAA); ok {
				oldIp6 = a.AAAA
			}
		}
		added := false
		ll := strings.TrimSuffix(strings.TrimSuffix(dns.Fqdn(rr.Header().Name), dns.Fqdn(zone)), ".")
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
		serial = int64(dns2.DefaultSerial())
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
		newSer, err := dns2.IncrementSerial(uint32(serial))
		if err != nil {
			newSer = dns2.DefaultSerial()
		}
		err = writeZoneFile(fn, zone, newSer, locations, ip4, ip6)
		if err != nil {
			return err
		}

	}

	return nil
}

func writeZoneFile(path string, zone string, serial uint32, locations []string, ip4 net.IP, ip6 net.IP) error {

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("$ORIGIN %s\n\n", zone))

	soa := dns.SOA{
		Hdr: dns.RR_Header{
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    180,
		},
		Ns:      dns.Fqdn(config.GetNameserver()),
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
		a := dns.A{
			Hdr: dns.RR_Header{
				Name:   l,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    180,
			},
			A: ip4,
		}
		sb.WriteString(fmt.Sprintf("%s\n", a.String()))
	}
	if ip6 != nil {
		aaaa := dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   l,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    180,
			},
			AAAA: ip6,
		}
		sb.WriteString(fmt.Sprintf("%s\n", aaaa.String()))
	}
}
