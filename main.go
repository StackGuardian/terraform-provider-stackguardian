package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	sgprovider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is injected at build time from the release tag via
// -ldflags "-X main.version=..." (see .goreleaser.yml), and is "dev" for local
// builds. It has to stay a var: the linker cannot write to a const, and it
// ignores the flag without any error if the symbol is missing or not a var.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// This is deliberately not the Terraform Registry address
		// (registry.terraform.io/StackGuardian/stackguardian). It mirrors the
		// HOSTNAME/NAMESPACE/NAME that `make install` writes to under
		// ~/.terraform.d/plugins, so a locally built provider can be consumed
		// with source = "terraform/provider/stackguardian". It is also the key
		// that -debug emits in TF_REATTACH_PROVIDERS, so changing it here means
		// changing the Makefile and any local development configs to match.
		Address: "terraform/provider/stackguardian",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), sgprovider.New(version, http.Header{}), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
