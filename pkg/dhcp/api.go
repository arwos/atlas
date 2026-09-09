/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"context"
	"encoding/json"

	"go.osspkg.com/goppy/v3/plugins/web"
	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"
)

func (s *Service) RouteTags() []string { return []string{"main"} }

func (s *Service) JSONRPCApiHandlers() map[string]jsonrpc.THandleFunc {
	return map[string]jsonrpc.THandleFunc{
		"dhcp.draft.get":          s.rpcDraft,
		"dhcp.active.get":         s.rpcActive,
		"dhcp.subnet.upsert":      s.rpcSubnet,
		"dhcp.subnet.delete":      s.rpcSubnetDelete,
		"dhcp.reservation.upsert": s.rpcReservation,
		"dhcp.reservation.delete": s.rpcReservationDelete,
		"dhcp.block.upsert":       s.rpcBlock,
		"dhcp.block.delete":       s.rpcBlockDelete,
		"dhcp.apply":              s.rpcApply,
		"dhcp.lease.list":         s.rpcLeases,
		"dhcp.lease.revoke":       s.rpcLeaseRevoke,
	}
}

func decode[T any](p json.RawMessage, out *T) error {
	if err := json.Unmarshal(p, out); err != nil {
		return err
	}
	return nil
}

func (s *Service) rpcDraft(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	return s.repo.snapshot(ctx, draftVersion)
}

func (s *Service) rpcActive(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	return s.repo.snapshot(ctx, activeVersion)
}

func (s *Service) rpcSubnet(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x Subnet
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	if e := s.validate(Snapshot{Subnets: []Subnet{x}}); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.repo.upsertSubnet(ctx, x)
}

func (s *Service) rpcSubnetDelete(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x struct {
		ID int64 `json:"id"`
	}
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.repo.deleteSubnet(ctx, x.ID)
}

func (s *Service) rpcReservation(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x Reservation
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	m, e := normalMAC(x.MAC)
	if e != nil {
		return nil, e
	}
	x.MAC = m
	return map[string]bool{"ok": true}, s.repo.upsertReservation(ctx, x)
}

func (s *Service) rpcReservationDelete(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x struct {
		ID int64 `json:"id"`
	}
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.repo.deleteReservation(ctx, x.ID)
}

func (s *Service) rpcBlock(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x Block
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	m, e := normalMAC(x.MAC)
	if e != nil {
		return nil, e
	}
	x.MAC = m
	return map[string]bool{"ok": true}, s.repo.upsertBlock(ctx, x)
}

func (s *Service) rpcBlockDelete(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x struct {
		ID int64 `json:"id"`
	}
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.repo.deleteBlock(ctx, x.ID)
}

func (s *Service) rpcApply(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	snap, e := s.repo.snapshot(ctx, draftVersion)
	if e != nil {
		return nil, e
	}
	if e = s.validate(snap); e != nil {
		return nil, e
	}
	if e = s.repo.replaceActive(ctx); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.reload(ctx)
}

func (s *Service) rpcLeases(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	return s.repo.leases(ctx)
}

func (s *Service) rpcLeaseRevoke(ctx context.Context, w web.Ctx, p json.RawMessage) (any, error) {
	var x struct {
		ID int64 `json:"id"`
	}
	if e := decode(p, &x); e != nil {
		return nil, e
	}
	return map[string]bool{"ok": true}, s.repo.revokeLease(ctx, x.ID)
}
