package workflowtemplaterevision

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ministepsNotificationRecepients = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	},
}

var ministepsWebhooks = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name": schema.StringAttribute{
				Required: true,
			},
			"webhook_url": schema.StringAttribute{
				Required: true,
			},
			"webhook_secret": schema.StringAttribute{
				Optional: true,
			},
		},
	},
}

var ministepsWorkflowChaining = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				Required: true,
			},
			"stack_id": schema.StringAttribute{
				Optional: true,
			},
			"stack_run_payload": schema.StringAttribute{
				Optional: true,
			},
			"workflow_id": schema.StringAttribute{
				Optional: true,
			},
			"workflow_run_payload": schema.StringAttribute{
				Optional: true,
			},
		},
	},
}

var terraformConfigSchema = schema.SingleNestedAttribute{
	MarkdownDescription: "Terraform configuration.",
	Optional:            true,
	Computed:            true,
	Attributes: map[string]schema.Attribute{
		"terraform_version": schema.StringAttribute{
			MarkdownDescription: "Terraform version to use.",
			Optional:            true,
			Computed:            true,
		},
		"drift_check": schema.BoolAttribute{
			MarkdownDescription: "Enable drift check.",
			Optional:            true,
			Computed:            true,
		},
		"drift_cron": schema.StringAttribute{
			MarkdownDescription: "Cron expression for drift check.",
			Optional:            true,
			Computed:            true,
		},
		"managed_terraform_state": schema.BoolAttribute{
			MarkdownDescription: "Enable managed terraform state.",
			Optional:            true,
			Computed:            true,
		},
		"approval_pre_apply": schema.BoolAttribute{
			MarkdownDescription: "Require approval before apply.",
			Optional:            true,
			Computed:            true,
		},
		"terraform_plan_options": schema.StringAttribute{
			MarkdownDescription: "Additional options for terraform plan.",
			Optional:            true,
			Computed:            true,
		},
		"terraform_init_options": schema.StringAttribute{
			MarkdownDescription: "Additional options for terraform init.",
			Optional:            true,
			Computed:            true,
		},
		"terraform_bin_path": schema.ListNestedAttribute{
			MarkdownDescription: "Mount points for terraform binary.",
			Optional:            true,
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"source": schema.StringAttribute{
						MarkdownDescription: "Source path for mount point.",
						Optional:            true,
					},
					"target": schema.StringAttribute{
						MarkdownDescription: "Target path for mount point.",
						Optional:            true,
					},
					"read_only": schema.BoolAttribute{
						Optional: true,
					},
				},
			},
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: "Timeout for terraform operations in seconds.",
			Optional:            true,
			Computed:            true,
		},
		"post_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: "Workflow steps configuration to run after apply.",
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: "Workflow steps configuration to run before apply.",
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: "Workflow steps configuration to run before plan.",
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"post_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: "Workflow steps configuration to run after plan.",
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_init_hooks": schema.ListAttribute{
			MarkdownDescription: "Hooks to run before init.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_plan_hooks": schema.ListAttribute{
			MarkdownDescription: "Hooks to run before plan.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_plan_hooks": schema.ListAttribute{
			MarkdownDescription: "Hooks to run after plan.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_apply_hooks": schema.ListAttribute{
			MarkdownDescription: "Hooks to run before apply.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_apply_hooks": schema.ListAttribute{
			MarkdownDescription: "Hooks to run after apply.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"run_pre_init_hooks_on_drift": schema.BoolAttribute{
			MarkdownDescription: "Run pre-init hooks on drift detection.",
			Optional:            true,
			Computed:            true,
		},
	},
}

var wfStepsConfig = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Step name.",
			Required:            true,
		},
		"environment_variables": schema.ListNestedAttribute{
			MarkdownDescription: "Environment variables for the step.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"config": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{

							"var_name": schema.StringAttribute{
								MarkdownDescription: "Variable name.",
								Optional:            true,
							},
							"secret_id": schema.StringAttribute{
								MarkdownDescription: "Secret ID.",
								Optional:            true,
							},
							"text_value": schema.StringAttribute{
								MarkdownDescription: "Text value.",
								Optional:            true,
							},
						},
					},
					"kind": schema.StringAttribute{
						MarkdownDescription: "Variable kind (TEXT, SECRET_REF).",
						Optional:            true,
					},
				},
			},
		},
		"approval": schema.BoolAttribute{
			MarkdownDescription: "Require approval for this step.",
			Optional:            true,
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: "Step timeout in seconds.",
			Optional:            true,
		},
		"cmd_override": schema.StringAttribute{
			MarkdownDescription: "Override command for the step (JSON).",
			Optional:            true,
		},
		"mount_points": schema.ListNestedAttribute{
			MarkdownDescription: "Mount points for the step.",
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"source": schema.StringAttribute{
						MarkdownDescription: "Source path.",
						Optional:            true,
					},
					"target": schema.StringAttribute{
						MarkdownDescription: "Destination path.",
						Optional:            true,
					},
					"read_only": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"wf_step_template_id": schema.StringAttribute{
			MarkdownDescription: "Workflow step template ID.",
			Required:            true,
		},
		"wf_step_input_data": schema.SingleNestedAttribute{
			MarkdownDescription: "Workflow step input data.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"schema_type": schema.StringAttribute{
					MarkdownDescription: "Schema type.",
					Optional:            true,
				},
				"data": schema.StringAttribute{
					MarkdownDescription: "Input data (JSON).",
					Optional:            true,
				},
			},
		},
	},
}

