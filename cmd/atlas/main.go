/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/plugins/orm"
	"go.osspkg.com/goppy/v3/plugins/orm/clients/sqlite"
	"go.osspkg.com/goppy/v3/plugins/web"
	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"
	"go.osspkg.com/logx"

	"go.arwos.org/atlas/app"
	"go.arwos.org/atlas/pkg"
)

var Version = "v0.0.0-dev"

func main() {
	svc := goppy.New("atlas", Version, "local infrastructure management platform")

	svc.Plugins(
		orm.WithORM(sqlite.Name),
		web.WithServer(),
		jsonrpc.WithTransport(
			jsonrpc.Path("/jsonrpc"),
			jsonrpc.ErrHandler(func(method string, err error) error {
				logx.Error("json-rpc error", "method", method, "err", err)
				return err
			}),
		),
	)

	svc.Plugins(
		app.Plugin,
		pkg.Plugins,
	)

	svc.Run()
}
