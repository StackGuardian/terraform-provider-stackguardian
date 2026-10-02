# Policies are guardrails evaluated during a run. Each example below supplies the
# policy body a different way, and scopes it differently.

resource "stackguardian_workflow_group" "production" {
  resource_name = "production"
  description   = "Production workloads"
}

# Example 1: policy written inline, enforced on one workflow group.
#
# The body lives in policy_input_data.data as a JSON string, so jsonencode keeps it
# readable. A run that violates it stops, because on_fail is FAIL.
resource "stackguardian_policy" "require_environment_tag" {
  resource_name = "require-environment-tag"
  description   = "EC2 instances must carry an Environment tag"
  policy_type   = "GENERAL"

  tags = ["compliance", "aws"]

  # A workflow group is a path, with no trailing slash. Use ["*"] to enforce
  # organization-wide instead.
  enforced_on = ["/wfgrps/${stackguardian_workflow_group.production.resource_name}"]

  policies_config = [{
    name    = "require-environment-tag"
    skip    = false
    on_fail = "FAIL"
    on_pass = "PASS"

    policy_input_data = {
      schema_type = "TIRITH_JSON"
      data = jsonencode({
        meta = {
          version           = "v1"
          required_provider = "stackguardian/terraform_plan"
        }
        eval_expression = "ec2-has-environment-tag"
        evaluators = [{
          id          = "ec2-has-environment-tag"
          description = "EC2 instances must be tagged Environment=Production"
          provider_args = {
            operation_type               = "attribute"
            terraform_resource_type      = "aws_instance"
            terraform_resource_attribute = "tags"
          }
          condition = {
            type            = "Contains"
            value           = { Environment = "Production" }
            error_tolerance = 0
          }
        }]
      })
    }

    # Omitting policy_vcs_config is valid: the backend applies the configuration
    # below at runtime, and a policy read back from the platform always carries it.
    # The provider does not fill it in for you, so writing it out explicitly is what
    # keeps configuration and state describing the same thing.
    policy_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        source_config_kind      = "SG_POLICY_FRAMEWORK" # Use for Tirith policies
        source_config_dest_kind = "INLINE"
      }
    }
  }]
}

# Example 2: hold a run for human approval rather than blocking it.
#
# on_fail = APPROVAL_REQUIRED pauses the run until enough approvers sign off. Here
# the evaluator fails whenever the run action is apply, so every apply is gated.
resource "stackguardian_policy" "approval_on_apply" {
  resource_name = "create-approval-on-apply"
  description   = "Approval needed before an apply"
  policy_type   = "GENERAL"

  enforced_on = ["/wfgrps/${stackguardian_workflow_group.production.resource_name}"]

  # Each approver is a user's email address, or an SSO group name to allow anyone in
  # that group. The fully qualified form "<user-pool-id>/local/<email>" is also
  # accepted. Read an existing policy with the stackguardian_policy data source to
  # see what your organization uses.
  approvers = [
    "platform-lead@example.com",
    "sre-oncall@example.com",
  ]
  number_of_approvals_required = 1

  policies_config = [{
    name    = "create-approval"
    skip    = false
    on_fail = "APPROVAL_REQUIRED"
    on_pass = "PASS"

    policy_input_data = {
      schema_type = "TIRITH_JSON"
      data = jsonencode({
        meta = {
          version           = "v1"
          required_provider = "stackguardian/terraform_plan"
        }
        eval_expression = "create-operation"
        evaluators = [{
          id          = "create-operation"
          description = "Apply detected -- stop and request approval"
          provider_args = {
            operation_type               = "action"
            terraform_resource_type      = "*"
            terraform_resource_attribute = ""
          }
          condition = {
            type            = "NotEquals"
            value           = "apply"
            error_tolerance = 0
          }
        }]
      })
    }

    policy_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        source_config_kind      = "SG_POLICY_FRAMEWORK" # Use for Tirith policies
        source_config_dest_kind = "INLINE"
      }
    }
  }]
}

# Example 3: use a policy published on the marketplace instead of writing one.
#
# policy_template_id is a path-form ID; templates published by StackGuardian live
# under the stackguardian org. There is no policy_input_data here -- the template
# carries the body.
resource "stackguardian_policy" "checkov_best_practices" {
  resource_name = "checkov-best-practices"
  description   = "Checkov best-practice checks from the marketplace"
  policy_type   = "GENERAL"

  enforced_on = ["*"] # the whole organization

  policies_config = [{
    name    = "checkov-best-practices"
    on_fail = "WARN" # record a warning, let the run continue
    on_pass = "PASS"

    policy_vcs_config = {
      use_marketplace_template = true
      policy_template_id       = "/stackguardian/checkov-best-practices:2"
    }
  }]
}

# Example 4: pull a Rego policy from your own git repository.
#
# The body is whatever the repo holds at ref, under working_dir. Set is_private and
# point auth at a VCS connector (`/integrations/<resource_name>`, which is what the
# connector's id attribute returns) to read a private repo.
resource "stackguardian_policy" "opa_from_git" {
  resource_name = "opa-from-git"
  description   = "Rego policies maintained alongside our platform code"
  policy_type   = "GENERAL"

  enforced_on = ["/wfgrps/${stackguardian_workflow_group.production.resource_name}"]

  policies_config = [{
    name    = "opa-from-git"
    on_fail = "FAIL"
    on_pass = "PASS"

    policy_vcs_config = {
      use_marketplace_template = false
      custom_source = {
        source_config_kind      = "OPA_REGO"
        source_config_dest_kind = "GITHUB_COM"
        config = {
          repo        = "https://github.com/example/platform-policies.git"
          ref         = "main"
          working_dir = "policies/terraform"
          is_private  = false
        }
      }
    }
  }]
}
