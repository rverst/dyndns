package dns

import (
  "fmt"
  dnsh "github.com/miekg/dns"
  "github.com/rs/zerolog/log"
  "github.com/rverst/dyndns/config"
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

  fn := zoneFilePath(zone)
  fi, err := os.Stat(fn)
  if err != nil {
    return createZoneFile(fn, zone, locations, ip4, ip6)
  }

  f, _ := os.Open(fn)
  zp := dnsh.NewZoneParser(f, "", "")
  for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
    // Do something with rr
    fmt.Println(rr.String())
  }

  if err := zp.Err(); err != nil {
    // log.Println(err)
  }
  fmt.Println("FILEINFO", fi)

  log.Info().Str("zone", zone).Strs("locations", locations).
    Interface("ip4", ip4).Interface("ip6", ip6).Msg("update zone")
  return nil
}

func createZoneFile(path string, zone string, locations []string, ip4 net.IP, ip6 net.IP) error {

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
    Serial:  DefaultSerial(),
    Refresh: 360,
    Retry:   180,
    Expire:  1800,
    Minttl:  180,
  }

  sb.WriteString(fmt.Sprintf("@%s", soa.String()))

  fmt.Println(sb.String())
  return nil
}
