# terraform-provider-admiral

The Terraform provider for Admiral. Go, terraform-plugin-framework, talking to
the Admiral API over gRPC through `go.admiral.io/sdk`.

## The CLI is the reference for client conventions

`../admiral-cli` is the other first-party client of the same API. When this
provider needs to do something the CLI already does, do it the CLI's way:

- **Connecting** -- `admiral-cli/internal/client`: API keys use the `Token`
  auth scheme, every RPC gets a deadline, `plaintext` (no TLS) and `insecure`
  (TLS, unverified) are distinct.
- **Filters** -- `admiral-cli/internal/filter`: the server's filter DSL and
  how a user-supplied value is quoted into it. `internal/provider/filter.go`
  is a port; keep them in step.
- **Names** -- attributes and environment variables follow the CLI's config
  keys (`server`, `api_key`/`ADMIRAL_API_KEY`, `insecure`, `plaintext`).
- **Update semantics** -- the API's Update RPCs take a field mask. A
  Terraform plan is the whole desired state, so the resource sends a mask of
  every mutable field; a value removed from the configuration is cleared.

## Read the knowledge store before deciding anything architectural

Settled cross-service decisions live in `../admiral-knowledge/store/`, a
sibling checkout. `records/` holds the numbered decisions; read those first.
It is local only. Most relevant here: **0002** (managed services are curated
Terraform components, not bespoke executors).

## Working here

- `make lint`, `make test`, `make generate-verify` are what CI runs.
- `make generate` needs a `terraform` binary on `PATH` (tfenv: `tfenv use`).
- Acceptance tests (`make testacc`) run against a real tenant and create and
  delete applications there. They need `ADMIRAL_API_KEY`.
- `docs/` is generated. Edit `templates/` and `examples/`, never `docs/`.
