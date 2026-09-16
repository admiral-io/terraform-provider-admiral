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
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

var (
	_ resource.Resource                = &environmentResource{}
	_ resource.ResourceWithConfigure   = &environmentResource{}
	_ resource.ResourceWithImportState = &environmentResource{}
)

type environmentResource struct {
	client client.AdmiralClient
}

type environmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Labels        types.Map    `tfsdk:"labels"`
}

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Admiral environment. An environment belongs to an application and is where its components are deployed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the application this environment belongs to. Changing it replaces the environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the environment, unique within the application. Must be URL-safe: lowercase alphanumeric and hyphens, starting with a letter and ending with an alphanumeric character, 1-63 characters.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the environment.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Key-value labels for the environment.",
			},
		},
	}
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &environmentv1.CreateEnvironmentRequest{
		ApplicationId: plan.ApplicationID.ValueString(),
		Name:          plan.Name.ValueString(),
		Description:   plan.Description.ValueString(),
	}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Labels = labels
	}

	result, err := r.client.Environment().CreateEnvironment(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Environment",
			"Could not create environment: "+err.Error(),
		)
		return
	}

	r.mapEnvironmentToState(ctx, result.Environment, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Environment().GetEnvironment(ctx, &environmentv1.GetEnvironmentRequest{
		EnvironmentId: state.ID.ValueString(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Environment",
			"Could not read environment ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.mapEnvironmentToState(ctx, result.Environment, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The plan is the whole desired state, so every mutable field goes in
	// the mask: a description or labels removed from the configuration is
	// cleared on the server rather than left as it was.
	env := &environmentv1.Environment{
		Id:            state.ID.ValueString(),
		ApplicationId: state.ApplicationID.ValueString(),
		Name:          plan.Name.ValueString(),
		Description:   plan.Description.ValueString(),
	}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		env.Labels = labels
	}

	result, err := r.client.Environment().UpdateEnvironment(ctx, &environmentv1.UpdateEnvironmentRequest{
		Environment: env,
		UpdateMask:  &fieldmaskpb.FieldMask{Paths: []string{"name", "description", "labels"}},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Environment",
			"Could not update environment: "+err.Error(),
		)
		return
	}

	r.mapEnvironmentToState(ctx, result.Environment, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Environment().DeleteEnvironment(ctx, &environmentv1.DeleteEnvironmentRequest{
		EnvironmentId: state.ID.ValueString(),
	})
	if err != nil && status.Code(err) != codes.NotFound {
		resp.Diagnostics.AddError(
			"Error Deleting Environment",
			"Could not delete environment: "+err.Error(),
		)
		return
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *environmentResource) mapEnvironmentToState(ctx context.Context, env *environmentv1.Environment, model *environmentResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(env.Id)
	model.ApplicationID = types.StringValue(env.ApplicationId)
	model.Name = types.StringValue(env.Name)

	model.Description = flattenDescription(env.Description, model.Description)
	model.Labels = flattenLabels(ctx, env.Labels, model.Labels, diags)
}
