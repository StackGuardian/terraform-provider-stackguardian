package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// var testAccProviders map[string]*schema.Provider
var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"stackguardian": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("STACKGUARDIAN_ORG_NAME") == "" {
		t.Fatal("STACKGUARDIAN_ORG_NAME must be set for acceptance tests")
	}

	if os.Getenv("STACKGUARDIAN_API_KEY") == "" {
		t.Fatal("STACKGUARDIAN_API_KEY must be set for acceptance tests")
	}

	// Needed ?
	// err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	// if err != nil {
	// 		t.Fatal(err)
	// }
}
