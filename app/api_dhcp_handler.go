/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"context"
	"encoding/json"

	"go.osspkg.com/goppy/v3/plugins/web"

	"go.arwos.org/atlas/pkg/dhcp"
)

func (a *API) rpcDraft(ctx context.Context, _ web.Ctx, _ json.RawMessage) (any, error) {
	snapshot, err := a.dhcp.Draft(ctx)
	return toDHCPConfigurationResponse(snapshot), err
}

func (a *API) rpcActive(ctx context.Context, _ web.Ctx, _ json.RawMessage) (any, error) {
	snapshot, err := a.dhcp.Active(ctx)
	return toDHCPConfigurationResponse(snapshot), err
}

func (a *API) rpcSubnet(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPSubnetUpsertRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.UpsertSubnet(ctx, toDHCPSubnet(request)))
}

func (a *API) rpcSubnetDelete(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPSubnetDeleteRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.DeleteSubnet(ctx, request.ID))
}

func (a *API) rpcReservation(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPReservationUpsertRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.UpsertReservation(ctx, toDHCPReservation(request)))
}

func (a *API) rpcReservationDelete(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPReservationDeleteRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.DeleteReservation(ctx, request.ID))
}

func (a *API) rpcBlock(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPBlockUpsertRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.UpsertBlock(ctx, toDHCPBlock(request)))
}

func (a *API) rpcBlockDelete(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPBlockDeleteRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.DeleteBlock(ctx, request.ID))
}

func (a *API) rpcApply(ctx context.Context, _ web.Ctx, _ json.RawMessage) (any, error) {
	return operationResponse(a.dhcp.Apply(ctx))
}

func (a *API) rpcLeases(ctx context.Context, _ web.Ctx, _ json.RawMessage) (any, error) {
	leases, err := a.dhcp.Leases(ctx)
	return toDHCPLeaseResponses(leases), err
}

func (a *API) rpcLeaseRevoke(ctx context.Context, _ web.Ctx, p json.RawMessage) (any, error) {
	var request DHCPLeaseRevokeRequest
	if err := decode(p, &request); err != nil {
		return nil, err
	}
	return operationResponse(a.dhcp.RevokeLease(ctx, request.ID))
}

func operationResponse(err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return OperationResponse{Success: true}, nil
}

func toDHCPSubnet(v DHCPSubnetUpsertRequest) dhcp.Subnet {
	return dhcp.Subnet{ID: v.ID, Interface: v.Interface, CIDR: v.CIDR, LeaseSeconds: v.LeaseSeconds, Router: v.Router, DNSServers: v.DNSServers, DomainSearch: v.DomainSearch, NTPServers: v.NTPServers, MTU: v.MTU, ClasslessRoutes: v.ClasslessRoutes}
}

func toDHCPReservation(v DHCPReservationUpsertRequest) dhcp.Reservation {
	return dhcp.Reservation{ID: v.ID, SubnetID: v.SubnetID, MAC: v.MAC, IP: v.IP}
}

func toDHCPBlock(v DHCPBlockUpsertRequest) dhcp.Block {
	return dhcp.Block{ID: v.ID, MAC: v.MAC}
}

func toDHCPConfigurationResponse(v dhcp.Snapshot) DHCPConfigurationResponse {
	out := DHCPConfigurationResponse{Subnets: make([]DHCPSubnetResponse, 0, len(v.Subnets)), Reservations: make([]DHCPReservationResponse, 0, len(v.Reservations)), Blocks: make([]DHCPBlockResponse, 0, len(v.Blocks))}
	for _, x := range v.Subnets {
		out.Subnets = append(out.Subnets, DHCPSubnetResponse{ID: x.ID, Interface: x.Interface, CIDR: x.CIDR, LeaseSeconds: x.LeaseSeconds, Router: x.Router, DNSServers: x.DNSServers, DomainSearch: x.DomainSearch, NTPServers: x.NTPServers, MTU: x.MTU, ClasslessRoutes: x.ClasslessRoutes})
	}
	for _, x := range v.Reservations {
		out.Reservations = append(out.Reservations, DHCPReservationResponse{ID: x.ID, SubnetID: x.SubnetID, MAC: x.MAC, IP: x.IP})
	}
	for _, x := range v.Blocks {
		out.Blocks = append(out.Blocks, DHCPBlockResponse{ID: x.ID, MAC: x.MAC})
	}
	return out
}

func toDHCPLeaseResponses(v []dhcp.Lease) []DHCPLeaseResponse {
	out := make([]DHCPLeaseResponse, 0, len(v))
	for _, x := range v {
		out = append(out, DHCPLeaseResponse{ID: x.ID, SubnetID: x.SubnetID, MAC: x.MAC, IP: x.IP, ExpiresAt: x.ExpiresAt})
	}
	return out
}
