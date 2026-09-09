/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package app provides Atlas transport adapters.
package app

import (
	"context"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"

	"go.arwos.org/atlas/pkg/dhcp"
)

// API adapts domain services to the JSON-RPC transport.
type API struct {
	rpc  jsonrpc.Transport
	dhcp *dhcp.Service
}

// NewAPI constructs the JSON-RPC API adapter.
func NewAPI(rpc jsonrpc.Transport, dhcp *dhcp.Service) *API {
	return &API{
		rpc:  rpc,
		dhcp: dhcp,
	}
}

// Up registers the API with the JSON-RPC transport.
func (a *API) Up(ctx context.Context) error {
	a.rpc.Add(a)

	return nil
}

// Down releases API resources.
func (a *API) Down() error {
	return nil
}

// RouteTags returns the transport route tags served by the API.
func (a *API) RouteTags() []string {
	return []string{"main"}
}

// JSONRPCApiHandlers returns the JSON-RPC method handlers.
func (a *API) JSONRPCApiHandlers() map[string]jsonrpc.THandleFunc {
	return map[string]jsonrpc.THandleFunc{
		"dhcp.draft.get":          a.rpcDraft,
		"dhcp.active.get":         a.rpcActive,
		"dhcp.subnet.upsert":      a.rpcSubnet,
		"dhcp.subnet.delete":      a.rpcSubnetDelete,
		"dhcp.reservation.upsert": a.rpcReservation,
		"dhcp.reservation.delete": a.rpcReservationDelete,
		"dhcp.block.upsert":       a.rpcBlock,
		"dhcp.block.delete":       a.rpcBlockDelete,
		"dhcp.apply":              a.rpcApply,
		"dhcp.lease.list":         a.rpcLeases,
		"dhcp.lease.revoke":       a.rpcLeaseRevoke,
	}
}
