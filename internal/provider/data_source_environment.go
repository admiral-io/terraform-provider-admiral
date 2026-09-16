package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.admiral.io/sdk/client"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

var (
	_ datasource.DataSource              = &environmentDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentDataSource{}
)

type environmentDataSource struct {
	client client.AdmiralClient
}

type environmentDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Labels        types.Map    `tfsdk:"labels"`
}

func NewEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read an Admiral environment, either by `id` or by `application_id` and `name`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the environment. Either `id`, or both `application_id` and `name`, must be specified.",
			},
			"application_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the application the environment belongs to. Required with `name`; names are only unique within an application.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the environment. Requires `application_id`.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the environment.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Key-value labels for the environment.",
			},
		},
	}
}

func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config environmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	known := func(v types.String) bool { return !v.IsNull() && !v.IsUnknown() }
	hasID := known(config.ID)
	hasApp := known(config.ApplicationID)
	hasName := known(config.Name)

	switch {
	case hasID && (hasApp || hasName):
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Specify either `id`, or `application_id` and `name`, not both.",
		)
		return
	case !hasID && (!hasApp || !hasName):
		resp.Diagnostics.AddError(
			"Missing Attribute",
			"Specify either `id`, or both `application_id` and `name`.",
		)
		return
	}

	var env *environmentv1.Environment

	if hasID {
		result, err := d.client.Environment().GetEnvironment(ctx, &environmentv1.GetEnvironmentRequest{
			EnvironmentId: config.ID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Environment",
				"Could not read environment ID "+config.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		env = result.Environment
	} else {
		appID := config.ApplicationID.ValueString()
		name := config.Name.ValueString()

		byApp, err := filterEq("application_id", appID)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("application_id"), "Invalid Application ID", err.Error())
			return
		}
		byName, err := filterEq("name", name)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("name"), "Invalid Environment Name", err.Error())
			return
		}

		// Names are unique within an application, so with both predicates a
		// second match should not happen; it is still reported rather than
		// silently picking one.
		result, err := d.client.Environment().ListEnvironments(ctx, &environmentv1.ListEnvironmentsRequest{
			Filter:   byApp + " AND " + byName,
			PageSize: 2,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Environment",
				"Could not find environment "+name+" in application "+appID+": "+err.Error(),
			)
			return
		}
		switch len(result.Environments) {
		case 0:
			resp.Diagnostics.AddError(
				"Environment Not Found",
				"No environment named "+name+" in application "+appID+".",
			)
			return
		case 1:
			env = result.Environments[0]
		default:
			ids := make([]string, 0, len(result.Environments))
			for _, e := range result.Environments {
				ids = append(ids, e.Id)
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Ambiguous Environment Name",
				fmt.Sprintf("More than one environment is named %q in application %s. Set `id` instead; candidates: %s.", name, appID, strings.Join(ids, ", ")),
			)
			return
		}
	}

	d.mapEnvironmentToState(ctx, env, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *environmentDataSource) mapEnvironmentToState(ctx context.Context, env *environmentv1.Environment, model *environmentDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(env.Id)
	model.ApplicationID = types.StringValue(env.ApplicationId)
	model.Name = types.StringValue(env.Name)

	if env.Description != "" {
		model.Description = types.StringValue(env.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(env.Labels) > 0 {
		labelsMap, d := types.MapValueFrom(ctx, types.StringType, env.Labels)
		diags.Append(d...)
		model.Labels = labelsMap
	} else {
		model.Labels = types.MapNull(types.StringType)
	}
}
