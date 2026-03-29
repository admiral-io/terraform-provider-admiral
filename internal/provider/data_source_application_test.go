package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccApplicationDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfigByName("test-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.admiral_application.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_application.test", "name", "test-ds-name"),
				),
			},
		},
	})
}

func TestAccApplicationDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfigByID("test-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.admiral_application.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_application.test", "name", "test-ds-id"),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfigByName(name string) string {
	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

data "admiral_application" "test" {
  name = admiral_application.test.name
}
`, name)
}

func testAccApplicationDataSourceConfigByID(name string) string {
	return fmt.Sprintf(`
resource "admiral_application" "test" {
  name = %[1]q
}

data "admiral_application" "test" {
  id = admiral_application.test.id
}
`, name)
}
