package runnergrouptoken

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *runnerGroupTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"runner_group_id": schema.StringAttribute{
				MarkdownDescription: constants.RunnerGroupId,
				Required:            true,
			},
			"runner_group_token": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ApiToken),
				Computed:            true,
			},
		},
	}
}

type runnerGroupTokenModel struct {
	RunnerGroupID    types.String `tfsdk:"runner_group_id"`
	RunnerGroupToken types.String `tfsdk:"runner_group_token"`
}

func buildAPIModelToRunnerGroupTokenModel(apiResponse *apiResponseModel) (*runnerGroupTokenModel, diag.Diagnostics) {
	runnerGroupTokenModelValue := &runnerGroupTokenModel{
		RunnerGroupToken: flatteners.String(apiResponse.Data),
	}
	return runnerGroupTokenModelValue, nil
}
