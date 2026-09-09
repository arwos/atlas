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

	"go.arwos.org/atlas/pkg"
)

var Version = "v0.0.0-dev"

func main() {
	svc := goppy.New("atlas", Version, "local infrastructure management platform")

	svc.Plugins(
		orm.WithORM(sqlite.Name),
		web.WithServer(),
		jsonrpc.WithTransport(),
	)

	svc.Plugins(pkg.Plugins)

	svc.Run()
}
