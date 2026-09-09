CREATE TABLE IF NOT EXISTS dns_zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    name TEXT NOT NULL,
    primary_ns TEXT NOT NULL,
    admin_email TEXT NOT NULL,
    ttl INTEGER NOT NULL DEFAULT 3600,
    serial INTEGER NOT NULL DEFAULT 1,
    UNIQUE(version, name)
);
CREATE TABLE IF NOT EXISTS dns_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    zone_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    ttl INTEGER NOT NULL DEFAULT 3600,
    priority INTEGER NOT NULL DEFAULT 0,
    UNIQUE(version, zone_id, name, type, value, priority)
);
CREATE TABLE IF NOT EXISTS dns_forwarders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    zone_name TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL,
    UNIQUE(version, zone_name)
);
CREATE TABLE IF NOT EXISTS dns_blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL CHECK(version IN ('draft', 'active')),
    name TEXT NOT NULL,
    UNIQUE(version, name)
);
