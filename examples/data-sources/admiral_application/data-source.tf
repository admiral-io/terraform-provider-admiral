# Look up an application by name
data "admiral_application" "example" {
  name = "my-app"
}

# Look up an application by ID
data "admiral_application" "by_id" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}