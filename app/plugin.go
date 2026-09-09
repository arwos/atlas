package app

import "go.osspkg.com/goppy/v3/plugin"

var Plugin = plugin.Kind{
	Inject: NewAPI,
}
