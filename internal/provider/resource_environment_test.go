package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccEnvironmentResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read.
			{
				Config: testAccEnvironmentResourceConfig("test-env-app", "staging", "A test environment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("admiral_environment.test", "id"),
					resource.TestCheckResourceAttrPair("admiral_environment.test", "application_id", "admiral_application.test", "id"),
					resource.TestCheckResourceAttr("admiral_environment.test", "name", "staging"),
					resource.TestCheckResourceAttr("admiral_environment.test", "description", "A test environment"),
				),
			},
			// ImportState.
			{
				ResourceName:      "admiral_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEnvironmentResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccEnvironmentResourceConfig("test-env-app-update", "staging", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "description", "Initial description"),
				),
			},
			// Update description and rename in place: the ID must survive.
			{
				Config: testAccEnvironmentResourceConfig("test-env-app-update", "staging-2", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "name", "staging-2"),
					resource.TestCheckResourceAttr("admiral_environment.test", "description", "Updated description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("admiral_environment.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Drop the description: it must be cleared, not kept.
			{
				Config: testAccEnvironmentResourceConfig("test-env-app-update", "staging-2", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("admiral_environment.test", "description"),
				),
			},
		},
	})
}

func TestAccEnvironmentResource_labels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentResourceConfigWithLabels("test-env-app-labels", "staging", map[string]string{
					"tier": "critical",
					"team": "platform",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "labels.tier", "critical"),
					resource.TestCheckResourceAttr("admiral_environment.test", "labels.team", "platform"),
				),
			},
			{
				Config: testAccEnvironmentResourceConfigWithLabels("test-env-app-labels", "staging", map[string]string{
					"tier": "standard",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "labels.tier", "standard"),
					resource.TestCheckNoResourceAttr("admiral_environment.test", "labels.team"),
				),
			},
			{
				Config: testAccEnvironmentResourceConfig("test-env-app-labels", "staging", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("admiral_environment.test", "labels.%"),
				),
			},
		},
	})
}

// See TestAccApplicationResource_emptyValues.
func TestAccEnvironmentResource_emptyValues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentResourceConfigEmptyValues("test-app-env-empty", "staging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "description", ""),
					resource.TestCheckResourceAttr("admiral_environment.test", "labels.%", "0"),
				),
			},
			// Update into the empty shape from a populated one.
			{
				Config: testAccEnvironmentResourceConfigWithLabels("test-app-env-empty", "staging", map[string]string{"env": "staging"}),
			},
			{
				Config: testAccEnvironmentResourceConfigEmptyValues("test-app-env-empty", "staging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("admiral_environment.test", "description", ""),
					resource.TestCheckResourceAttr("admiral_environment.test", "labels.%", "0"),
				),
			},
		},
	})
}

func testAccEnvironmentResourceConfig(appName, envName, description string) string {
	desc := ""
	if description != "" {
		desc = fmt.Sprintf("  description    = %q\n", description)
	}

	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

resource "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = %[2]q
%[3]s}
`, appName, envName, desc)
}

func testAccEnvironmentResourceConfigWithLabels(appName, envName string, labels map[string]string) string {
	labelEntries := ""
	for k, v := range labels {
		labelEntries += fmt.Sprintf("    %q = %q\n", k, v)
	}

	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

resource "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = %[2]q
  labels = {
%[3]s  }
}
`, appName, envName, labelEntries)
}

func testAccEnvironmentResourceConfigEmptyValues(appName, envName string) string {
	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

resource "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = %[2]q
  description    = ""
  labels         = {}
}
`, appName, envName)
}
