# Look up an environment by name within a project
data "admiral_project" "example" {
  name = "my-project"
}

data "admiral_environment" "production" {
  project_id = data.admiral_project.example.id
  name       = "production"
}

# Look up an environment by ID
data "admiral_environment" "by_id" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}
