# Example 1: Basic workflow with VCS config (minimal required fields)
resource "stackguardian_workflow_git" "basic" {
  workflow_group_id = "my-workflow-group"
  id                = "my-workflow"
  wf_type           = "CUSTOM"

  description = "A basic git-backed workflow"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }
}

# Example 2: Workflow with Nested Workflow Group
resource "stackguardian_workflow_git" "nested_wfgrp" {
  workflow_group_id = "my-workflow-group/my-nested-group"
  id                = "my-workflow-with-env"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  environment_variables = [
    {
      kind = "PLAIN_TEXT"
      config = {
        var_name   = "MY_VAR"
        text_value = "my-value"
      }
    }
  ]
}

# Example 3: Workflow with user schedules
resource "stackguardian_workflow_git" "with_schedule" {
  workflow_group_id = "my-workflow-group"
  id                = "my-scheduled-workflow"
  wf_type           = "CUSTOM"

  vcs_config = {
    iac_vcs_config = {
      custom_source = {
        source_config_dest_kind = "GIT_OTHER"
        config = {
          is_private = false
          repo       = "https://github.com/my-org/my-repo.git"
        }
      }
    }
  }

  user_schedules = [
    {
      cron  = "0 8 ? * MON *"
      state = "ENABLED"
      desc  = "Runs on schedule"
    }
  ]
}
