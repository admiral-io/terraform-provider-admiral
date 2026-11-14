# Look up a project by name
data "admiral_project" "example" {
  name = "my-project"
}

# Look up a project by ID
data "admiral_project" "by_id" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}
