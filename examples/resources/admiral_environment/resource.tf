resource "admiral_project" "example" {
  name = "my-project"
}

resource "admiral_environment" "production" {
  project_id  = admiral_project.example.id
  name        = "production"
  description = "Customer-facing environment"

  labels = {
    tier = "critical"
  }
}
