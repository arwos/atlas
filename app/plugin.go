/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import "go.osspkg.com/goppy/v3/plugin"

// Plugin registers the application API adapter.
var Plugin = plugin.Kind{
	Inject: NewAPI,
}
