package main

import (
	"go.osspkg.com/goppy/v3"
)

var Version = "v0.0.0-dev"

func main() {
	svc := goppy.New("atlas-runner", Version, "runner for local infrastructure management platform")

	svc.Run()
}
