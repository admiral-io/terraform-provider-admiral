package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read.
			{
				Config: testAccApplicationResourceConfig("test-app", "A test application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("admiral_application.test", "id"),
					resource.TestCheckResourceAttr("admiral_application.test", "name", "test-app"),
					resource.TestCheckResourceAttr("admiral_application.test", "description", "A test application"),
				),
			},
			// ImportState.
			{
				ResourceName:      "admiral_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccApplicationResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccApplicationResourceConfig("test-app-update", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_application.test", "name", "test-app-update"),
					resource.TestCheckResourceAttr("admiral_application.test", "description", "Initial description"),
				),
			},
			// Update description.
			{
				Config: testAccApplicationResourceConfig("test-app-update", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_application.test", "name", "test-app-update"),
					resource.TestCheckResourceAttr("admiral_application.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccApplicationResource_labels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with labels.
			{
				Config: testAccApplicationResourceConfigWithLabels("test-app-labels", map[string]string{
					"env":  "staging",
					"team": "platform",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_application.test", "labels.env", "staging"),
					resource.TestCheckResourceAttr("admiral_application.test", "labels.team", "platform"),
				),
			},
			// Update labels.
			{
				Config: testAccApplicationResourceConfigWithLabels("test-app-labels", map[string]string{
					"env": "production",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_application.test", "labels.env", "production"),
					resource.TestCheckNoResourceAttr("admiral_application.test", "labels.team"),
				),
			},
			// Remove all labels.
			{
				Config: testAccApplicationResourceConfig("test-app-labels", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("admiral_application.test", "labels.%"),
				),
			},
		},
	})
}

func TestAccApplicationResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccApplicationResourceConfig("test-app-import", "Import test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("admiral_application.test", "id"),
				),
			},
			// Import by ID.
			{
				ResourceName:      "admiral_application.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccApplicationResourceConfig(name, description string) string {
	if description != "" {
		return fmt.Sprintf(`
resource "admiral_application" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
	}

	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}
`, name)
}

func testAccApplicationResourceConfigWithLabels(name string, labels map[string]string) string {
	labelEntries := ""
	for k, v := range labels {
		labelEntries += fmt.Sprintf("    %q = %q\n", k, v)
	}

	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
  labels = {
%[2]s  }
}
`, name, labelEntries)
}
