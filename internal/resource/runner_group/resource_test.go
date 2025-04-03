package runnergroup_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var azureStorageBackendAccessKey = os.Getenv("TEST_AZURE_STORAGE_BACKEND_ACCESS_KEY")

var (
	testAccResource = `resource "stackguardian_runner_group" "example-runner-group" {
  resource_name = "example-runner-group"

  description = "RunnerGroup created using provider"
  
  tags = [
    "provider",
    "runnergroup"
  ]
	
	storage_backend_config = {
		type = "aws_s3"
		aws_region = "eu-central-1"
		s3_bucket_name = "http-proxy-private-runner"
		auth = {
			integration_id = "/integrations/taher-aws"
		}
	}

}
`
	testAccResourceUpdate = `resource "stackguardian_runner_group" "example-runner-group" {
  resource_name = "example-runner-group"

  description = "RunnerGroup created using provider"
  
  tags = [
    "provider",
    "runnergroup",
		"update"
  ]
	
	storage_backend_config = {
		type = "aws_s3"
		aws_region = "eu-central-1"
		s3_bucket_name = "http-proxy-private-runner"
		auth = {
			integration_id = "/integrations/taher-aws"
		}
	}
}`
)

func TestAccRunnerGroupAWSS3(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResource,
			},
			{
				Config: testAccResourceUpdate,
			},
		},
	})
}

func TestAccRunnerGroupAzureBlobStorage(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "stackguardian_runner_group" "example-runner-group2" {
  max_number_of_runners = 2
  resource_name     = "runnergroup"
  storage_backend_config = {
    azure_blob_storage_access_key   = "%s"
    azure_blob_storage_account_name = "blobfbitv1"
    type                            = "azure_blob_storage"
  }
}`, azureStorageBackendAccessKey),
			},
			{
				Config: fmt.Sprintf(`resource "stackguardian_runner_group" "example-runner-group2" {
  max_number_of_runners = 5
  resource_name     = "runnergroup"
  storage_backend_config = {
    azure_blob_storage_access_key   = "%s"
    azure_blob_storage_account_name = "blobfbitv1"
    type                            = "azure_blob_storage"
  }
}`, azureStorageBackendAccessKey),
			},
		},
	})
}
