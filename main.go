package main

import (
	"github.com/eveldcorp/waypoint-plugin-terraform/builder"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)

func main() {
	sdk.Main(sdk.WithComponents(
		&builder.Builder{},
		// &platform.Platform{},
	))
}
