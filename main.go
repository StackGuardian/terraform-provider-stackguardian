package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	provider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: provider.Provider,
			// TODO: fill in ProviderAddr
		})
}
