//nolint:revive // JSON-RPC methods mirror generated contract names.
package app

import (
	"context"

	contract "go.arwos.org/atlas/app/types"
	"go.arwos.org/atlas/pkg/dns"
)

func (a *API) DraftGetDNS(ctx context.Context) (contract.DNSConfigurationResponse, error) {
	x, e := a.dns.Draft(ctx)
	return dnsResponse(x), e
}

func (a *API) ActiveGetDNS(ctx context.Context) (contract.DNSConfigurationResponse, error) {
	x, e := a.dns.Active(ctx)
	return dnsResponse(x), e
}

func (a *API) ZoneUpsert(ctx context.Context, x contract.DNSZoneUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dns.UpsertZone(ctx, dns.Zone{ID: x.ID, Name: x.Name, PrimaryNS: x.PrimaryNS, AdminEmail: x.AdminEmail, TTL: ttl(x.TTL)}))
}

func (a *API) ZoneDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dns.DeleteZone(ctx, id))
}

func (a *API) RecordUpsert(ctx context.Context, x contract.DNSRecordUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dns.UpsertRecord(ctx, dns.Record{ID: x.ID, ZoneID: x.ZoneID, Name: x.Name, Type: x.Type, Value: x.Value, TTL: ttl(x.TTL), Priority: x.Priority}))
}

func (a *API) RecordDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dns.DeleteRecord(ctx, id))
}

func (a *API) ForwarderUpsert(ctx context.Context, x contract.DNSForwarderUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dns.UpsertForwarder(ctx, dns.Forwarder{ID: x.ID, ZoneName: x.ZoneName, Address: x.Address}))
}

func (a *API) ForwarderDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dns.DeleteForwarder(ctx, id))
}

func (a *API) DNSBlockUpsert(ctx context.Context, x contract.DNSBlockUpsertRequest) (contract.OperationResponse, error) {
	return operationResponse(a.dns.UpsertBlock(ctx, dns.Block{ID: x.ID, Name: x.Name}))
}

func (a *API) DNSBlockDelete(ctx context.Context, id int64) (contract.OperationResponse, error) {
	return operationResponse(a.dns.DeleteBlock(ctx, id))
}

func (a *API) ApplyDNS(ctx context.Context) (contract.OperationResponse, error) {
	return operationResponse(a.dns.Apply(ctx))
}

func ttl(v uint32) uint32 {
	if v == 0 {
		return defaultRecordTTL
	}
	return v
}

const defaultRecordTTL = 3600

func dnsResponse(x dns.Snapshot) contract.DNSConfigurationResponse {
	out := contract.DNSConfigurationResponse{Zones: make([]contract.DNSZoneResponse, 0, len(x.Zones)), Records: make([]contract.DNSRecordResponse, 0, len(x.Records)), Forwarders: make([]contract.DNSForwarderResponse, 0, len(x.Forwarders)), Blocks: make([]contract.DNSBlockResponse, 0, len(x.Blocks))}
	for _, v := range x.Zones {
		out.Zones = append(out.Zones, contract.DNSZoneResponse{ID: v.ID, Name: v.Name, PrimaryNS: v.PrimaryNS, AdminEmail: v.AdminEmail, TTL: v.TTL, Serial: v.Serial})
	}
	for _, v := range x.Records {
		out.Records = append(out.Records, contract.DNSRecordResponse{ID: v.ID, ZoneID: v.ZoneID, Name: v.Name, Type: v.Type, Value: v.Value, TTL: v.TTL, Priority: v.Priority})
	}
	for _, v := range x.Forwarders {
		out.Forwarders = append(out.Forwarders, contract.DNSForwarderResponse{ID: v.ID, ZoneName: v.ZoneName, Address: v.Address})
	}
	for _, v := range x.Blocks {
		out.Blocks = append(out.Blocks, contract.DNSBlockResponse{ID: v.ID, Name: v.Name})
	}
	return out
}

var _ contract.DNS = (*API)(nil)
