package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		prior types.String
		want  types.String
	}{
		{name: "server value wins", value: "desc", prior: types.StringNull(), want: types.StringValue("desc")},
		{name: "server value replaces prior", value: "new", prior: types.StringValue("old"), want: types.StringValue("new")},
		{name: "empty with null prior stays null", value: "", prior: types.StringNull(), want: types.StringNull()},
		{name: "empty with unknown prior becomes null", value: "", prior: types.StringUnknown(), want: types.StringNull()},
		{name: "empty keeps planned empty string", value: "", prior: types.StringValue(""), want: types.StringValue("")},
		{name: "empty with non-empty prior becomes null", value: "", prior: types.StringValue("gone"), want: types.StringNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := flattenDescription(tt.value, tt.prior); !got.Equal(tt.want) {
				t.Fatalf("flattenDescription(%q, %v) = %v, want %v", tt.value, tt.prior, got, tt.want)
			}
		})
	}
}

func TestFlattenLabels(t *testing.T) {
	t.Parallel()

	empty := types.MapValueMust(types.StringType, map[string]attr.Value{})
	one := types.MapValueMust(types.StringType, map[string]attr.Value{"env": types.StringValue("prod")})

	tests := []struct {
		name   string
		labels map[string]string
		prior  types.Map
		want   types.Map
	}{
		{name: "server labels win", labels: map[string]string{"env": "prod"}, prior: types.MapNull(types.StringType), want: one},
		{name: "server labels replace planned empty map", labels: map[string]string{"env": "prod"}, prior: empty, want: one},
		{name: "empty with null prior stays null", labels: nil, prior: types.MapNull(types.StringType), want: types.MapNull(types.StringType)},
		{name: "empty with unknown prior becomes null", labels: nil, prior: types.MapUnknown(types.StringType), want: types.MapNull(types.StringType)},
		{name: "empty keeps planned empty map", labels: map[string]string{}, prior: empty, want: empty},
		{name: "empty with non-empty prior becomes null", labels: nil, prior: one, want: types.MapNull(types.StringType)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			got := flattenLabels(context.Background(), tt.labels, tt.prior, &diags)
			if diags.HasError() {
				t.Fatalf("flattenLabels returned diagnostics: %v", diags)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("flattenLabels(%v, %v) = %v, want %v", tt.labels, tt.prior, got, tt.want)
			}
		})
	}
}
