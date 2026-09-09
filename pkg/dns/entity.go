// Package dns manages authoritative and forwarding DNS configuration.
//
//nolint:revive // Domain entities are intentionally compact data carriers.
package dns

const (
	draftVersion  = "draft"
	activeVersion = "active"
)

type Zone struct {
	ID                          int64
	Name, PrimaryNS, AdminEmail string
	TTL                         uint32
	Serial                      uint32
}
type Record struct {
	ID, ZoneID        int64
	Name, Type, Value string
	TTL               uint32
	Priority          uint16
}
type Forwarder struct {
	ID                int64
	ZoneName, Address string
}
type Block struct {
	ID   int64
	Name string
}
type Snapshot struct {
	Zones      []Zone
	Records    []Record
	Forwarders []Forwarder
	Blocks     []Block
}
