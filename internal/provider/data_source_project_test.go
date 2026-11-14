package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectDataSource_byName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfigByName("test-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.admiral_project.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_project.test", "name", "test-ds-name"),
				),
			},
		},
	})
}

func TestAccProjectDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfigByID("test-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.admiral_project.test", "id"),
					resource.TestCheckResourceAttr("data.admiral_project.test", "name", "test-ds-id"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfigByName(name string) string {
	return fmt.Sprintf(`
resource "admiral_project" "test" {
  name = %[1]q
}

data "admiral_project" "test" {
  name = admiral_project.test.name
}
`, name)
}

func testAccProjectDataSourceConfigByID(name string) string {
	return fmt.Sprintf(`
resource "admiral_project" "test" {
  name = %[1]q
}

data "admiral_project" "test" {
  id = admiral_project.test.id
}
`, name)
}
