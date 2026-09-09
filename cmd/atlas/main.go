package main

import (
	"go.arwos.org/atlas/pkg"
	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/plugins/orm"
	"go.osspkg.com/goppy/v3/plugins/orm/clients/sqlite"
)

var Version = "v0.0.0-dev"

func main() {
	svc := goppy.New("atlas", Version, "local infrastructure management platform")

	svc.Plugins(
		orm.WithORM(sqlite.Name),
	)

	svc.Plugins(pkg.Plugins)

	svc.Run()
}
