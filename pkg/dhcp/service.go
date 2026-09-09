/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"

	"go.arwos.org/atlas/pkg/database"
)

type Service struct {
	repo   *repository
	config *Config
	mu     sync.RWMutex
	active Snapshot
	conn   *net.UDPConn
}

func NewService(db *database.Service, cfg *ConfigGroup, rpc jsonrpc.Transport) *Service {
	s := &Service{
		repo: &repository{
			db: db,
		},
		config: &cfg.Config,
	}
	rpc.Add(s)
	return s
}

func (s *Service) Up(ctx context.Context) error {
	if err := s.repo.bootstrap(ctx, s.config.Bootstrap); err != nil {
		return err
	}
	return s.reload(ctx)
}

func (s *Service) Down() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
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
	s.mu.Unlock()
	// Binding UDP/67 is deliberately delayed until an active interface has a usable IPv4.
	// The packet engine is independent from the socket and can be exercised without privileges.
	return nil
}

func (s *Service) validate(snap Snapshot) error {
	seen := map[string]netip.Prefix{}
	for _, sub := range snap.Subnets {
		p, err := netip.ParsePrefix(sub.CIDR)
		if err != nil || !p.Addr().Is4() {
			return fmt.Errorf("dhcp: invalid IPv4 CIDR %q", sub.CIDR)
		}
		if sub.Interface == "" || sub.LeaseSeconds <= 0 {
			return fmt.Errorf("dhcp: interface and positive lease_seconds are required")
		}
		for _, other := range seen {
			if other.Overlaps(p) {
				return fmt.Errorf("dhcp: overlapping subnets")
			}
		}
		seen[sub.CIDR] = p
	}
	byID := map[int64]netip.Prefix{}
	for _, sub := range snap.Subnets {
		p, _ := netip.ParsePrefix(sub.CIDR)
		byID[sub.ID] = p
	}
	macs := map[string]bool{}
	ips := map[string]bool{}
	for _, r := range snap.Reservations {
		m, err := normalMAC(r.MAC)
		if err != nil {
			return err
		}
		if macs[m] || ips[r.IP] {
			return fmt.Errorf("dhcp: duplicate reservation")
		}
		macs[m] = true
		ips[r.IP] = true
		p, ok := byID[r.SubnetID]
		ip, err := netip.ParseAddr(r.IP)
		if !ok || err != nil || !p.Contains(ip) {
			return fmt.Errorf("dhcp: reservation IP is outside subnet")
		}
	}
	for _, b := range snap.Blocks {
		if _, err := normalMAC(b.MAC); err != nil {
			return err
		}
	}
	return nil
}
func normalMAC(in string) (string, error) {
	m, e := net.ParseMAC(in)
	if e != nil || len(m) != 6 {
		return "", fmt.Errorf("dhcp: invalid MAC %q", in)
	}
	return strings.ToLower(m.String()), nil
}
