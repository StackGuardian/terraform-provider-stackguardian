# ------------------------------------------------------------------
# stackguardian_workflow_template
# ------------------------------------------------------------------
resource "stackguardian_workflow_template" "example" {
  template_name      = "my-terraform-template-1"
  source_config_kind = "TERRAFORM"
  is_public          = "0"
  description        = "Example workflow template with all attributes set"
  tags               = ["terraform", "production", "example"]

  context_tags = {
    team        = "platform"
    cost-center = "eng-123"
  }

  runtime_source = {
    source_config_dest_kind = "GITHUB_COM"
    config = {
      auth                       = "/integrations/my-github-connector"
      git_core_auto_crlf         = false
      git_sparse_checkout_config = "--cone infra/modules"
      include_sub_module         = true
      is_private                 = true
      ref                        = "main"
      repo                       = "https://github.com/example/terraform-modules.git"
      working_dir                = "modules/network"
    }
  }

  vcs_triggers = {
    type = "GITHUB_COM"
    create_tag = {
      create_revision = {
        enabled = true
      }
    }
  }
}

# ------------------------------------------------------------------
# stackguardian_workflow_template_revision
# ------------------------------------------------------------------
resource "stackguardian_workflow_template_revision" "example" {
  template_id        = stackguardian_workflow_template.example.id
  alias              = "v1"
  description        = "Example workflow template revision with all attributes set"
  notes              = "Initial revision with full configuration"
  source_config_kind = "TERRAFORM"
  is_public          = "0"

  user_job_cpu    = 500
  user_job_memory = 1024

  tags = ["terraform", "production", "example"]
  context_tags = {
    team = "platform"
  }

  approvers                    = ["user1@example.com", "user2@example.com"]
  number_of_approvals_required = 1

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "AWS_DEFAULT_REGION"
        text_value = "us-east-1"
      }
    },
  ]

  runner_constraints = {
    type  = "private"
    names = ["my-runner-group"]
  }

  user_schedules = [
    {
      name  = "nightly-drift-check"
      cron  = "cron(0 2 * * ? *)"
      state = "ENABLED"
      desc  = "Run drift check nightly at 2 AM UTC"
    }
  ]

  deployment_platform_config = [
    {
      kind = "AWS_RBAC"
      config = {
        integration_id = "/integrations/aws-rbac-integration"
        profile_name   = "default"
      }
    }
  ]

  terraform_config = {
    terraform_version           = "1.5.7"
    drift_check                 = true
    drift_cron                  = "cron(0 3 * * ? *)"
    managed_terraform_state     = true
    approval_pre_apply          = true
    terraform_plan_options      = "-var-file=prod.tfvars"
    terraform_init_options      = "-upgrade"
    timeout                     = 3600
    run_pre_init_hooks_on_drift = true

    terraform_bin_path = [
      {
        source    = "/opt/terraform/bin"
        target    = "/usr/local/bin"
        read_only = true
      }
    ]

    pre_init_hooks   = ["echo pre-init"]
    pre_plan_hooks   = ["echo pre-plan"]
    post_plan_hooks  = ["echo post-plan"]
    pre_apply_hooks  = ["echo pre-apply"]
    post_apply_hooks = ["echo post-apply"]

    pre_plan_wf_steps_config = [
      {
        name                = "lint"
        wf_step_template_id = "/stackguardian/tflint:1"
        approval            = false
        timeout             = 300
        cmd_override        = "tflint --recursive"
        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "TFLINT_LOG"
              text_value = "info"
            }
          }
        ]
        mount_points = [
          {
            source    = "/cache/tflint"
            target    = "/root/.tflint.d"
            read_only = false
          }
        ]
        wf_step_input_data = {
          schema_type = "FORM_JSONSCHEMA"
          data        = jsonencode({ severity = "warning" })
        }
      }
    ]

    post_apply_wf_steps_config = [
      {
        name                = "notify-slack"
        wf_step_template_id = "/stackguardian/slack-notify:1"
        approval            = false
      }
    ]
  }
}
