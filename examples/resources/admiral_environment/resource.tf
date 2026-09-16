resource "admiral_application" "example" {
  name = "my-app"
}

resource "admiral_environment" "production" {
  application_id = admiral_application.example.id
  name           = "production"
  description    = "Customer-facing environment"

  labels = {
    tier = "critical"
  }
}
