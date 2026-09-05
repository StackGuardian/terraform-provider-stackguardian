# A workflow runs IaC straight from a git repository, with no template in between.
# Every workflow lives inside a workflow group, so each example declares its group
# and references it -- that also makes Terraform create the two in the right order.

# Example 1: the smallest workflow that runs.
#
# A public repository needs no connector: is_private = false and the runner clones
# it anonymously. wf_type decides which engine executes the code.
resource "stackguardian_workflow_group" "sandbox" {
  resource_name = "sandbox"
  description   = "Scratch workflows, safe to destroy"
}

resource "stackguardian_workflow_git" "basic" {
  workflow_group_id = stackguardian_workflow_group.sandbox.resource_name
  id                = "hello-terraform"
  wf_type           = "TERRAFORM"

  description = "Smallest git-backed workflow that runs"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          repo       = "https://github.com/example-org/hello-terraform.git"
          is_private = false
        }
      }
    }
  }
}

# Example 2: a nested group, input variables and environment variables.
#
# Groups nest with `/`, and the nested group is an ordinary resource of its own --
# creating "platform/networking" does not create "platform" for you, so both are
# declared. Anything secret belongs in a VAULT_SECRET variable or a `${secret::...}`
# reference, never in PLAIN_TEXT, which is visible in configuration and state.
resource "stackguardian_workflow_group" "platform" {
  resource_name = "platform"
  description   = "Platform team workflows"
}

resource "stackguardian_workflow_group" "networking" {
  resource_name = "platform/networking"
  description   = "Networking workflows, owned by the platform team"
}

resource "stackguardian_workflow_git" "vpc_staging" {
  workflow_group_id = stackguardian_workflow_group.networking.resource_name
  id                = "vpc-staging"
  wf_type           = "TERRAFORM"

  description = "Staging VPC, deployed from the platform monorepo"
  tags        = ["networking", "staging"]

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          repo        = "https://github.com/example-org/platform-infra.git"
          ref         = "main"
          working_dir = "environments/staging/vpc"
          is_private  = false
        }
      }
    }

    # Values handed to the IaC itself. `data` is a JSON string, so jsonencode keeps
    # it readable. A "${secret::<name>}" reference is resolved by StackGuardian at
    # run time, so the value never lands in configuration or state. It is written
    # "$${" below because Terraform would otherwise read it as its own
    # interpolation; the doubled dollar escapes that and sends a literal "${".
    iac_input_data = {
      schema_type = "RAW_JSON"
      data = jsonencode({
        vpc_cidr    = "10.20.0.0/16"
        environment = "staging"
        api_key     = "$${secret::platform_api_key}"
      })
    }
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "TF_LOG"
        text_value = "INFO"
      }
    },
    {
      kind = "VAULT_SECRET"
      config = {
        var_name  = "DATADOG_API_KEY"
        secret_id = "/secrets/datadog-api-key"
      }
    },
  ]
}

# Example 3: a production workflow, end to end.
#
# This is the shape a real deployment tends to take: a private repository read
# through a VCS connector, credentials from a cloud connector, drift detection on a
# schedule, an approval gate before apply, and notifications when a run finishes.
resource "stackguardian_workflow_group" "production" {
  resource_name = "production"
  description   = "Production workloads"
}

resource "stackguardian_workflow_git" "vpc_production" {
  workflow_group_id = stackguardian_workflow_group.production.resource_name
  id                = "vpc-production"
  wf_type           = "TERRAFORM"

  description = "Production VPC and its managed identities"
  tags        = ["networking", "production"]

  # Contextual tags are free-form key/value pairs used for filtering and reporting.
  context_tags = {
    team        = "platform"
    cost_center = "infra-001"
  }

  # Credentials the workflow deploys with. `kind` must match the connector's own
  # kind, and integration_id is a path-form ID -- reference a stackguardian_connector
  # resource's id instead of a literal string where you manage the connector too.
  deployment_platform_config = [
    {
      kind = "AWS_RBAC"
      config = {
        integration_id = "/integrations/production-aws"
        profile_name   = "production-aws"
      }
    }
  ]

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        # GITHUB_COM rather than GIT_OTHER, so `auth` names a VCS connector that
        # can read the private repository.
        source_config_dest_kind = "GITHUB_COM"
        config = {
          repo               = "https://github.com/example-org/platform-infra.git"
          ref                = "main"
          working_dir        = "environments/production/vpc"
          is_private         = true
          auth               = "/integrations/example-github"
          include_sub_module = false
        }
      }
    }

    iac_input_data = {
      schema_type = "RAW_JSON"
      data = jsonencode({
        vpc_cidr    = "10.10.0.0/16"
        environment = "production"
      })
    }
  }

  terraform_config = {
    # Bare version, with no TERRAFORM-/OPENTOFU- prefix; a patch wildcard such as
    # "1.9.x" is accepted too. The engine comes from wf_type, not from this value.
    terraform_version = "1.5.7"

    # StackGuardian stores the state file. Set false only if state lives in your
    # own backend.
    managed_terraform_state = true

    # Hold every apply for a human. `approvers` below decides who may release it --
    # a gate with no approvers can be released by anyone.
    approval_pre_apply = true

    # Re-read real infrastructure on a schedule and report what no longer matches.
    drift_check = true
    drift_cron  = "0 */6 * * ? *" # every six hours

    terraform_plan_options = "-parallelism=10"

    # Shell commands run inside the plan container at each lifecycle point. Hooks
    # are skipped during drift runs unless the matching run_*_on_drift flag is set.
    pre_init_hooks           = ["echo 'preparing workspace'"]
    post_plan_hooks          = ["echo 'plan complete'"]
    run_pre_init_hooks_on_drift = false

    timeout = 3600
  }

  approvers                    = ["platform-lead@example.com", "sre-oncall@example.com"]
  number_of_approvals_required = 1

  # "shared" runs on StackGuardian-managed infrastructure. Use "private" with
  # `names` to pin the workflow to your own runner group.
  runner_constraints = {
    type = "shared"
  }

  environment_variables = [
    {
      kind = "VAULT_SECRET"
      config = {
        var_name  = "TF_VAR_datadog_api_key"
        secret_id = "/secrets/datadog-api-key"
      }
    }
  ]

  # Run automatically when main moves. `createWfRun` is the supported action key,
  # and the run only happens when the pushed branch equals tracked_branch.
  vcs_triggers = {
    tracked_branch = "main"
    push = {
      createWfRun = { enabled = true }
    }
    pull_request_opened = {
      createWfRun = { enabled = true }
    }
    # Pull requests get a plan only; the apply happens on merge to main.
    plan_only = true
  }

  # A nightly apply keeps production converged with the repository.
  user_schedules = [
    {
      cron  = "0 2 * * ? *"
      state = "ENABLED"
      desc  = "Nightly convergence apply at 02:00 UTC"
    }
  ]

  # Post-run actions, keyed by outcome. drift_detected only fires while
  # terraform_config.drift_check is enabled.
  mini_steps = {
    notifications = {
      email = {
        completed      = [{ recipients = ["platform-team@example.com"] }]
        errored        = [{ recipients = ["platform-team@example.com", "sre-oncall@example.com"] }]
        drift_detected = [{ recipients = ["platform-team@example.com"] }]
      }
    }
  }

  # Raise these for large states; the platform defaults are 512 / 1024.
  user_job_cpu    = 1024
  user_job_memory = 2048
}
