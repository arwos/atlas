/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import "time"

type (
	// DHCPDraftGetRequest requests the editable DHCP configuration.
	DHCPDraftGetRequest struct{}
	// DHCPActiveGetRequest requests the active DHCP configuration.
	DHCPActiveGetRequest struct{}
	// DHCPApplyRequest applies the editable DHCP configuration.
	DHCPApplyRequest struct{}
	// DHCPLeaseListRequest requests DHCP leases.
	DHCPLeaseListRequest struct{}
)

// DHCPSubnetUpsertRequest creates or updates a DHCP subnet.
type DHCPSubnetUpsertRequest struct {
	ID              int64    `json:"id,omitempty"`
	Interface       string   `json:"interface"`
	CIDR            string   `json:"cidr"`
	LeaseSeconds    int64    `json:"lease_seconds"`
	Router          string   `json:"router,omitempty"`
	DNSServers      []string `json:"dns_servers,omitempty"`
	DomainSearch    string   `json:"domain_search,omitempty"`
	NTPServers      []string `json:"ntp_servers,omitempty"`
	MTU             int      `json:"mtu,omitempty"`
	ClasslessRoutes []string `json:"classless_routes,omitempty"`
}

// DHCPSubnetDeleteRequest deletes a DHCP subnet.
type DHCPSubnetDeleteRequest struct {
	ID int64 `json:"id"`
}

// DHCPReservationUpsertRequest creates or updates a DHCP reservation.
type DHCPReservationUpsertRequest struct {
	ID       int64  `json:"id,omitempty"`
	SubnetID int64  `json:"subnet_id"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
}

// DHCPReservationDeleteRequest deletes a DHCP reservation.
type DHCPReservationDeleteRequest struct {
	ID int64 `json:"id"`
}

// DHCPBlockUpsertRequest creates or updates a DHCP block.
type DHCPBlockUpsertRequest struct {
	ID  int64  `json:"id,omitempty"`
	MAC string `json:"mac"`
}

// DHCPBlockDeleteRequest deletes a DHCP block.
type DHCPBlockDeleteRequest struct {
	ID int64 `json:"id"`
}

// DHCPLeaseRevokeRequest revokes a DHCP lease.
type DHCPLeaseRevokeRequest struct {
	ID int64 `json:"id"`
}

// OperationResponse reports whether a requested operation succeeded.
type OperationResponse struct {
	Success bool `json:"success"`
}

// DHCPSubnetResponse represents a DHCP subnet over JSON-RPC.
type DHCPSubnetResponse struct {
	ID              int64    `json:"id"`
	Interface       string   `json:"interface"`
	CIDR            string   `json:"cidr"`
	LeaseSeconds    int64    `json:"lease_seconds"`
	Router          string   `json:"router,omitempty"`
	DNSServers      []string `json:"dns_servers,omitempty"`
	DomainSearch    string   `json:"domain_search,omitempty"`
	NTPServers      []string `json:"ntp_servers,omitempty"`
	MTU             int      `json:"mtu,omitempty"`
	ClasslessRoutes []string `json:"classless_routes,omitempty"`
}

// DHCPReservationResponse represents a DHCP reservation over JSON-RPC.
type DHCPReservationResponse struct {
	ID       int64  `json:"id"`
	SubnetID int64  `json:"subnet_id"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
}

// DHCPBlockResponse represents a DHCP block over JSON-RPC.
type DHCPBlockResponse struct {
	ID  int64  `json:"id"`
	MAC string `json:"mac"`
}

// DHCPLeaseResponse represents a DHCP lease over JSON-RPC.
type DHCPLeaseResponse struct {
	ID        int64     `json:"id"`
	SubnetID  int64     `json:"subnet_id"`
	MAC       string    `json:"mac"`
	IP        string    `json:"ip"`
	ExpiresAt time.Time `json:"expires_at"`
}

// DHCPConfigurationResponse represents a complete DHCP configuration over JSON-RPC.
type DHCPConfigurationResponse struct {
	Subnets      []DHCPSubnetResponse      `json:"subnets"`
	Reservations []DHCPReservationResponse `json:"reservations"`
	Blocks       []DHCPBlockResponse       `json:"blocks"`
}
