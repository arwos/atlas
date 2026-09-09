//nolint:revive // Public methods form the DNS service boundary consumed by app.
package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	mdns "github.com/miekg/dns"
	"go.osspkg.com/logx"

	"go.arwos.org/atlas/pkg/database"
)

type cacheEntry struct {
	message *mdns.Msg
	stored  time.Time
	expires time.Time
}

const (
	srvValueParts   = 3
	soaRefresh      = 3600
	soaRetry        = 600
	soaExpire       = 86400
	upstreamTimeout = 5 * time.Second
)

type Service struct {
	repo   *repository
	config *Config
	mu     sync.RWMutex
	active Snapshot
	cache  map[string]cacheEntry
	udp    *mdns.Server
	tcp    *mdns.Server
}

func NewService(db *database.Service, cfg *ConfigGroup) *Service {
	return &Service{repo: &repository{db: db}, config: &cfg.Config, cache: make(map[string]cacheEntry)}
}

func (s *Service) Up(ctx context.Context) error {
	if err := s.reload(ctx); err != nil {
		return err
	}
	handler := mdns.HandlerFunc(s.serveDNS)
	packetConn, err := net.ListenPacket("udp", s.config.Listen)
	if err != nil {
		return fmt.Errorf("dns: listen UDP: %w", err)
	}
	listener, err := net.Listen("tcp", s.config.Listen)
	if err != nil {
		if closeErr := packetConn.Close(); closeErr != nil {
			logx.Error("DNS service", "do", "close UDP listener", "err", closeErr)
		}
		return fmt.Errorf("dns: listen TCP: %w", err)
	}
	s.udp = &mdns.Server{PacketConn: packetConn, Net: "udp", Handler: handler}
	s.tcp = &mdns.Server{Listener: listener, Net: "tcp", Handler: handler}
	go s.serve(s.udp, "UDP")
	go s.serve(s.tcp, "TCP")
	return nil
}

func (s *Service) serve(server *mdns.Server, protocol string) {
	if err := server.ActivateAndServe(); err != nil {
		logx.Error("DNS service", "do", "serve "+protocol, "err", err)
	}
}

