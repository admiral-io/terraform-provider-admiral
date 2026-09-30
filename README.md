# Terraform Provider for Admiral

The Admiral Terraform provider allows you to manage [Admiral](https://admiral.io/?utm_source=github&utm_medium=referral&utm_campaign=terraform-provider-admiral) platform resources using infrastructure as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.26 (to build the provider)

## Usage

```hcl
terraform {
  required_providers {
    admiral = {
      source  = "admiral-io/admiral"
      version = "~> 0.1"
    }
  }
}

provider "admiral" {
  # server  = "api.admiral.io:443"  # optional; or set ADMIRAL_SERVER
  # api_key = "admp_..."            # or set ADMIRAL_API_KEY (recommended)
}

resource "admiral_application" "my_app" {
  name        = "my-app"
  description = "My application"

  labels = {
    team = "platform"
  }
}

resource "admiral_environment" "production" {
  application_id = admiral_application.my_app.id
  name           = "production"
}

data "admiral_application" "other_app" {
  name = "other-app"
}
```

## Authentication

The provider authenticates with an Admiral API key. Provide it in one of two ways:

- Set the `ADMIRAL_API_KEY` environment variable (recommended)
- Set the `api_key` attribute in the provider configuration block

```shell
export ADMIRAL_API_KEY="admp_..."
```

Create a key with the [Admiral CLI](https://github.com/admiral-io/admiral-cli) or in the console.

### Connecting to a local server

`server` accepts any `host:port`. Against a local development server, set `plaintext = true` to skip TLS entirely, or `insecure = true` to use TLS without verifying the certificate. Neither is appropriate for `api.admiral.io`.

## Documentation

Full provider documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/admiral-io/admiral/latest/docs).

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
export ADMIRAL_API_KEY="admp_..."
make testacc
```

### Generating Documentation

Documentation is generated from provider schemas and example files using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs). It needs a `terraform` binary on `PATH`; CI fails if `docs/` is stale.

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

## Admiral

[Admiral](https://admiral.io/?utm_source=github&utm_medium=referral&utm_campaign=terraform-provider-admiral) is a control plane for coordinating infrastructure and application delivery across environments. This repository is one of its
[open-source tools](https://github.com/admiral-io).

- [Documentation](https://admiral.io/docs?utm_source=github&utm_medium=referral&utm_campaign=terraform-provider-admiral)
- A bug in this repository: [open an issue](https://github.com/admiral-io/terraform-provider-admiral/issues/new/choose)
- Anything else about Admiral, or not sure where it goes: [admiral-community](https://github.com/admiral-io/admiral-community)
- A security vulnerability: email [security@admiral.io](mailto:security@admiral.io), never a public issue

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
