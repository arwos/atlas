/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"go.osspkg.com/goppy/v3/plugins/orm"
	"go.osspkg.com/logx"

	"go.arwos.org/atlas/pkg/database"
)

type repository struct{ db *database.Service }

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logx.Error("DHCP repository", "do", "close rows", "err", err)
	}
}

func (r *repository) exec(ctx context.Context, n, q string, a ...any) error {
	return r.db.Master().CallContext(ctx, n, func(ctx context.Context, db orm.DB) error { _, e := db.ExecContext(ctx, q, a...); return e })
}
func (r *repository) bootstrap(ctx context.Context, b BootstrapConfig) error {
	var n int
	err := r.db.Master().CallContext(ctx, "dhcp.bootstrap", func(ctx context.Context, db orm.DB) error {
		rows, e := db.QueryContext(ctx, "SELECT count(*) FROM dhcp_subnets WHERE version=?", activeVersion)
		if e != nil {
			return e
		}
		defer closeRows(rows)
		if rows.Next() {
			return rows.Scan(&n)
		}
		return nil
	})
	if err != nil || n > 0 {
		return err
	}
	return r.db.Master().TxContext(ctx, "dhcp.bootstrap", func(ctx context.Context, db orm.DB) error {
		for _, v := range []string{draftVersion, activeVersion} {
			if _, e := db.ExecContext(ctx, "INSERT INTO dhcp_subnets(version,interface_name,cidr,lease_seconds) VALUES(?,?,?,?)", v, b.Interface, b.CIDR, b.LeaseSeconds); e != nil {
				return e
			}
		}
		return nil
	})
}
func (r *repository) snapshot(ctx context.Context, v string) (out Snapshot, err error) {
	err = r.db.Master().CallContext(ctx, "dhcp.snapshot", func(ctx context.Context, db orm.DB) error {
		rows, e := db.QueryContext(ctx, `SELECT id,interface_name,cidr,lease_seconds,router,dns_servers,domain_search,ntp_servers,mtu,classless_routes FROM dhcp_subnets WHERE version=? ORDER BY id`, v)
		if e != nil {
			return e
		}
		defer closeRows(rows)
		for rows.Next() {
			var s Subnet
			var dns, ntp, routes string
			if e = rows.Scan(&s.ID, &s.Interface, &s.CIDR, &s.LeaseSeconds, &s.Router, &dns, &s.DomainSearch, &ntp, &s.MTU, &routes); e != nil {
				return e
			}
			if e = json.Unmarshal([]byte(dns), &s.DNSServers); e != nil {
				logx.Error("DHCP repository", "do", "decode dns servers", "err", e, "subnet_id", s.ID)
				return e
			}
			if e = json.Unmarshal([]byte(ntp), &s.NTPServers); e != nil {
				logx.Error("DHCP repository", "do", "decode ntp servers", "err", e, "subnet_id", s.ID)
				return e
			}
			if e = json.Unmarshal([]byte(routes), &s.ClasslessRoutes); e != nil {
				logx.Error("DHCP repository", "do", "decode classless routes", "err", e, "subnet_id", s.ID)
				return e
			}
			out.Subnets = append(out.Subnets, s)
		}
		if e = rows.Err(); e != nil {
			return e
		}
		rows, e = db.QueryContext(ctx, "SELECT id,subnet_id,mac,ip FROM dhcp_reservations WHERE version=?", v)
		if e != nil {
			return e
		}
		defer closeRows(rows)
		for rows.Next() {
			var x Reservation
			if e = rows.Scan(&x.ID, &x.SubnetID, &x.MAC, &x.IP); e != nil {
				return e
			}
			out.Reservations = append(out.Reservations, x)
		}
		rows, e = db.QueryContext(ctx, "SELECT id,mac FROM dhcp_blocks WHERE version=?", v)
		if e != nil {
			return e
		}
		defer closeRows(rows)
		for rows.Next() {
			var x Block
			if e = rows.Scan(&x.ID, &x.MAC); e != nil {
				return e
			}
			out.Blocks = append(out.Blocks, x)
		}
		return rows.Err()
	})
	return
}
func (r *repository) replaceActive(ctx context.Context) error {
	return r.db.Master().TxContext(ctx, "dhcp.apply", func(ctx context.Context, db orm.DB) error {
		for _, q := range []string{"DELETE FROM dhcp_reservations WHERE version='active'", "DELETE FROM dhcp_blocks WHERE version='active'", "DELETE FROM dhcp_subnets WHERE version='active'"} {
			if _, e := db.ExecContext(ctx, q); e != nil {
				return e
			}
		}
		if _, e := db.ExecContext(ctx, `INSERT INTO dhcp_subnets(version,interface_name,cidr,lease_seconds,router,dns_servers,domain_search,ntp_servers,mtu,classless_routes) SELECT 'active',interface_name,cidr,lease_seconds,router,dns_servers,domain_search,ntp_servers,mtu,classless_routes FROM dhcp_subnets WHERE version='draft'`); e != nil {
			return e
		}
		if _, e := db.ExecContext(ctx, `INSERT INTO dhcp_reservations(version,subnet_id,mac,ip) SELECT 'active',a.id,r.mac,r.ip FROM dhcp_reservations r JOIN dhcp_subnets d ON d.id=r.subnet_id AND d.version='draft' JOIN dhcp_subnets a ON a.version='active' AND a.interface_name=d.interface_name AND a.cidr=d.cidr WHERE r.version='draft'`); e != nil {
			return e
		}
		_, e := db.ExecContext(ctx, `INSERT INTO dhcp_blocks(version,mac) SELECT 'active',mac FROM dhcp_blocks WHERE version='draft'`)
		return e
	})
}
func (r *repository) upsertSubnet(ctx context.Context, s Subnet) error {
	dns, err := json.Marshal(s.DNSServers)
	if err != nil {
		logx.Error("DHCP repository", "do", "encode dns servers", "err", err)
		return err
	}
	ntp, err := json.Marshal(s.NTPServers)
	if err != nil {
		logx.Error("DHCP repository", "do", "encode ntp servers", "err", err)
		return err
	}
	routes, err := json.Marshal(s.ClasslessRoutes)
	if err != nil {
		logx.Error("DHCP repository", "do", "encode classless routes", "err", err)
		return err
	}
	if s.ID == 0 {
		return r.exec(ctx, "dhcp.subnet.create", `INSERT INTO dhcp_subnets(version,interface_name,cidr,lease_seconds,router,dns_servers,domain_search,ntp_servers,mtu,classless_routes) VALUES('draft',?,?,?,?,?,?,?,?,?)`, s.Interface, s.CIDR, s.LeaseSeconds, s.Router, string(dns), s.DomainSearch, string(ntp), s.MTU, string(routes))
	}
	return r.exec(ctx, "dhcp.subnet.update", `UPDATE dhcp_subnets SET interface_name=?,cidr=?,lease_seconds=?,router=?,dns_servers=?,domain_search=?,ntp_servers=?,mtu=?,classless_routes=? WHERE id=? AND version='draft'`, s.Interface, s.CIDR, s.LeaseSeconds, s.Router, string(dns), s.DomainSearch, string(ntp), s.MTU, string(routes), s.ID)
}
func (r *repository) deleteSubnet(ctx context.Context, id int64) error {
	return r.exec(ctx, "dhcp.subnet.delete", "DELETE FROM dhcp_subnets WHERE id=? AND version='draft'", id)
}
func (r *repository) upsertReservation(ctx context.Context, x Reservation) error {
	if x.ID == 0 {
		return r.exec(ctx, "dhcp.reservation.create", "INSERT INTO dhcp_reservations(version,subnet_id,mac,ip) VALUES('draft',?,?,?)", x.SubnetID, x.MAC, x.IP)
	}
	return r.exec(ctx, "dhcp.reservation.update", "UPDATE dhcp_reservations SET subnet_id=?,mac=?,ip=? WHERE id=? AND version='draft'", x.SubnetID, x.MAC, x.IP, x.ID)
}
func (r *repository) deleteReservation(ctx context.Context, id int64) error {
	return r.exec(ctx, "dhcp.reservation.delete", "DELETE FROM dhcp_reservations WHERE id=? AND version='draft'", id)
}
func (r *repository) upsertBlock(ctx context.Context, x Block) error {
	if x.ID == 0 {
		return r.exec(ctx, "dhcp.block.create", "INSERT INTO dhcp_blocks(version,mac) VALUES('draft',?)", x.MAC)
	}
	return r.exec(ctx, "dhcp.block.update", "UPDATE dhcp_blocks SET mac=? WHERE id=? AND version='draft'", x.MAC, x.ID)
}
func (r *repository) deleteBlock(ctx context.Context, id int64) error {
	return r.exec(ctx, "dhcp.block.delete", "DELETE FROM dhcp_blocks WHERE id=? AND version='draft'", id)
}
func (r *repository) leases(ctx context.Context) (out []Lease, err error) {
	err = r.db.Master().CallContext(ctx, "dhcp.leases", func(ctx context.Context, db orm.DB) error {
		rows, e := db.QueryContext(ctx, "SELECT id,subnet_id,mac,ip,expires_at FROM dhcp_leases ORDER BY expires_at DESC")
		if e != nil {
			return e
		}
		defer closeRows(rows)
		for rows.Next() {
			var l Lease
			var u int64
			if e = rows.Scan(&l.ID, &l.SubnetID, &l.MAC, &l.IP, &u); e != nil {
				return e
			}
			l.ExpiresAt = time.Unix(u, 0)
			out = append(out, l)
		}
		return rows.Err()
	})
	return
}
func (r *repository) revokeLease(ctx context.Context, id int64) error {
	return r.exec(ctx, "dhcp.lease.revoke", "DELETE FROM dhcp_leases WHERE id=?", id)
}
