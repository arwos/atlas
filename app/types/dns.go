//nolint:revive // Transport DTO names intentionally mirror JSON-RPC contract names.
package contract

import "context"

type DNS interface {
	DraftGetDNS(ctx context.Context) (configuration DNSConfigurationResponse, err error)
	ActiveGetDNS(ctx context.Context) (configuration DNSConfigurationResponse, err error)
	ZoneUpsert(ctx context.Context, zone DNSZoneUpsertRequest) (operation OperationResponse, err error)
	ZoneDelete(ctx context.Context, id int64) (operation OperationResponse, err error)
	RecordUpsert(ctx context.Context, record DNSRecordUpsertRequest) (operation OperationResponse, err error)
	RecordDelete(ctx context.Context, id int64) (operation OperationResponse, err error)
	ForwarderUpsert(ctx context.Context, forwarder DNSForwarderUpsertRequest) (operation OperationResponse, err error)
	ForwarderDelete(ctx context.Context, id int64) (operation OperationResponse, err error)
	DNSBlockUpsert(ctx context.Context, block DNSBlockUpsertRequest) (operation OperationResponse, err error)
	DNSBlockDelete(ctx context.Context, id int64) (operation OperationResponse, err error)
	ApplyDNS(ctx context.Context) (operation OperationResponse, err error)
}
type DNSZoneUpsertRequest struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name"`
	PrimaryNS  string `json:"primary_ns"`
	AdminEmail string `json:"admin_email"`
	TTL        uint32 `json:"ttl,omitempty"`
}
type DNSRecordUpsertRequest struct {
	ID       int64  `json:"id,omitempty"`
	ZoneID   int64  `json:"zone_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl,omitempty"`
	Priority uint16 `json:"priority,omitempty"`
}
type DNSForwarderUpsertRequest struct {
	ID       int64  `json:"id,omitempty"`
	ZoneName string `json:"zone_name,omitempty"`
	Address  string `json:"address"`
}
type DNSBlockUpsertRequest struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}
type DNSZoneResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PrimaryNS  string `json:"primary_ns"`
	AdminEmail string `json:"admin_email"`
	TTL        uint32 `json:"ttl"`
	Serial     uint32 `json:"serial"`
}
type DNSRecordResponse struct {
	ID       int64  `json:"id"`
	ZoneID   int64  `json:"zone_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl"`
	Priority uint16 `json:"priority,omitempty"`
}
type DNSForwarderResponse struct {
	ID       int64  `json:"id"`
	ZoneName string `json:"zone_name,omitempty"`
	Address  string `json:"address"`
}
type DNSBlockResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type DNSConfigurationResponse struct {
	Zones      []DNSZoneResponse      `json:"zones"`
	Records    []DNSRecordResponse    `json:"records"`
	Forwarders []DNSForwarderResponse `json:"forwarders"`
	Blocks     []DNSBlockResponse     `json:"blocks"`
}
