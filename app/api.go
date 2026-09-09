/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package app provides Atlas transport adapters.
package app

import (
	"context"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"

	"go.arwos.org/atlas/app/transport"
	contract "go.arwos.org/atlas/app/types"
	"go.arwos.org/atlas/pkg/dhcp"
	"go.arwos.org/atlas/pkg/dns"
)

// API adapts domain services to the JSON-RPC transport.
type API struct {
	rpc  jsonrpc.Transport
	dhcp *dhcp.Service
	dns  *dns.Service
	apis []jsonrpc.TApi
}

// NewAPI constructs the JSON-RPC API adapter.
func NewAPI(rpc jsonrpc.Transport, dhcp *dhcp.Service, dns *dns.Service) *API {
	api := &API{
		rpc:  rpc,
		dhcp: dhcp,
		dns:  dns,
	}
	api.apis = []jsonrpc.TApi{
		transport.NewJSONRPCDHCPTransport(api, []string{"main"}),
		transport.NewJSONRPCDNSTransport(api, []string{"main"}),
	}
	return api
}

// Up registers the API with the JSON-RPC transport.
func (a *API) Up(ctx context.Context) error {
	for _, api := range a.apis {
		a.rpc.Add(api)
	}

	return nil
}

// Down releases API resources.
func (a *API) Down() error {
	return nil
}

var (
	_ contract.DHCP = (*API)(nil)
	_ contract.DNS  = (*API)(nil)
)
