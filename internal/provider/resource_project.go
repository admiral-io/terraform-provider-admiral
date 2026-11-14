package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/sdk/client"
	projectv1 "go.admiral.io/sdk/proto/admiral/api/project/v1"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

type projectResource struct {
	client client.AdmiralClient
}

type projectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Labels      types.Map    `tfsdk:"labels"`
}

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Admiral project. A project is the unit a team owns, and it holds the environments its components are deployed to.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the project. Must be URL-safe: lowercase alphanumeric and hyphens, 1-63 characters.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the project.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Key-value labels for the project.",
			},
		},
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	project := &projectv1.Project{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}
	createReq := &projectv1.CreateProjectRequest{Project: project}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		project.Labels = labels
	}

	// Call API
	result, err := r.client.Project().CreateProject(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Project",
			"Could not create project: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapProjectToState(ctx, result.Project, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Project().GetProject(ctx, &projectv1.GetProjectRequest{
		ProjectId: state.ID.ValueString(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read project ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.mapProjectToState(ctx, result.Project, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The plan is the whole desired state, so every mutable field goes in
	// the mask: a description or labels removed from the configuration is
	// cleared on the server rather than left as it was.
	project := &projectv1.Project{
		Id:          state.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		project.Labels = labels
	}

	result, err := r.client.Project().UpdateProject(ctx, &projectv1.UpdateProjectRequest{
		Project:    project,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "labels"}},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Project",
			"Could not update project: "+err.Error(),
		)
		return
	}

	r.mapProjectToState(ctx, result.Project, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Project().DeleteProject(ctx, &projectv1.DeleteProjectRequest{
		ProjectId: state.ID.ValueString(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		resp.Diagnostics.AddError(
			"Error Deleting Project",
			"Could not delete project: "+err.Error(),
		)
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectResource) mapProjectToState(ctx context.Context, project *projectv1.Project, model *projectResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(project.Id)
	model.Name = types.StringValue(project.Name)

	model.Description = flattenDescription(project.Description, model.Description)
	model.Labels = flattenLabels(ctx, project.Labels, model.Labels, diags)
}
