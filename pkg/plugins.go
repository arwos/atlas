/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package pkg registers Atlas domain plugins.
package pkg

import (
	"go.osspkg.com/goppy/v3/plugin"
	"go.osspkg.com/goppy/v3/plugins/orm"

	"go.arwos.org/atlas/pkg/database"
	"go.arwos.org/atlas/pkg/dhcp"
)

// Plugins registers the Atlas domain services and configuration.
var Plugins = plugin.Inject(
	database.NewService,
	orm.WithMigration(database.Migrations()),
	plugin.Kind{
		Config: &dhcp.ConfigGroup{},
		Inject: dhcp.NewService,
	},
)
