package acctest

import (
	"net/http"
	"testing"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	stackguardianprovider "github.com/StackGuardian/terraform-provider-stackguardian/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func ProviderFactories(customHeader http.Header) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"stackguardian": providerserver.NewProtocol6WithError(stackguardianprovider.New("", customHeader)()),
	}
}

func TestAccPreCheck(t *testing.T) {
	cfg := config.Get()
	if cfg.ApiKey == "" {
		t.Fatal("STACKGUARDIAN_API_KEY must be set for acceptance tests")
	}
	if cfg.OrgName == "" {
		t.Fatal("STACKGUARDIAN_ORG_NAME must be set for acceptance tests")
	}
	if cfg.ApiUri == "" {
		t.Fatal("STACKGUARDIAN_API_URI must be set for acceptance tests")
	}
}

func SGClient() *sgclient.Client {
	cfg := config.Get()
	client := sgclient.NewClient(
		sgoption.WithBaseURL(cfg.ApiUri),
		sgoption.WithApiKey(cfg.FormatApiKey()),
	)

	return client
}
