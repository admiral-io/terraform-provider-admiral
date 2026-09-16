package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API does not distinguish an absent description or label map from an
// empty one, but Terraform does: `description = ""` and `labels = {}` plan a
// known empty value, and the applied state must match the plan exactly. An
// empty server value therefore keeps the empty shape when the prior value
// (plan on create/update, state on read) already had it, and is null
// otherwise so an omitted attribute stays omitted.

func flattenDescription(description string, prior types.String) types.String {
	if description != "" {
		return types.StringValue(description)
	}
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() == "" {
		return prior
	}
	return types.StringNull()
}

func flattenLabels(ctx context.Context, labels map[string]string, prior types.Map, diags *diag.Diagnostics) types.Map {
	if len(labels) > 0 {
		m, d := types.MapValueFrom(ctx, types.StringType, labels)
		diags.Append(d...)
		return m
	}
	if !prior.IsNull() && !prior.IsUnknown() && len(prior.Elements()) == 0 {
		return prior
	}
	return types.MapNull(types.StringType)
}