// Schema defines the schema for the resource.
func (r *workflowTemplateRevisionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow template revision resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID of the parent workflow template.",
				Required:            true,
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow template revision"),
				Optional:            true,
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionAlias,
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionNotes,
				Optional:            true,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (TERRAFORM, OPENTOFU, ANSIBLE_PLAYBOOK, HELM, KUBECTL, CLOUDFORMATION, CUSTOM).",
				Required:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionIsPublic,
				Optional:            true,
				Computed:            true,
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.TemplateRevisionDeprecation,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecationEffectiveDate,
						Optional:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationMessage,
						Optional:            true,
					},
				},
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: "List of environment variables for the revision.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"config": schema.SingleNestedAttribute{
							MarkdownDescription: "Configuration for the environment variable.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"var_name": schema.StringAttribute{
									MarkdownDescription: "Name of the variable.",
									Required:            true,
								},
								"secret_id": schema.StringAttribute{
									MarkdownDescription: "ID of the secret (if using vault secret).",
									Optional:            true,
								},
								"text_value": schema.StringAttribute{
									MarkdownDescription: "Text value (if using plain text).",
									Optional:            true,
								},
							},
						},
						"kind": schema.StringAttribute{
							MarkdownDescription: "Kind of the environment variable (TEXT, SECRET_REF).",
							Required:            true,
						},
					},
				},
			},
			"input_schemas": schema.ListNestedAttribute{
				MarkdownDescription: "List of input schemas for the revision.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the schema.",
							Optional:            true,
						},
						"encoded_data": schema.StringAttribute{
							MarkdownDescription: "Encoded schema data.",
							Optional:            true,
						},
						"ui_schema_data": schema.StringAttribute{
							MarkdownDescription: "UI schema data (JSON).",
							Optional:            true,
						},
					},
				},
			},
			"mini_steps": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"notifications": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"email": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"approval_required": ministepsNotificationRecepients,
									"cancelled":         ministepsNotificationRecepients,
									"completed":         ministepsNotificationRecepients,
									"drift_detected":    ministepsNotificationRecepients,
									"errored":           ministepsNotificationRecepients,
								},
							},
						},
					},
					"webhooks": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"approval_required": ministepsWebhooks,
							"cancelled":         ministepsWebhooks,
							"completed":         ministepsWebhooks,
							"drift_detected":    ministepsWebhooks,
							"errored":           ministepsWebhooks,
						},
					},
					"wf_chaining": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"completed": ministepsWorkflowChaining,
							"errored":   ministepsWorkflowChaining,
						},
					},
				},
			},
			"runner_constraints": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
					},
					"names": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow template revision"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: "Scheduled runs for the workflow template revision.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cron": schema.StringAttribute{
							MarkdownDescription: "Cron expression defining the schedule.",
							Required:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "State of the schedule (ENABLED, DISABLED).",
							Optional:            true,
						},
						"desc": schema.StringAttribute{
							MarkdownDescription: "Description of the schedule.",
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the schedule.",
							Optional:            true,
						},
						"inputs": schema.SingleNestedAttribute{
							MarkdownDescription: "Input overrides applied when the schedule triggers a run.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"context_tags": schema.MapAttribute{
									MarkdownDescription: "Context tags to set for the triggered run.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"enable_chaining": schema.BoolAttribute{
									MarkdownDescription: "Enable workflow chaining for the triggered run.",
									Optional:            true,
								},
								"environment_variables": schema.ListNestedAttribute{
									MarkdownDescription: "Environment variable overrides for the triggered run.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"config": schema.SingleNestedAttribute{
												Required: true,
												Attributes: map[string]schema.Attribute{
													"var_name": schema.StringAttribute{
														Optional: true,
													},
													"secret_id": schema.StringAttribute{
														Optional: true,
													},
													"text_value": schema.StringAttribute{
														Optional: true,
													},
												},
											},
											"kind": schema.StringAttribute{
												Optional: true,
											},
										},
									},
								},
								"mini_steps": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"notifications": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"email": schema.SingleNestedAttribute{
													Optional: true,
													Attributes: map[string]schema.Attribute{
														"approval_required": ministepsNotificationRecepients,
														"cancelled":         ministepsNotificationRecepients,
														"completed":         ministepsNotificationRecepients,
														"drift_detected":    ministepsNotificationRecepients,
														"errored":           ministepsNotificationRecepients,
													},
												},
											},
										},
										"webhooks": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"approval_required": ministepsWebhooks,
												"cancelled":         ministepsWebhooks,
												"completed":         ministepsWebhooks,
												"drift_detected":    ministepsWebhooks,
												"errored":           ministepsWebhooks,
											},
										},
										"wf_chaining": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"completed": ministepsWorkflowChaining,
												"errored":   ministepsWorkflowChaining,
											},
										},
									},
								},
								"scheduled_at": schema.StringAttribute{
									MarkdownDescription: "Override for the scheduled run time.",
									Optional:            true,
								},
								"terraform_action": schema.SingleNestedAttribute{
									MarkdownDescription: "Terraform action to perform (plan, apply, destroy).",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"action": schema.StringAttribute{
											MarkdownDescription: "Terraform action enum value.",
											Optional:            true,
										},
									},
								},
								"terraform_config": terraformConfigSchema,
								"user_job_cpu": schema.Int64Attribute{
									MarkdownDescription: "Override for user job CPU.",
									Optional:            true,
								},
								"user_job_memory": schema.Int64Attribute{
									MarkdownDescription: "Override for user job memory.",
									Optional:            true,
								},
								"vcs_config": schema.SingleNestedAttribute{
									MarkdownDescription: "VCS configuration override for the triggered run.",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"iac_vcs_config": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"use_marketplace_template": schema.BoolAttribute{
													Optional: true,
												},
												"iac_template_id": schema.StringAttribute{
													Optional: true,
												},
												"custom_source": schema.SingleNestedAttribute{
													Optional: true,
													Attributes: map[string]schema.Attribute{
														"source_config_dest_kind": schema.StringAttribute{
															Optional: true,
														},
														"config": schema.SingleNestedAttribute{
															Optional: true,
															Attributes: map[string]schema.Attribute{
																"is_private": schema.BoolAttribute{
																	Optional: true,
																},
																"auth": schema.StringAttribute{
																	Optional: true,
																},
																"git_core_auto_crlf": schema.BoolAttribute{
																	Optional: true,
																},
																"git_sparse_checkout_config": schema.StringAttribute{
																	Optional: true,
																},
																"include_sub_module": schema.BoolAttribute{
																	Optional: true,
																},
																"ref": schema.StringAttribute{
																	Optional: true,
																},
																"repo": schema.StringAttribute{
																	Optional: true,
																},
																"working_dir": schema.StringAttribute{
																	Optional: true,
																},
															},
														},
													},
												},
											},
										},
										"iac_input_data": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"schema_id": schema.StringAttribute{
													Optional: true,
												},
												"schema_type": schema.StringAttribute{
													Optional: true,
												},
												"data": schema.StringAttribute{
													MarkdownDescription: "Input data as a JSON string.",
													Optional:            true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: "Context tags for the revision.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: "List of approvers for the revision.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"number_of_approvals_required": schema.Int64Attribute{
				MarkdownDescription: "Number of approvals required.",
				Optional:            true,
				Computed:            true,
			},
			"user_job_cpu": schema.Int64Attribute{
				MarkdownDescription: "User job CPU.",
				Optional:            true,
				Computed:            true,
			},
			"user_job_memory": schema.Int64Attribute{
				MarkdownDescription: "User job memory.",
				Optional:            true,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the revision.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: "Source config destination kind.",
						Optional:            true,
						Computed:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: "Source configuration details.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"is_private": schema.BoolAttribute{
								MarkdownDescription: "Whether the source is private.",
								Optional:            true,
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: "Authentication method.",
								Optional:            true,
							},
							"git_core_auto_crlf": schema.BoolAttribute{
								MarkdownDescription: "Git core auto CRLF setting.",
								Optional:            true,
							},
							"git_sparse_checkout_config": schema.StringAttribute{
								MarkdownDescription: "Git sparse checkout configuration.",
								Optional:            true,
							},
							"include_sub_module": schema.BoolAttribute{
								MarkdownDescription: "Whether to include submodules.",
								Optional:            true,
							},
							"ref": schema.StringAttribute{
								MarkdownDescription: "Git reference (branch, tag, commit).",
								Optional:            true,
							},
							"repo": schema.StringAttribute{
								MarkdownDescription: "Repository URL.",
								Optional:            true,
							},
							"working_dir": schema.StringAttribute{
								MarkdownDescription: "Working directory within the repository.",
								Optional:            true,
							},
						},
					},
				},
			},
			"terraform_config": terraformConfigSchema,
			"deployment_platform_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Deployment platform configuration for the revision.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: "Deployment platform kind (AWS_STATIC, AWS_RBAC, AWS_OIDC, AZURE_STATIC, AZURE_OIDC, GCP_STATIC, GCP_OIDC).",
						Required:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: "Deployment platform configuration details.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"integration_id": schema.StringAttribute{
								MarkdownDescription: "Integration ID for the deployment platform.",
								Required:            true,
							},
							"profile_name": schema.StringAttribute{
								MarkdownDescription: "Profile name for the deployment platform.",
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: "Workflow steps configuration.",
				Optional:            true,
				Computed:            true,
				NestedObject:        wfStepsConfig,
			},
		},
	}
}
