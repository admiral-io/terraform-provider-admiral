package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read.
			{
				Config: testAccProjectResourceConfig("test-project", "A test project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("admiral_project.test", "id"),
					resource.TestCheckResourceAttr("admiral_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("admiral_project.test", "description", "A test project"),
				),
			},
			// ImportState.
			{
				ResourceName:      "admiral_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccProjectResourceConfig("test-project-update", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "name", "test-project-update"),
					resource.TestCheckResourceAttr("admiral_project.test", "description", "Initial description"),
				),
			},
			// Update description.
			{
				Config: testAccProjectResourceConfig("test-project-update", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "name", "test-project-update"),
					resource.TestCheckResourceAttr("admiral_project.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccProjectResource_labels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with labels.
			{
				Config: testAccProjectResourceConfigWithLabels("test-project-labels", map[string]string{
					"env":  "staging",
					"team": "platform",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "labels.env", "staging"),
					resource.TestCheckResourceAttr("admiral_project.test", "labels.team", "platform"),
				),
			},
			// Update labels.
			{
				Config: testAccProjectResourceConfigWithLabels("test-project-labels", map[string]string{
					"env": "production",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "labels.env", "production"),
					resource.TestCheckNoResourceAttr("admiral_project.test", "labels.team"),
				),
			},
			// Remove all labels.
			{
				Config: testAccProjectResourceConfig("test-project-labels", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("admiral_project.test", "labels.%"),
				),
			},
		},
	})
}

// An explicit `labels = {}` or `description = ""` must round-trip as-is:
// the framework rejects an apply whose result differs from the plan, and
// the API returns both as empty.
func TestAccProjectResource_emptyValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfigEmptyValues("test-project-empty"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "description", ""),
					resource.TestCheckResourceAttr("admiral_project.test", "labels.%", "0"),
				),
			},
			// Update into the empty shape from a populated one.
			{
				Config: testAccProjectResourceConfigWithLabels("test-project-empty", map[string]string{"env": "staging"}),
			},
			{
				Config: testAccProjectResourceConfigEmptyValues("test-project-empty"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_project.test", "description", ""),
					resource.TestCheckResourceAttr("admiral_project.test", "labels.%", "0"),
				),
			},
		},
	})
}

func TestAccProjectResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccProjectResourceConfig("test-project-import", "Import test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("admiral_project.test", "id"),
				),
			},
			// Import by ID.
			{
				ResourceName:      "admiral_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectResourceConfig(name, description string) string {
	if description != "" {
		return fmt.Sprintf(`
resource "admiral_project" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
	}

	return fmt.Sprintf(`
resource "admiral_project" "test" {
  name = %[1]q
}
`, name)
}

func testAccProjectResourceConfigWithLabels(name string, labels map[string]string) string {
	labelEntries := ""
	for k, v := range labels {
		labelEntries += fmt.Sprintf("    %q = %q\n", k, v)
	}

	return fmt.Sprintf(`
resource "admiral_project" "test" {
  name = %[1]q
  labels = {
%[2]s  }
}
`, name, labelEntries)
}

func testAccProjectResourceConfigEmptyValues(name string) string {
	return fmt.Sprintf(`
resource "admiral_project" "test" {
  name        = %[1]q
  description = ""
  labels      = {}
}
`, name)
}
