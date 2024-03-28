package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	provider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: provider.Provider,
			Debug:        debug,
			ProviderAddr: "terraform/provider/stackguardian",
		})
}
