/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"time"
)

const (
	draftVersion  = "draft"
	activeVersion = "active"
)

type Subnet struct {
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

type Reservation struct {
	ID       int64  `json:"id"`
	SubnetID int64  `json:"subnet_id"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
}

type Block struct {
	ID  int64  `json:"id"`
	MAC string `json:"mac"`
}

type Lease struct {
	ID        int64     `json:"id"`
	SubnetID  int64     `json:"subnet_id"`
	MAC       string    `json:"mac"`
	IP        string    `json:"ip"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Snapshot struct {
	Subnets      []Subnet      `json:"subnets"`
	Reservations []Reservation `json:"reservations"`
	Blocks       []Block       `json:"blocks"`
}
