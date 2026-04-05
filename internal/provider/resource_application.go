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

	"go.admiral.io/sdk/client"
	applicationv1 "go.admiral.io/sdk/proto/admiral/application/v1"
)

var (
	_ resource.Resource                = &applicationResource{}
	_ resource.ResourceWithConfigure   = &applicationResource{}
	_ resource.ResourceWithImportState = &applicationResource{}
)

type applicationResource struct {
	client client.AdmiralClient
}

type applicationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Labels      types.Map    `tfsdk:"labels"`
}

func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Admiral application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the application. Must be URL-safe: lowercase alphanumeric and hyphens, 1-63 characters.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the application.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Key-value labels for the application.",
			},
		},
	}
}

func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	createReq := &applicationv1.CreateApplicationRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		createReq.Description = &desc
	}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Labels = labels
	}

	// Call API
	result, err := r.client.Application().CreateApplication(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Application",
			"Could not create application: "+err.Error(),
		)
		return
	}

	// Map response to state
	r.mapApplicationToState(ctx, result.Application, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Application().GetApplication(ctx, &applicationv1.GetApplicationRequest{
		ApplicationId: state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.mapApplicationToState(ctx, result.Application, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the application object for the update
	app := &applicationv1.Application{
		Id:   state.ID.ValueString(),
		Name: plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() {
		app.Description = plan.Description.ValueString()
	}

	if !plan.Labels.IsNull() {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.Labels = labels
	}

	result, err := r.client.Application().UpdateApplication(ctx, &applicationv1.UpdateApplicationRequest{
		Application: app,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Application",
			"Could not update application: "+err.Error(),
		)
		return
	}

	r.mapApplicationToState(ctx, result.Application, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Application().DeleteApplication(ctx, &applicationv1.DeleteApplicationRequest{
		ApplicationId: state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Application",
			"Could not delete application: "+err.Error(),
		)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *applicationResource) mapApplicationToState(ctx context.Context, app *applicationv1.Application, model *applicationResourceModel, diags *diag.Diagnostics) {
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
