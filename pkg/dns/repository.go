package dns

import (
	"context"
	"database/sql"

	"go.osspkg.com/goppy/v3/plugins/orm"
	"go.osspkg.com/logx"

	"go.arwos.org/atlas/pkg/database"
)

type repository struct{ db *database.Service }

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logx.Error("DNS repository", "do", "close rows", "err", err)
	}
}

func (r *repository) exec(ctx context.Context, name, q string, args ...any) error {
	return r.db.Master().CallContext(ctx, name, func(ctx context.Context, db orm.DB) error { _, err := db.ExecContext(ctx, q, args...); return err })
}

func (r *repository) snapshot(ctx context.Context, version string) (Snapshot, error) {
	var out Snapshot
	err := r.db.Master().CallContext(ctx, "dns.snapshot", func(ctx context.Context, db orm.DB) error {
		var err error
		if out.Zones, err = queryZones(ctx, db, version); err != nil {
			return err
		}
		if out.Records, err = queryRecords(ctx, db, version); err != nil {
			return err
		}
		if out.Forwarders, err = queryForwarders(ctx, db, version); err != nil {
			return err
		}
		out.Blocks, err = queryBlocks(ctx, db, version)
		return err
	})
	return out, err
}

func queryZones(ctx context.Context, db orm.DB, v string) ([]Zone, error) {
	rows, e := db.QueryContext(ctx, "SELECT id,name,primary_ns,admin_email,ttl,serial FROM dns_zones WHERE version=? ORDER BY id", v)
	if e != nil {
		return nil, e
	}
	defer closeRows(rows)
	var out []Zone
	for rows.Next() {
		var x Zone
		if e = rows.Scan(&x.ID, &x.Name, &x.PrimaryNS, &x.AdminEmail, &x.TTL, &x.Serial); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func queryRecords(ctx context.Context, db orm.DB, v string) ([]Record, error) {
	rows, e := db.QueryContext(ctx, "SELECT id,zone_id,name,type,value,ttl,priority FROM dns_records WHERE version=? ORDER BY id", v)
	if e != nil {
		return nil, e
	}
	defer closeRows(rows)
	var out []Record
	for rows.Next() {
		var x Record
		if e = rows.Scan(&x.ID, &x.ZoneID, &x.Name, &x.Type, &x.Value, &x.TTL, &x.Priority); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func queryForwarders(ctx context.Context, db orm.DB, v string) ([]Forwarder, error) {
	rows, e := db.QueryContext(ctx, "SELECT id,zone_name,address FROM dns_forwarders WHERE version=? ORDER BY id", v)
	if e != nil {
		return nil, e
	}
	defer closeRows(rows)
	var out []Forwarder
	for rows.Next() {
		var x Forwarder
		if e = rows.Scan(&x.ID, &x.ZoneName, &x.Address); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func queryBlocks(ctx context.Context, db orm.DB, v string) ([]Block, error) {
	rows, e := db.QueryContext(ctx, "SELECT id,name FROM dns_blocks WHERE version=? ORDER BY id", v)
	if e != nil {
		return nil, e
	}
	defer closeRows(rows)
	var out []Block
	for rows.Next() {
		var x Block
		if e = rows.Scan(&x.ID, &x.Name); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *repository) replaceActive(ctx context.Context) error {
	return r.db.Master().TxContext(ctx, "dns.apply", func(ctx context.Context, db orm.DB) error {
		for _, q := range []string{"DELETE FROM dns_records WHERE version='active'", "DELETE FROM dns_zones WHERE version='active'", "DELETE FROM dns_forwarders WHERE version='active'", "DELETE FROM dns_blocks WHERE version='active'"} {
			if _, e := db.ExecContext(ctx, q); e != nil {
				return e
			}
		}
		for _, q := range []string{"INSERT INTO dns_zones(version,name,primary_ns,admin_email,ttl,serial) SELECT 'active',name,primary_ns,admin_email,ttl,serial FROM dns_zones WHERE version='draft'", "INSERT INTO dns_records(version,zone_id,name,type,value,ttl,priority) SELECT 'active',a.id,r.name,r.type,r.value,r.ttl,r.priority FROM dns_records r JOIN dns_zones d ON d.id=r.zone_id AND d.version='draft' JOIN dns_zones a ON a.version='active' AND a.name=d.name WHERE r.version='draft'", "INSERT INTO dns_forwarders(version,zone_name,address) SELECT 'active',zone_name,address FROM dns_forwarders WHERE version='draft'", "INSERT INTO dns_blocks(version,name) SELECT 'active',name FROM dns_blocks WHERE version='draft'"} {
			if _, e := db.ExecContext(ctx, q); e != nil {
				return e
			}
		}
		return nil
	})
}

func (r *repository) upsertZone(ctx context.Context, x Zone) error {
	if x.ID == 0 {
		return r.exec(ctx, "dns.zone.create", "INSERT INTO dns_zones(version,name,primary_ns,admin_email,ttl,serial) VALUES('draft',?,?,?,?,?)", x.Name, x.PrimaryNS, x.AdminEmail, x.TTL, x.Serial)
	}
	return r.exec(ctx, "dns.zone.update", "UPDATE dns_zones SET name=?,primary_ns=?,admin_email=?,ttl=?,serial=serial+1 WHERE id=? AND version='draft'", x.Name, x.PrimaryNS, x.AdminEmail, x.TTL, x.ID)
}

func (r *repository) deleteZone(ctx context.Context, id int64) error {
	return r.db.Master().TxContext(ctx, "dns.zone.delete", func(ctx context.Context, db orm.DB) error {
		if _, e := db.ExecContext(ctx, "DELETE FROM dns_records WHERE zone_id=? AND version='draft'", id); e != nil {
			return e
		}
		_, e := db.ExecContext(ctx, "DELETE FROM dns_zones WHERE id=? AND version='draft'", id)
		return e
	})
}

func (r *repository) upsertRecord(ctx context.Context, x Record) error {
	if x.ID == 0 {
		return r.exec(ctx, "dns.record.create", "INSERT INTO dns_records(version,zone_id,name,type,value,ttl,priority) VALUES('draft',?,?,?,?,?,?)", x.ZoneID, x.Name, x.Type, x.Value, x.TTL, x.Priority)
	}
	return r.exec(ctx, "dns.record.update", "UPDATE dns_records SET zone_id=?,name=?,type=?,value=?,ttl=?,priority=? WHERE id=? AND version='draft'", x.ZoneID, x.Name, x.Type, x.Value, x.TTL, x.Priority, x.ID)
}

func (r *repository) deleteRecord(ctx context.Context, id int64) error {
	return r.exec(ctx, "dns.record.delete", "DELETE FROM dns_records WHERE id=? AND version='draft'", id)
}

func (r *repository) upsertForwarder(ctx context.Context, x Forwarder) error {
	if x.ID == 0 {
		return r.exec(ctx, "dns.forwarder.create", "INSERT INTO dns_forwarders(version,zone_name,address) VALUES('draft',?,?)", x.ZoneName, x.Address)
	}
	return r.exec(ctx, "dns.forwarder.update", "UPDATE dns_forwarders SET zone_name=?,address=? WHERE id=? AND version='draft'", x.ZoneName, x.Address, x.ID)
}

func (r *repository) deleteForwarder(ctx context.Context, id int64) error {
	return r.exec(ctx, "dns.forwarder.delete", "DELETE FROM dns_forwarders WHERE id=? AND version='draft'", id)
}

func (r *repository) upsertBlock(ctx context.Context, x Block) error {
	if x.ID == 0 {
		return r.exec(ctx, "dns.block.create", "INSERT INTO dns_blocks(version,name) VALUES('draft',?)", x.Name)
	}
	return r.exec(ctx, "dns.block.update", "UPDATE dns_blocks SET name=? WHERE id=? AND version='draft'", x.Name, x.ID)
}

func (r *repository) deleteBlock(ctx context.Context, id int64) error {
	return r.exec(ctx, "dns.block.delete", "DELETE FROM dns_blocks WHERE id=? AND version='draft'", id)
}
