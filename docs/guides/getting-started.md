---
page_title: "Getting started"
subcategory: ""
description: |-
  Configure the provider, create an application with its environments, and bring existing ones under management.
---

# Getting started

This guide configures the provider, creates an application with three
environments, and shows how to bring an application created elsewhere under
Terraform.

## Configure the provider

Create an API key in the Admiral console or with the
[Admiral CLI](https://github.com/admiral-io/admiral-cli), and export it:

```shell
export ADMIRAL_API_KEY="admp_..."
```

Then declare the provider:

```terraform
terraform {
  required_providers {
    admiral = {
      source  = "admiral-io/admiral"
      version = "~> 0.2"
    }
  }
}

provider "admiral" {}
```

## Create an application and its environments

An application is what a team owns in Admiral. Its environments are where
its components are deployed. Create the environments from one list so adding
one is a one-line change:

```terraform
resource "admiral_application" "billing" {
  name        = "billing-api"
  description = "Handles billing"

  labels = {
    team = "payments"
  }
}

resource "admiral_environment" "billing" {
  for_each = toset(["dev", "staging", "production"])

  application_id = admiral_application.billing.id
  name           = each.key
}
```

Run `terraform plan` to review the changes, then `terraform apply`.

## Read what another team manages

Use the data sources to refer to an application or environment without
managing it:

```terraform
data "admiral_application" "platform" {
  name = "platform"
}

data "admiral_environment" "platform_production" {
  application_id = data.admiral_application.platform.id
  name           = "production"
}
```

## Manage an existing application

An application created in the console or with the CLI can be brought under
Terraform with an `import` block (Terraform 1.5 and later). Look
up its ID with `admiral app get <name> -o wide`, then:

```terraform
import {
  to = admiral_application.billing
  id = "<application-uuid>"
}
```

`terraform plan` shows the import and any difference between the
configuration and the application as it exists.

## Next steps

- The [Admiral documentation](https://admiral.io/docs?utm_source=terraform-registry&utm_medium=referral&utm_campaign=terraform-provider-admiral)
  explains applications, environments and components.
- Report a bug in the provider in its
  [issue tracker](https://github.com/admiral-io/terraform-provider-admiral/issues/new/choose).
