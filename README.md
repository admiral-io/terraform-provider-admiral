# Terraform Provider for Admiral

The Admiral Terraform provider allows you to manage [Admiral](https://admiral.io) platform resources using infrastructure as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.26 (to build the provider)

## Usage

```hcl
terraform {
  required_providers {
    admiral = {
      source  = "admiral/admiral"
      version = "~> 1"
    }
  }
}

provider "admiral" {
  # host  = "api.admiral.io"    # optional, defaults to api.admiral.io:443
  # token = "..."               # or set ADMIRAL_TOKEN env var
}

resource "admiral_application" "my_app" {
  name        = "my-app"
  description = "My application"

  labels = {
    team = "platform"
  }
}

data "admiral_application" "other_app" {
  name = "other-app"
}
```

## Authentication

The provider requires an Admiral API token. You can provide it in one of two ways:

- Set the `ADMIRAL_TOKEN` environment variable (recommended)
- Set the `token` attribute in the provider configuration block

```shell
export ADMIRAL_TOKEN="your-api-token"
```

## Documentation

Full provider documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/admiral/admiral/latest/docs).

## Developing the Provider

### Building

```shell
make build
```

### Running Tests

```shell
make test
```

### Running Acceptance Tests

Acceptance tests run against a real Admiral instance and require valid credentials.

```shell
export ADMIRAL_TOKEN="your-api-token"
make testacc
```

### Generating Documentation

Documentation is generated from provider schemas and example files using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).

```shell
make generate
```

### Linting

```shell
make lint
```

### All Available Targets

```shell
make help
```
