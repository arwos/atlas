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

// Subnet is a DHCP subnet configuration.
type Subnet struct {
	ID              int64
	Interface       string
	CIDR            string
	LeaseSeconds    int64
	Router          string
	DNSServers      []string
	DomainSearch    string
	NTPServers      []string
	MTU             int
	ClasslessRoutes []string
}

// Reservation assigns an address to a hardware address.
type Reservation struct {
	ID       int64
	SubnetID int64
	MAC      string
	IP       string
}

// Block denies DHCP service to a hardware address.
type Block struct {
	ID  int64
	MAC string
}

// Lease is an active DHCP address assignment.
type Lease struct {
	ID        int64
	SubnetID  int64
	MAC       string
	IP        string
	ExpiresAt time.Time
}

// Snapshot is a complete DHCP configuration revision.
type Snapshot struct {
	Subnets      []Subnet
	Reservations []Reservation
	Blocks       []Block
}
