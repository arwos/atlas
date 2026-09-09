CREATE TABLE IF NOT EXISTS dhcp_subnets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    interface_name TEXT NOT NULL,
    cidr TEXT NOT NULL,
    lease_seconds INTEGER NOT NULL,
    router TEXT NOT NULL DEFAULT '',
    dns_servers TEXT NOT NULL DEFAULT '[]',
    domain_search TEXT NOT NULL DEFAULT '',
    ntp_servers TEXT NOT NULL DEFAULT '[]',
    mtu INTEGER NOT NULL DEFAULT 0,
    classless_routes TEXT NOT NULL DEFAULT '[]',
    UNIQUE(version, interface_name, cidr)
);
CREATE TABLE IF NOT EXISTS dhcp_reservations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    subnet_id INTEGER NOT NULL,
    mac TEXT NOT NULL,
    ip TEXT NOT NULL,
    UNIQUE(version, mac),
    UNIQUE(version, ip)
);
CREATE TABLE IF NOT EXISTS dhcp_blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    mac TEXT NOT NULL,
    UNIQUE(version, mac)
);
CREATE TABLE IF NOT EXISTS dhcp_leases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subnet_id INTEGER NOT NULL,
    mac TEXT NOT NULL,
    ip TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    UNIQUE(subnet_id, mac),
    UNIQUE(subnet_id, ip)
);
CREATE TABLE IF NOT EXISTS dhcp_conflicts (
    subnet_id INTEGER NOT NULL,
    ip TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    PRIMARY KEY(subnet_id, ip)
);
