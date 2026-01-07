package policy_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_policy" "%s" {
  resource_name = "%s"

  description = "Policy created using provider"
  
  policy_type = "GENERAL"

  number_of_approvals_required = 0

  tags = [
    "provider",
    "policy"
  ]

  policies_config = [{
    name = "example_policy_config" # Required: Policy name.

    skip = false # Optional: Whether to skip this policy.

    on_fail = "FAIL" # Required: Action on failure (e.g., FAIL, WARN, PASS, APPROVAL_REQUIRED).

    on_pass = "PASS" # Required: Action on pass (e.g., FAIL, WARN, PASS, APPROVAL_REQUIRED).

    policy_input_data = {      # Optional: Nested block for policy input data.
      schema_type = "RAW_JSON" # Required: Type of input schema (e.g., FORM_JSONSCHEMA, RAW_JSON).
      data = jsonencode(
        {
          "meta" : {
            "version" : "v1",
            "required_provider" : "stackguardian/terraform_plan"
          },
          "evaluators" : [
            {
              "id" : "ec2_has_environment_tag",
              "description" : "Ensure that EC2 instances have the 'Environment: Production' tag",
              "provider_args" : {
                "operation_type" : "attribute",
                "terraform_resource_type" : "aws_instance",
                "terraform_resource_attribute" : "tags"
              },
              "condition" : {
                "type" : "Contains",
                "value" : {
                  "Environment" : "Production"
                },
                "error_tolerance" : 2
              }
            }
          ],
          "eval_expression" : "ec2_has_environment_tag"
        }
      )
    }
  }]
}
`
	testAccResourceUpdate = `resource "stackguardian_policy" "%s" {
  resource_name = "%s"

  description = "Policy created using provider"

  policy_type = "GENERAL"

  number_of_approvals_required = 0

  tags = [
    "provider",
    "policy"
  ]

  policies_config = [{
    name = "example_policy_config" # Required: Policy name.

    skip = true # Optional: Whether to skip this policy.

    on_fail = "FAIL" # Required: Action on failure (e.g., FAIL, WARN, PASS, APPROVAL_REQUIRED).

    on_pass = "PASS" # Required: Action on pass (e.g., FAIL, WARN, PASS, APPROVAL_REQUIRED).

    policy_input_data = {      # Optional: Nested block for policy input data.
      schema_type = "RAW_JSON" # Required: Type of input schema (e.g., FORM_JSONSCHEMA, RAW_JSON).
      data = jsonencode(
        {
          "meta" : {
            "version" : "v1",
            "required_provider" : "stackguardian/terraform_plan"
          },
          "evaluators" : [
            {
              "id" : "ec2_has_environment_tag",
              "description" : "Ensure that EC2 instances have the 'Environment: Production' tag",
              "provider_args" : {
                "operation_type" : "attribute",
                "terraform_resource_type" : "aws_instance",
                "terraform_resource_attribute" : "tags"
              },
              "condition" : {
                "type" : "Contains",
                "value" : {
                  "Environment" : "Production"
                },
                "error_tolerance" : 2
              }
            }
          ],
          "eval_expression" : "ec2_has_environment_tag"
        }
      )
    }
  }]
}`
)

func TestAccPolicy(t *testing.T) {
	resourceName := "example-policy"
	policyName := "example-policy"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, resourceName, policyName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, resourceName, policyName),
			},
		},
	})
}

func TestAccPolicyRecreateOnExternalDelete(t *testing.T) {
	resourceName := "example-policy2"
	policyName := "example-policy2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, resourceName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_policy.%s", resourceName), "resource_name", policyName),
				),
			},
			{
				PreConfig: func() {
					client := acctest.SGClient()
					err := client.Policies.DeletePolicy(context.TODO(), os.Getenv("STACKGUARDIAN_ORG_NAME"), policyName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: fmt.Sprintf(testAccResource, resourceName, policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_policy.%s", resourceName), "resource_name", policyName),
				),
			},
		},
	})
}

func TestAccPolicyOptionalId(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	testResource := `resource "stackguardian_policy" "example-policy3" {
  id = "example_policy3"
  resource_name = "example_policy3_resource_name"
  policy_type = "GENERAL"
}`
	testUpdateResource := `resource "stackguardian_policy" "example-policy3" {
  id = "example_policy3"
  resource_name = "example_policy3_resource_name"
  policy_type = "GENERAL"
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResource,
				//Check:  resource.TestCheckResourceAttr("aws-cloud-connector-example2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_policy.example-policy3",
						tfjsonpath.New("id"),
						knownvalue.StringExact("example_policy3"),
					),
				},
			},
			{
				Config: testUpdateResource,
			},
		},
	})
}
