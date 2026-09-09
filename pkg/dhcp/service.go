/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"

	"go.arwos.org/atlas/pkg/database"
)

// Service manages DHCP configuration and its active lifecycle.
type Service struct {
	repo   *repository
	config *Config
	mu     sync.RWMutex
	active Snapshot
	conn   *net.UDPConn
}

// NewService constructs a DHCP service.
func NewService(db *database.Service, cfg *ConfigGroup) *Service {
	return &Service{
		repo: &repository{
			db: db,
		},
		config: &cfg.Config,
	}
}

// Up initializes DHCP state from the active configuration.
func (s *Service) Up(ctx context.Context) error {
	if err := s.repo.bootstrap(ctx, s.config.Bootstrap); err != nil {
		return err
	}
	return s.reload(ctx)
}

// Down closes DHCP resources.
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

// Draft returns the editable DHCP configuration.
func (s *Service) Draft(ctx context.Context) (Snapshot, error) {
	return s.repo.snapshot(ctx, draftVersion)
}

// Active returns the configuration currently used by the DHCP service.
func (s *Service) Active(ctx context.Context) (Snapshot, error) {
	return s.repo.snapshot(ctx, activeVersion)
}

// UpsertSubnet creates or updates a draft subnet.
func (s *Service) UpsertSubnet(ctx context.Context, subnet Subnet) error {
	if err := s.validate(Snapshot{Subnets: []Subnet{subnet}}); err != nil {
		return err
	}
	return s.repo.upsertSubnet(ctx, subnet)
}

// DeleteSubnet deletes a draft subnet.
func (s *Service) DeleteSubnet(ctx context.Context, id int64) error {
	return s.repo.deleteSubnet(ctx, id)
}

// UpsertReservation creates or updates a draft reservation.
func (s *Service) UpsertReservation(ctx context.Context, reservation Reservation) error {
	mac, err := normalMAC(reservation.MAC)
	if err != nil {
		return err
	}
	reservation.MAC = mac
	return s.repo.upsertReservation(ctx, reservation)
}

// DeleteReservation deletes a draft reservation.
func (s *Service) DeleteReservation(ctx context.Context, id int64) error {
	return s.repo.deleteReservation(ctx, id)
}

// UpsertBlock creates or updates a draft block.
func (s *Service) UpsertBlock(ctx context.Context, block Block) error {
	mac, err := normalMAC(block.MAC)
	if err != nil {
		return err
	}
	block.MAC = mac
	return s.repo.upsertBlock(ctx, block)
}

// DeleteBlock deletes a draft block.
func (s *Service) DeleteBlock(ctx context.Context, id int64) error {
	return s.repo.deleteBlock(ctx, id)
}

// Apply validates and activates the draft configuration.
func (s *Service) Apply(ctx context.Context) error {
	snapshot, err := s.Draft(ctx)
	if err != nil {
		return err
	}
	if err = s.validate(snapshot); err != nil {
		return err
	}
	if err = s.repo.replaceActive(ctx); err != nil {
		return err
	}
	return s.reload(ctx)
}

// Leases returns active DHCP leases.
func (s *Service) Leases(ctx context.Context) ([]Lease, error) {
	return s.repo.leases(ctx)
}

// RevokeLease removes a DHCP lease.
func (s *Service) RevokeLease(ctx context.Context, id int64) error {
	return s.repo.revokeLease(ctx, id)
}

func (s *Service) validate(snap Snapshot) error {
	byID, err := validateSubnets(snap.Subnets)
	if err != nil {
		return err
	}
	if err = validateReservations(snap.Reservations, byID); err != nil {
		return err
	}
	return validateBlocks(snap.Blocks)
}

func validateSubnets(subnets []Subnet) (map[int64]netip.Prefix, error) {
	seen := map[string]netip.Prefix{}
	byID := make(map[int64]netip.Prefix, len(subnets))
	for _, sub := range subnets {
		p, err := netip.ParsePrefix(sub.CIDR)
		if err != nil || !p.Addr().Is4() {
			return nil, fmt.Errorf("dhcp: invalid IPv4 CIDR %q", sub.CIDR)
		}
		if sub.Interface == "" || sub.LeaseSeconds <= 0 {
			return nil, errors.New("dhcp: interface and positive lease_seconds are required")
		}
		for _, other := range seen {
			if other.Overlaps(p) {
				return nil, errors.New("dhcp: overlapping subnets")
			}
		}
		seen[sub.CIDR] = p
		byID[sub.ID] = p
	}
	return byID, nil
}

func validateReservations(reservations []Reservation, byID map[int64]netip.Prefix) error {
	macs := map[string]bool{}
	ips := map[string]bool{}
	for _, r := range reservations {
		m, err := normalMAC(r.MAC)
		if err != nil {
			return err
		}
		if macs[m] || ips[r.IP] {
			return errors.New("dhcp: duplicate reservation")
		}
		macs[m] = true
		ips[r.IP] = true
		p, ok := byID[r.SubnetID]
		ip, err := netip.ParseAddr(r.IP)
		if !ok || err != nil || !p.Contains(ip) {
			return errors.New("dhcp: reservation IP is outside subnet")
		}
	}
	return nil
}

func validateBlocks(blocks []Block) error {
	for _, b := range blocks {
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
