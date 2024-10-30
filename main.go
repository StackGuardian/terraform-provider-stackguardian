package main

import (
	"context"
	"flag"
	"log"

	sgprovider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

const Version = "0.0.1"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// NOTE: This is not a typical Terraform Registry provider address,
		// such as registry.terraform.io/hashicorp/hashicups. This specific
		// provider address is used in these tutorials in conjunction with a
		// specific Terraform CLI configuration for manual development testing
		// of this provider.
		Address: "terraform/provider/stackguardian",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), sgprovider.New(Version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
