package app

import (
	"context"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"

	"go.arwos.org/atlas/pkg/dhcp"
)

type API struct {
	rpc  jsonrpc.Transport
	dhcp *dhcp.Service
}

func NewAPI(rpc jsonrpc.Transport, dhcp *dhcp.Service) *API {
	return &API{
		rpc:  rpc,
		dhcp: dhcp,
	}
}

func (a *API) Up(ctx context.Context) error {
	a.rpc.Add(a)

	return nil
}

func (a *API) Down() error {
	return nil
}

func (a *API) RouteTags() []string {
	return []string{"main"}
}

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
