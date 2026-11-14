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
	projectv1 "go.admiral.io/sdk/proto/admiral/api/project/v1"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

type projectDataSource struct {
	client client.AdmiralClient
}

type projectDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Labels      types.Map    `tfsdk:"labels"`
}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read an Admiral project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the project. Exactly one of `id` or `name` must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The name of the project. Exactly one of `id` or `name` must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the project.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Key-value labels for the project.",
			},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectDataSourceModel
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

	var project *projectv1.Project

	if hasID {
		result, err := d.client.Project().GetProject(ctx, &projectv1.GetProjectRequest{
			ProjectId: config.ID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Project",
				"Could not read project ID "+config.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		project = result.Project
	} else {
		name := config.Name.ValueString()
		filter, err := filterEq("name", name)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Invalid Project Name",
				err.Error(),
			)
			return
		}

		// Names are not unique server-side, so ask for two: a second match
		// means the lookup is ambiguous and the caller must pass an ID.
		result, err := d.client.Project().ListProjects(ctx, &projectv1.ListProjectsRequest{
			Filter:   filter,
			PageSize: 2,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Project",
				"Could not find project with name "+name+": "+err.Error(),
			)
			return
		}
		switch len(result.Projects) {
		case 0:
			resp.Diagnostics.AddError(
				"Project Not Found",
				"No project found with name "+name+".",
			)
			return
		case 1:
			project = result.Projects[0]
		default:
			ids := make([]string, 0, len(result.Projects))
			for _, p := range result.Projects {
				ids = append(ids, p.Id)
			}
			resp.Diagnostics.AddAttributeError(
				path.Root("name"),
				"Ambiguous Project Name",
				fmt.Sprintf("More than one project is named %q. Set `id` instead; candidates: %s.", name, strings.Join(ids, ", ")),
			)
			return
		}
	}

	d.mapProjectToState(ctx, project, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *projectDataSource) mapProjectToState(ctx context.Context, project *projectv1.Project, model *projectDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(project.Id)
	model.Name = types.StringValue(project.Name)

	if project.Description != "" {
		model.Description = types.StringValue(project.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(project.Labels) > 0 {
		labelsMap, d := types.MapValueFrom(ctx, types.StringType, project.Labels)
		diags.Append(d...)
		model.Labels = labelsMap
	} else {
		model.Labels = types.MapNull(types.StringType)
	}
}
