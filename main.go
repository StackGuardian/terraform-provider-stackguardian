package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"

	stackguardian_tf_provider "github.com/StackGuardian/terraform-provider-stackguardian/stackguardian-tf-provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return stackguardian_tf_provider.Provider()
		},
	})
}
