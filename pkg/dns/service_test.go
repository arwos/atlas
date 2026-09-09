//nolint:testpackage // Internal helpers require package-private access.
package dns

import (
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

func TestRoutingPriority(t *testing.T) {
	zones := []Zone{{ID: 1, Name: "internal.example"}}
	forwarders := []Forwarder{{ZoneName: "", Address: "1.1.1.1:53"}, {ZoneName: "example", Address: "2.2.2.2:53"}, {ZoneName: "corp.example", Address: "3.3.3.3:53"}}
	if !blocked("www.ads.example", []Block{{Name: "ads.example"}}) {
		t.Fatal("subdomain must be blocked")
	}
	if zone, ok := findZone("host.internal.example", zones); !ok || zone.ID != 1 {
		t.Fatal("local zone was not selected")
	}
	if got := findForwarder("a.corp.example", forwarders); got != "3.3.3.3:53" {
		t.Fatalf("specific forwarder = %q", got)
	}
	if got := findForwarder("other.net", forwarders); got != "1.1.1.1:53" {
		t.Fatalf("default forwarder = %q", got)
	}
}

func TestAuthoritativeAnswers(t *testing.T) {
	zone := Zone{ID: 1, Name: "example.test", PrimaryNS: "ns.example.test", AdminEmail: "hostmaster@example.test", TTL: 60, Serial: 1}
	records := []Record{{ZoneID: 1, Name: "www.example.test", Type: "A", Value: "192.0.2.1", TTL: 60}}
	request := mdns.Question{Name: "www.example.test.", Qtype: mdns.TypeA, Qclass: mdns.ClassINET}
	response := new(mdns.Msg)
	(&Service{}).answerZone(response, request, zone, records)
	if !response.Authoritative || len(response.Answer) != 1 || response.Rcode != mdns.RcodeSuccess {
		t.Fatalf("unexpected authoritative response: %#v", response)
	}
	response = new(mdns.Msg)
	(&Service{}).answerZone(response, mdns.Question{Name: "www.example.test.", Qtype: mdns.TypeTXT, Qclass: mdns.ClassINET}, zone, records)
	if response.Rcode != mdns.RcodeSuccess || len(response.Answer) != 0 {
		t.Fatal("existing name without type must be NODATA")
	}
	response = new(mdns.Msg)
	(&Service{}).answerZone(response, mdns.Question{Name: "missing.example.test.", Qtype: mdns.TypeA, Qclass: mdns.ClassINET}, zone, records)
	if response.Rcode != mdns.RcodeNameError {
		t.Fatal("missing name must be NXDOMAIN")
	}
}

func TestCacheTTLs(t *testing.T) {
	message := new(mdns.Msg)
	message.Answer = []mdns.RR{&mdns.A{Hdr: mdns.RR_Header{Name: "example.test.", Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: 90}}}
	service := &Service{config: &Config{CacheEntries: 2, CacheMaxTTL: time.Hour}, cache: make(map[string]cacheEntry)}
	service.putCache("key", message)
	entry, ok := service.cache["key"]
	if !ok || time.Until(entry.expires) > 91*time.Second {
		t.Fatal("response TTL was not cached")
	}
	entry.stored = time.Now().Add(-30 * time.Second)
	decrementTTL(entry.message, uint32(time.Since(entry.stored).Seconds()))
	if got := entry.message.Answer[0].Header().Ttl; got < 59 || got > 60 {
		t.Fatalf("TTL = %d, want 60", got)
	}
}

func TestValidation(t *testing.T) {
	if err := validateRecord(Record{ZoneID: 1, Name: "_sip.example.test", Type: "SRV", Value: "10 5060 sip.example.test", TTL: 60}); err != nil {
		t.Fatalf("valid SRV rejected: %v", err)
	}
	if err := validateRecord(Record{ZoneID: 1, Name: "example.test", Type: "A", Value: "invalid", TTL: 60}); err == nil {
		t.Fatal("invalid A accepted")
	}
}
