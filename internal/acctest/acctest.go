package acctest

import (
	"os"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	stackguardianprovider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"stackguardian": providerserver.NewProtocol6WithError(stackguardianprovider.New("")()),
	}
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("STACKGUARDIAN_API_KEY"); v == "" {
		t.Fatal("STACKGUARDIAN_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("STACKGUARDIAN_ORG_NAME"); v == "" {
		t.Fatal("STACKGUARDIAN_ORG_NAME must be set for acceptance tests")
	}
	if v := os.Getenv("STACKGUARDIAN_API_URI"); v == "" {
		t.Fatal("STACKGUARDIAN_API_URI must be set for acceptance tests")
	}
}

func SGClient() *sgclient.Client {
	client := sgclient.NewClient(
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithApiKey("apikey "+os.Getenv("STACKGUARDIAN_API_KEY")),
	)

	return client
}
