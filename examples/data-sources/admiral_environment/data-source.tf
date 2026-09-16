# Look up an environment by name within an application
data "admiral_application" "example" {
  name = "my-app"
}

data "admiral_environment" "production" {
  application_id = data.admiral_application.example.id
  name           = "production"
}

# Look up an environment by ID
data "admiral_environment" "by_id" {
  id = "550e8400-e29b-41d4-a716-446655440000"
}
