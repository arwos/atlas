package pkg

import (
	"go.arwos.org/atlas/pkg/database"
	"go.arwos.org/atlas/pkg/dhcp"
	"go.osspkg.com/goppy/v3/plugin"
)

var Plugins = plugin.Inject(
	database.NewService,
	dhcp.NewService,
)
