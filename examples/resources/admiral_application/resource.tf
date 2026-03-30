resource "admiral_application" "example" {
  name        = "my-app"
  description = "An example application managed by Terraform"

  labels = {
    team = "platform"
    tier = "critical"
  }
}
