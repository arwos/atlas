package app

import "time"

type DHCPDraftGetRequest struct{}
type DHCPActiveGetRequest struct{}
type DHCPApplyRequest struct{}
type DHCPLeaseListRequest struct{}

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

type DHCPSubnetDeleteRequest struct {
	ID int64 `json:"id"`
}
type DHCPReservationUpsertRequest struct {
	ID       int64  `json:"id,omitempty"`
	SubnetID int64  `json:"subnet_id"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
}
type DHCPReservationDeleteRequest struct {
	ID int64 `json:"id"`
}
type DHCPBlockUpsertRequest struct {
	ID  int64  `json:"id,omitempty"`
	MAC string `json:"mac"`
}
type DHCPBlockDeleteRequest struct {
	ID int64 `json:"id"`
}
type DHCPLeaseRevokeRequest struct {
	ID int64 `json:"id"`
}

type OperationResponse struct {
	Success bool `json:"success"`
}
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
type DHCPReservationResponse struct {
	ID       int64  `json:"id"`
	SubnetID int64  `json:"subnet_id"`
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
}
type DHCPBlockResponse struct {
	ID  int64  `json:"id"`
	MAC string `json:"mac"`
}
type DHCPLeaseResponse struct {
	ID        int64     `json:"id"`
	SubnetID  int64     `json:"subnet_id"`
	MAC       string    `json:"mac"`
	IP        string    `json:"ip"`
	ExpiresAt time.Time `json:"expires_at"`
}
type DHCPConfigurationResponse struct {
	Subnets      []DHCPSubnetResponse      `json:"subnets"`
	Reservations []DHCPReservationResponse `json:"reservations"`
	Blocks       []DHCPBlockResponse       `json:"blocks"`
}
