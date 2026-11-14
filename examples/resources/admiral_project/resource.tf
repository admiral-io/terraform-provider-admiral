resource "admiral_project" "example" {
  name        = "my-project"
  description = "An example project managed by Terraform"

  labels = {
    team = "platform"
    tier = "critical"
  }
}
