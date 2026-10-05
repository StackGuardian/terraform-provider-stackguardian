# Example 1: Minimal stack — every per-workflow field left unset below is
# inherited from the stack template revision referenced by template_group_id,
# then from that workflow's workflow template revision.
# workflows_config.workflows must list the id of every workflow
# the referenced revision defines, in the same order — the id alone is
# enough per entry, everything else about that workflow is inherited.
resource "stackguardian_stack" "basic" {
  id                = "my-stack"
  workflow_group_id = "my-workflow-group"
  template_group_id = "my-stack-template:1"

  workflows_config = {
    workflows = [
      { id = "d8dfaf15-2ad9-da29-8af0-c6b288b12089" }
    ]
  }
}

# Example 2: Overriding a workflow's settings, declaring the stack's actions, and
# declaring a user schedule. workflows_config.workflows[].id must match the id
# of a workflow declared on the referenced stack template revision.
resource "stackguardian_stack" "with_overrides" {
  id                = "my-stack-custom"
  workflow_group_id = "my-workflow-group"
  template_group_id = "my-stack-template:1"

  tags = ["production"]

  workflows_config = {
    workflows = [
      {
        id = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"

        terraform_config = {
          terraform_version = "1.5.7"
        }

        environment_variables = [
          {
            kind = "PLAIN_TEXT"
            config = {
              var_name   = "ENVIRONMENT"
              text_value = "production"
            }
          }
        ]

        user_schedules = [
          {
            cron  = "0 8 ? * MON *"
            state = "ENABLED"
          }
        ]
      }
    ]
  }

  # Leaving actions unset (as in the "basic" example above) inherits the
  # actions resolved from the stack template revision —
  # either its own actions, or a generated apply/plan/destroy set. Declaring
  # it here REPLACES that inherited value wholesale, not merges into it — so
  # every action the stack needs ("apply" and "destroy" here) must be declared,
  # or it won't exist on the stack at all.
  actions = {
    apply = {
      name = "apply"
      order = {
        "d8dfaf15-2ad9-da29-8af0-c6b288b12089" = {
          parameters = {
            terraform_action = {
              action = "apply"
            }
          }
        }
      }
    }
    destroy = {
      name = "destroy"
      order = {
        "d8dfaf15-2ad9-da29-8af0-c6b288b12089" = {
          parameters = {
            terraform_action = {
              action = "destroy"
            }
          }
        }
      }
    }
  }
}
