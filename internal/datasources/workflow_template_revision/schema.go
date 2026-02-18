package workflowtemplaterevision

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ministepsNotificationRecepients = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	},
}

var ministepsWebhooks = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name": schema.StringAttribute{
				Computed: true,
			},
			"webhook_url": schema.StringAttribute{
				Computed: true,
			},
			"webhook_secret": schema.StringAttribute{
				Computed: true,
			},
		},
	},
}

var ministepsWorkflowChaining = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				Computed: true,
			},
			"stack_id": schema.StringAttribute{
				Computed: true,
			},
			"stack_run_payload": schema.StringAttribute{
				Computed: true,
			},
			"workflow_id": schema.StringAttribute{
				Computed: true,
			},
			"workflow_run_payload": schema.StringAttribute{
				Computed: true,
			},
		},
	},
}

var miniStepsSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"notifications": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"email": schema.SingleNestedAttribute{
					Computed: true,
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
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"approval_required": ministepsWebhooks,
				"cancelled":         ministepsWebhooks,
				"completed":         ministepsWebhooks,
				"drift_detected":    ministepsWebhooks,
				"errored":           ministepsWebhooks,
			},
		},
		"wf_chaining": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"completed": ministepsWorkflowChaining,
				"errored":   ministepsWorkflowChaining,
			},
		},
	},
}

var environmentVariablesSchema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"config": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"var_name": schema.StringAttribute{
						Computed: true,
					},
					"secret_id": schema.StringAttribute{
						Computed: true,
					},
					"text_value": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"kind": schema.StringAttribute{
				Computed: true,
			},
		},
	},
}

var mountPointsSchema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				Computed: true,
			},
			"target": schema.StringAttribute{
				Computed: true,
			},
			"read_only": schema.BoolAttribute{
				Computed: true,
			},
		},
	},
}

var wfStepsConfigSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Computed: true,
		},
		"wf_step_template_id": schema.StringAttribute{
			Computed: true,
		},
		"timeout": schema.Int64Attribute{
			Computed: true,
		},
		"approval": schema.BoolAttribute{
			Computed: true,
		},
		"cmd_override": schema.StringAttribute{
			Computed: true,
		},
		"environment_variables": environmentVariablesSchema,
		"mount_points":          mountPointsSchema,
		"wf_step_input_data": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"schema_type": schema.StringAttribute{
					Computed: true,
				},
				"data": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	},
}

var terraformConfigSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"terraform_version": schema.StringAttribute{
			Computed: true,
		},
		"terraform_plan_options": schema.StringAttribute{
			Computed: true,
		},
		"terraform_init_options": schema.StringAttribute{
			Computed: true,
		},
		"terraform_bin_path": schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"source": schema.StringAttribute{
						Computed: true,
					},
					"target": schema.StringAttribute{
						Computed: true,
					},
					"read_only": schema.BoolAttribute{
						Computed: true,
					},
				},
			},
		},
		"timeout": schema.Int64Attribute{
			Computed: true,
		},
		"managed_terraform_state": schema.BoolAttribute{
			Computed: true,
		},
		"drift_check": schema.BoolAttribute{
			Computed: true,
		},
		"drift_cron": schema.StringAttribute{
			Computed: true,
		},
		"approval_pre_apply": schema.BoolAttribute{
			Computed: true,
		},
		"run_pre_init_hooks_on_drift": schema.BoolAttribute{
			Computed: true,
		},
		"pre_init_hooks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"pre_plan_hooks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"post_plan_hooks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"pre_apply_hooks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"post_apply_hooks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"post_apply_wf_steps_config": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: wfStepsConfigSchema,
		},
		"pre_apply_wf_steps_config": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: wfStepsConfigSchema,
		},
		"pre_plan_wf_steps_config": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: wfStepsConfigSchema,
		},
		"post_plan_wf_steps_config": schema.ListNestedAttribute{
			Computed:     true,
			NestedObject: wfStepsConfigSchema,
		},
	},
}

var runtimeSourceSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"source_config_dest_kind": schema.StringAttribute{
			Computed: true,
		},
		"config": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"is_private": schema.BoolAttribute{
					Computed: true,
				},
				"auth": schema.StringAttribute{
					Computed: true,
				},
				"git_core_auto_crlf": schema.BoolAttribute{
					Computed: true,
				},
				"git_sparse_checkout_config": schema.StringAttribute{
					Computed: true,
				},
				"include_sub_module": schema.BoolAttribute{
					Computed: true,
				},
				"ref": schema.StringAttribute{
					Computed: true,
				},
				"repo": schema.StringAttribute{
					Computed: true,
				},
				"working_dir": schema.StringAttribute{
					Computed: true,
				},
			},
		},
	},
}

func (d *workflowTemplateRevisionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a workflow template revision.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID of the parent workflow template.",
				Optional:            true,
				Computed:            true,
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow template revision"),
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				Computed: true,
			},
			"notes": schema.StringAttribute{
				Computed: true,
			},
			"source_config_kind": schema.StringAttribute{
				Computed: true,
			},
			"is_public": schema.StringAttribute{
				Computed: true,
			},
			"deprecation": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						Computed: true,
					},
					"message": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"environment_variables": environmentVariablesSchema,
			"input_schemas": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed: true,
						},
						"encoded_data": schema.StringAttribute{
							Computed: true,
						},
						"ui_schema_data": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"mini_steps":          miniStepsSchema,
			"runner_constraints": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"names": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow template revision"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"user_schedules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cron": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"desc": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"inputs": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"context_tags": schema.MapAttribute{
									Computed:    true,
									ElementType: types.StringType,
								},
								"enable_chaining": schema.BoolAttribute{
									Computed: true,
								},
								"environment_variables": environmentVariablesSchema,
								"mini_steps":            miniStepsSchema,
								"scheduled_at": schema.StringAttribute{
									Computed: true,
								},
								"terraform_action": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"action": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								"terraform_config": terraformConfigSchema,
								"user_job_cpu": schema.Int64Attribute{
									Computed: true,
								},
								"user_job_memory": schema.Int64Attribute{
									Computed: true,
								},
								"vcs_config": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"iac_vcs_config": schema.SingleNestedAttribute{
											Computed: true,
											Attributes: map[string]schema.Attribute{
												"use_marketplace_template": schema.BoolAttribute{
													Computed: true,
												},
												"iac_template_id": schema.StringAttribute{
													Computed: true,
												},
												"custom_source": runtimeSourceSchema,
											},
										},
										"iac_input_data": schema.SingleNestedAttribute{
											Computed: true,
											Attributes: map[string]schema.Attribute{
												"schema_id": schema.StringAttribute{
													Computed: true,
												},
												"schema_type": schema.StringAttribute{
													Computed: true,
												},
												"data": schema.StringAttribute{
													Computed: true,
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
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"number_of_approvals_required": schema.Int64Attribute{
				Computed: true,
			},
			"user_job_cpu": schema.Int64Attribute{
				Computed: true,
			},
			"user_job_memory": schema.Int64Attribute{
				Computed: true,
			},
			"runtime_source":  runtimeSourceSchema,
			"terraform_config": terraformConfigSchema,
			"deployment_platform_config": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						Computed: true,
					},
					"config": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"integration_id": schema.StringAttribute{
								Computed: true,
							},
							"profile_name": schema.StringAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				Computed:     true,
				NestedObject: wfStepsConfigSchema,
			},
		},
	}
}
