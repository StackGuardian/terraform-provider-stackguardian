package policy_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_policy" "example-policy" {
  resource_name = "example-policy"

  description = "Policy created using provider"

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
	testAccResourceUpdate = `resource "stackguardian_policy" "example-policy" {
  resource_name = "example-policy"

  description = "Policy created using provider"

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
