/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v3"
)

var Version = "v0.0.0-dev"

func main() {
	svc := goppy.New("atlas-runner", Version, "runner for local infrastructure management platform")

	svc.Run()
}
