package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfigByName("test-env-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.admiral_environment.test", "id", "admiral_environment.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_environment.test", "name", "staging"),
					resource.TestCheckResourceAttr("data.admiral_environment.test", "description", "Looked up by name"),
				),
			},
		},
	})
}

func TestAccEnvironmentDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfigByID("test-env-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.admiral_environment.test", "id", "admiral_environment.test", "id"),
					resource.TestCheckResourceAttrPair("data.admiral_environment.test", "application_id", "admiral_application.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_environment.test", "name", "staging"),
				),
			},
		},
	})
}

func testAccEnvironmentDataSourceConfigByName(appName string) string {
	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

resource "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = "staging"
  description    = "Looked up by name"
}

data "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = admiral_environment.test.name
}
`, appName)
}

func testAccEnvironmentDataSourceConfigByID(appName string) string {
	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

resource "admiral_environment" "test" {
  application_id = admiral_application.test.id
  name           = "staging"
}

data "admiral_environment" "test" {
  id = admiral_environment.test.id
}
`, appName)
}