func (s *Service) Down() error {
	var first error
	if s.udp != nil {
		if err := s.udp.Shutdown(); err != nil {
			first = err
		}
	}
	if s.tcp != nil {
		if err := s.tcp.Shutdown(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (s *Service) reload(ctx context.Context) error {
	snap, err := s.repo.snapshot(ctx, activeVersion)
	if err != nil {
		return err
	}
	if err = s.validate(snap); err != nil {
		return err
	}
	s.mu.Lock()
	s.active = snap
	s.cache = make(map[string]cacheEntry)
	s.mu.Unlock()
	return nil
}

func (s *Service) Draft(ctx context.Context) (Snapshot, error) {
	return s.repo.snapshot(ctx, draftVersion)
}

func (s *Service) Active(ctx context.Context) (Snapshot, error) {
	return s.repo.snapshot(ctx, activeVersion)
}

func (s *Service) UpsertZone(ctx context.Context, x Zone) error {
	if err := validateZone(x); err != nil {
		return err
	}
	return s.repo.upsertZone(ctx, x)
}
func (s *Service) DeleteZone(ctx context.Context, id int64) error { return s.repo.deleteZone(ctx, id) }
func (s *Service) UpsertRecord(ctx context.Context, x Record) error {
	if err := validateRecord(x); err != nil {
		return err
	}
	return s.repo.upsertRecord(ctx, x)
}

func (s *Service) DeleteRecord(ctx context.Context, id int64) error {
	return s.repo.deleteRecord(ctx, id)
}

func (s *Service) UpsertForwarder(ctx context.Context, x Forwarder) error {
	if err := validateForwarder(x); err != nil {
		return err
	}
	return s.repo.upsertForwarder(ctx, x)
}

func (s *Service) DeleteForwarder(ctx context.Context, id int64) error {
	return s.repo.deleteForwarder(ctx, id)
}

func (s *Service) UpsertBlock(ctx context.Context, x Block) error {
	n, err := normalName(x.Name)
	if err != nil {
		return err
	}
	x.Name = n
	return s.repo.upsertBlock(ctx, x)
}

func (s *Service) DeleteBlock(ctx context.Context, id int64) error {
	return s.repo.deleteBlock(ctx, id)
}

func (s *Service) Apply(ctx context.Context) error {
	snap, err := s.Draft(ctx)
	if err != nil {
		return err
	}
	if err = s.validate(snap); err != nil {
		return err
	}
	if err = s.repo.replaceActive(ctx); err != nil {
		return err
	}
	return s.reload(ctx)
}

//nolint:revive // Validation is kept together to preserve snapshot invariants.
func (s *Service) validate(snapshot Snapshot) error {
	zones := map[int64]Zone{}
	for _, x := range snapshot.Zones {
		if err := validateZone(x); err != nil {
			return err
		}
		zones[x.ID] = x
	}
	for _, x := range snapshot.Records {
		zone, ok := zones[x.ZoneID]
		if !ok {
			return errors.New("dns: record refers to unknown zone")
		}
		if x.Name != zone.Name && !strings.HasSuffix(x.Name, "."+zone.Name) {
			return errors.New("dns: record name is outside its zone")
		}
		if err := validateRecord(x); err != nil {
			return err
		}
	}
	seenDefault := false
	for _, x := range snapshot.Forwarders {
		if err := validateForwarder(x); err != nil {
			return err
		}
		if x.ZoneName == "" {
			if seenDefault {
				return errors.New("dns: multiple default forwarders")
			}
			seenDefault = true
		}
	}
	for _, x := range snapshot.Blocks {
		if _, err := normalName(x.Name); err != nil {
			return err
		}
	}
	return nil
}

func normalName(v string) (string, error) {
	v = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(v)), ".")
	if v == "" || len(v) > 253 {
		return "", fmt.Errorf("dns: invalid name %q", v)
	}
	for _, label := range strings.Split(v, ".") {
		if len(label) == 0 || len(label) > 63 {
			return "", fmt.Errorf("dns: invalid name %q", v)
		}
	}
	return v, nil
}

func validateZone(x Zone) error {
	var err error
	if x.Name, err = normalName(x.Name); err != nil {
		return err
	}
	if _, err = normalName(x.PrimaryNS); err != nil {
		return err
	}
	if !strings.Contains(x.AdminEmail, "@") || x.TTL == 0 {
		return errors.New("dns: primary NS, admin email and positive TTL are required")
	}
	return nil
}

//nolint:revive // Record validation is type-specific by DNS design.
func validateRecord(x Record) error {
	if _, err := normalName(x.Name); err != nil {
		return err
	}
	if x.ZoneID <= 0 || x.TTL == 0 {
		return errors.New("dns: record zone and positive TTL are required")
	}
	switch strings.ToUpper(x.Type) {
	case "A":
		if _, e := netip.ParseAddr(x.Value); e != nil {
			return e
		}
	case "AAAA":
		ip, e := netip.ParseAddr(x.Value)
		if e != nil || !ip.Is6() {
			return errors.New("dns: invalid AAAA record")
		}
	case "CNAME", "MX", "NS":
		if _, e := normalName(x.Value); e != nil {
			return e
		}
	case "SRV":
		parts := strings.Fields(x.Value)
		if len(parts) != srvValueParts {
			return errors.New("dns: SRV value must be weight port target")
		}
		var weight, port uint16
		if _, err := fmt.Sscan(parts[0], &weight); err != nil {
			return errors.New("dns: invalid SRV weight")
		}
		if _, err := fmt.Sscan(parts[1], &port); err != nil {
			return errors.New("dns: invalid SRV port")
		}
		if _, err := normalName(parts[2]); err != nil {
			return err
		}
	case "TXT":
	case "":
		return errors.New("dns: record type is required")
	default:
		return fmt.Errorf("dns: unsupported record type %q", x.Type)
	}
	return nil
}

func validateForwarder(x Forwarder) error {
	if x.ZoneName != "" {
		if _, e := normalName(x.ZoneName); e != nil {
			return e
		}
	}
	if _, _, e := net.SplitHostPort(x.Address); e != nil {
		return fmt.Errorf("dns: invalid upstream: %w", e)
	}
	return nil
}

func (s *Service) serveDNS(w mdns.ResponseWriter, req *mdns.Msg) {
	res := new(mdns.Msg)
	res.SetReply(req)
	if len(req.Question) != 1 {
		res.Rcode = mdns.RcodeFormatError
		s.write(w, res)
		return
	}
	q := req.Question[0]
	name := strings.TrimSuffix(strings.ToLower(q.Name), ".")
	s.mu.RLock()
	snap := s.active
	if blocked(name, snap.Blocks) {
		s.mu.RUnlock()
		res.Rcode = mdns.RcodeNameError
		s.write(w, res)
		return
	}
	if z, ok := findZone(name, snap.Zones); ok {
		s.answerZone(res, q, z, snap.Records)
		s.mu.RUnlock()
		s.write(w, res)
		return
	}
	key := fmt.Sprintf("%s/%d/%d", q.Name, q.Qtype, q.Qclass)
	if ent, ok := s.cache[key]; ok && time.Now().Before(ent.expires) {
		out := ent.message.Copy()
		decrementTTL(out, uint32(time.Since(ent.stored).Seconds()))
		s.mu.RUnlock()
		s.write(w, out)
		return
	}
	up := findForwarder(name, snap.Forwarders)
	s.mu.RUnlock()
	if up == "" {
		res.Rcode = mdns.RcodeServerFailure
		s.write(w, res)
		return
	}
	out, err := forward(req, up)
	if err != nil {
		res.Rcode = mdns.RcodeServerFailure
		s.write(w, res)
		return
	}
	s.putCache(key, out)
	s.write(w, out)
}

func (s *Service) write(w mdns.ResponseWriter, msg *mdns.Msg) {
	if err := w.WriteMsg(msg); err != nil {
		logx.Error("DNS service", "do", "write response", "err", err)
	}
}

func blocked(name string, bs []Block) bool {
	for _, b := range bs {
		if name == b.Name || strings.HasSuffix(name, "."+b.Name) {
			return true
		}
	}
	return false
}

func findZone(name string, zs []Zone) (Zone, bool) {
	var out Zone
	for _, z := range zs {
		if name == z.Name || strings.HasSuffix(name, "."+z.Name) {
			if len(z.Name) > len(out.Name) {
				out = z
			}
		}
	}
	return out, out.ID != 0
}

func findForwarder(name string, fs []Forwarder) string {
	var address, zone string
	for _, f := range fs {
		if f.ZoneName == "" {
			if address == "" {
				address = f.Address
			}
			continue
		}
		if (name == f.ZoneName || strings.HasSuffix(name, "."+f.ZoneName)) && len(f.ZoneName) > len(zone) {
			address, zone = f.Address, f.ZoneName
		}
	}
	return address
}

func (s *Service) answerZone(res *mdns.Msg, q mdns.Question, z Zone, rs []Record) {
	res.Authoritative = true
	found := false
	for _, r := range rs {
		if r.ZoneID == z.ID && strings.EqualFold(r.Name, strings.TrimSuffix(q.Name, ".")) {
			found = true
			if rr := toRR(r); rr != nil && (q.Qtype == mdns.TypeANY || q.Qtype == rr.Header().Rrtype) {
				res.Answer = append(res.Answer, rr)
			}
		}
	}
	if !found {
		res.Rcode = mdns.RcodeNameError
	}
	res.Ns = []mdns.RR{soa(z)}
}

func soa(z Zone) mdns.RR {
	return &mdns.SOA{Hdr: mdns.RR_Header{Name: mdns.Fqdn(z.Name), Rrtype: mdns.TypeSOA, Class: mdns.ClassINET, Ttl: z.TTL}, Ns: mdns.Fqdn(z.PrimaryNS), Mbox: mdns.Fqdn(strings.Replace(z.AdminEmail, "@", ".", 1)), Serial: z.Serial, Refresh: soaRefresh, Retry: soaRetry, Expire: soaExpire, Minttl: z.TTL}
}

func toRR(r Record) mdns.RR {
	h := mdns.RR_Header{Name: mdns.Fqdn(r.Name), Rrtype: map[string]uint16{"A": mdns.TypeA, "AAAA": mdns.TypeAAAA, "CNAME": mdns.TypeCNAME, "MX": mdns.TypeMX, "TXT": mdns.TypeTXT, "NS": mdns.TypeNS, "SRV": mdns.TypeSRV}[strings.ToUpper(r.Type)], Class: mdns.ClassINET, Ttl: r.TTL}
	switch strings.ToUpper(r.Type) {
	case "A":
		return &mdns.A{Hdr: h, A: net.ParseIP(r.Value)}
	case "AAAA":
		return &mdns.AAAA{Hdr: h, AAAA: net.ParseIP(r.Value)}
	case "CNAME":
		return &mdns.CNAME{Hdr: h, Target: mdns.Fqdn(r.Value)}
	case "NS":
		return &mdns.NS{Hdr: h, Ns: mdns.Fqdn(r.Value)}
	case "TXT":
		return &mdns.TXT{Hdr: h, Txt: []string{r.Value}}
	case "MX":
		return &mdns.MX{Hdr: h, Preference: r.Priority, Mx: mdns.Fqdn(r.Value)}
	case "SRV":
		parts := strings.Fields(r.Value)
		var weight, port uint16
		if _, err := fmt.Sscan(parts[0], &weight); err != nil {
			return nil
		}
		if _, err := fmt.Sscan(parts[1], &port); err != nil {
			return nil
		}
		return &mdns.SRV{Hdr: h, Priority: r.Priority, Weight: weight, Port: port, Target: mdns.Fqdn(parts[2])}
	}
	return nil
}

func forward(req *mdns.Msg, addr string) (*mdns.Msg, error) {
	c := &mdns.Client{Net: "udp", Timeout: upstreamTimeout}
	m, _, err := c.Exchange(req, addr)
	if err == nil && m.Truncated {
		c.Net = "tcp"
		m, _, err = c.Exchange(req, addr)
	}
	return m, err
}

func (s *Service) putCache(key string, m *mdns.Msg) {
	ttl := minTTL(m)
	if ttl == 0 {
		return
	}
	if limit := uint32(s.config.CacheMaxTTL.Seconds()); ttl > limit {
		ttl = limit
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cache) >= s.config.CacheEntries {
		for k := range s.cache {
			delete(s.cache, k)
			break
		}
	}
	now := time.Now()
	s.cache[key] = cacheEntry{message: m.Copy(), stored: now, expires: now.Add(time.Duration(ttl) * time.Second)}
}

func minTTL(m *mdns.Msg) uint32 {
	var n uint32
	for _, rr := range append(append(m.Answer, m.Ns...), m.Extra...) {
		if n == 0 || rr.Header().Ttl < n {
			n = rr.Header().Ttl
		}
	}
	return n
}

func decrementTTL(m *mdns.Msg, n uint32) {
	for _, rr := range append(append(m.Answer, m.Ns...), m.Extra...) {
		if rr.Header().Ttl > n {
			rr.Header().Ttl -= n
		} else {
			rr.Header().Ttl = 0
		}
	}
}
