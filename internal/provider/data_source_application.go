package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.admiral.io/sdk/client"
	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
)

var (
	_ datasource.DataSource              = &applicationDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationDataSource{}
)

type applicationDataSource struct {
	client client.AdmiralClient
}

type applicationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Labels      types.Map    `tfsdk:"labels"`
}

func NewApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

func (d *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read an Admiral application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the application. Exactly one of `id` or `name` must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the application. Exactly one of `id` or `name` must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the application.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Key-value labels for the application.",
			},
		},
	}
}

func (d *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config applicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one of id or name is provided.
	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Attribute",
			"Exactly one of `id` or `name` must be specified.",
		)
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Only one of `id` or `name` may be specified, not both.",
		)
		return
	}

	var app *applicationv1.Application

	if hasID {
		result, err := d.client.Application().GetApplication(ctx, &applicationv1.GetApplicationRequest{
			ApplicationId: config.ID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Application",
				"Could not read application ID "+config.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		app = result.Application
	} else {
		result, err := d.client.Application().ListApplications(ctx, &applicationv1.ListApplicationsRequest{
			Filter:   "field['name'] = '" + config.Name.ValueString() + "'",
			PageSize: 1,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Application",
				"Could not find application with name "+config.Name.ValueString()+": "+err.Error(),
			)
			return
		}
		if len(result.Applications) == 0 {
			resp.Diagnostics.AddError(
				"Application Not Found",
				"No application found with name "+config.Name.ValueString()+".",
			)
			return
		}
		app = result.Applications[0]
	}

	d.mapApplicationToState(ctx, app, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *applicationDataSource) mapApplicationToState(ctx context.Context, app *applicationv1.Application, model *applicationDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(app.Id)
	model.Name = types.StringValue(app.Name)

	if app.Description != "" {
		model.Description = types.StringValue(app.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(app.Labels) > 0 {
		labelsMap, d := types.MapValueFrom(ctx, types.StringType, app.Labels)
		diags.Append(d...)
		model.Labels = labelsMap
	} else {
		model.Labels = types.MapNull(types.StringType)
	}
}
