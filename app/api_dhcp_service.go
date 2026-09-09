/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"context"

	contract "go.arwos.org/atlas/app/types"
	"go.arwos.org/atlas/pkg/dhcp"
)

// DraftGet returns the editable DHCP configuration.
func (a *API) DraftGet(ctx context.Context) (contract.DHCPConfigurationResponse, error) {
	snapshot, err := a.dhcp.Draft(ctx)
	return toDHCPConfigurationResponse(snapshot), err
}

// ActiveGet returns the active DHCP configuration.
func (a *API) ActiveGet(ctx context.Context) (contract.DHCPConfigurationResponse, error) {
	snapshot, err := a.dhcp.Active(ctx)
	return toDHCPConfigurationResponse(snapshot), err
}

// SubnetUpsert creates or updates a draft DHCP subnet.
func (a *API) SubnetUpsert(ctx context.Context, subnet contract.DHCPSubnetUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.UpsertSubnet(ctx, toDHCPSubnet(subnet)))
}

// SubnetDelete deletes a draft DHCP subnet.
func (a *API) SubnetDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.DeleteSubnet(ctx, id))
}

// ReservationUpsert creates or updates a draft DHCP reservation.
func (a *API) ReservationUpsert(ctx context.Context, reservation contract.DHCPReservationUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.UpsertReservation(ctx, toDHCPReservation(reservation)))
}

// ReservationDelete deletes a draft DHCP reservation.
func (a *API) ReservationDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.DeleteReservation(ctx, id))
}

// BlockUpsert creates or updates a draft DHCP block.
func (a *API) BlockUpsert(ctx context.Context, block contract.DHCPBlockUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.UpsertBlock(ctx, toDHCPBlock(block)))
}

// BlockDelete deletes a draft DHCP block.
func (a *API) BlockDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.DeleteBlock(ctx, id))
}

// Apply activates the draft DHCP configuration.
func (a *API) Apply(ctx context.Context) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.Apply(ctx))
}

// LeaseList returns DHCP leases.
func (a *API) LeaseList(ctx context.Context) ([]contract.DHCPLeaseResponse, error) {
	leases, err := a.dhcp.Leases(ctx)
	return toDHCPLeaseResponses(leases), err
}

// LeaseRevoke removes a DHCP lease.
func (a *API) LeaseRevoke(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dhcp.RevokeLease(ctx, id))
}

func operationResponse(err error) (contract.OperationResponse, error) {
	if err != nil {
		return contract.OperationResponse{}, err
	}
	return contract.OperationResponse{Success: true}, nil
}

func toDHCPSubnet(v contract.DHCPSubnetUpsertRequest) dhcp.Subnet {
	return dhcp.Subnet{ID: v.ID, Interface: v.Interface, CIDR: v.CIDR, LeaseSeconds: v.LeaseSeconds, Router: v.Router, DNSServers: v.DNSServers, DomainSearch: v.DomainSearch, NTPServers: v.NTPServers, MTU: v.MTU, ClasslessRoutes: v.ClasslessRoutes}
}

func toDHCPReservation(v contract.DHCPReservationUpsertRequest) dhcp.Reservation {
	return dhcp.Reservation{ID: v.ID, SubnetID: v.SubnetID, MAC: v.MAC, IP: v.IP}
}

func toDHCPBlock(v contract.DHCPBlockUpsertRequest) dhcp.Block {
	return dhcp.Block{ID: v.ID, MAC: v.MAC}
}

func toDHCPConfigurationResponse(v dhcp.Snapshot) contract.DHCPConfigurationResponse {
	out := contract.DHCPConfigurationResponse{Subnets: make([]contract.DHCPSubnetResponse, 0, len(v.Subnets)), Reservations: make([]contract.DHCPReservationResponse, 0, len(v.Reservations)), Blocks: make([]contract.DHCPBlockResponse, 0, len(v.Blocks))}
	for _, subnet := range v.Subnets {
		out.Subnets = append(out.Subnets, contract.DHCPSubnetResponse{ID: subnet.ID, Interface: subnet.Interface, CIDR: subnet.CIDR, LeaseSeconds: subnet.LeaseSeconds, Router: subnet.Router, DNSServers: subnet.DNSServers, DomainSearch: subnet.DomainSearch, NTPServers: subnet.NTPServers, MTU: subnet.MTU, ClasslessRoutes: subnet.ClasslessRoutes})
	}
	for _, reservation := range v.Reservations {
		out.Reservations = append(out.Reservations, contract.DHCPReservationResponse{ID: reservation.ID, SubnetID: reservation.SubnetID, MAC: reservation.MAC, IP: reservation.IP})
	}
	for _, block := range v.Blocks {
		out.Blocks = append(out.Blocks, contract.DHCPBlockResponse{ID: block.ID, MAC: block.MAC})
	}
	return out
}

func toDHCPLeaseResponses(leases []dhcp.Lease) []contract.DHCPLeaseResponse {
	out := make([]contract.DHCPLeaseResponse, 0, len(leases))
	for _, lease := range leases {
		out = append(out, contract.DHCPLeaseResponse{ID: lease.ID, SubnetID: lease.SubnetID, MAC: lease.MAC, IP: lease.IP, ExpiresAt: lease.ExpiresAt})
	}
	return out
}
