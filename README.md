# terraform-provider-luckperms

Terraform provider for managing [LuckPerms](https://luckperms.net/) permissions, groups, and tracks via the [LuckPerms REST API](https://github.com/LuckPerms/rest-api).

**Status:** Production-ready. Not yet published to the Terraform Registry.

## Quick Start

```hcl
provider "luckperms" {
  base_url = "http://localhost:8080"
}

resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Администрация"
  weight       = 100
  prefix       = "100.<#f1c40f>⭐"
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name

  node {
    key   = "*"
    value = true
  }

  node {
    key   = "group.moderator"
    value = true
  }
}

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = [
    luckperms_group.moderator.name,
    luckperms_group.admin.name,
  ]
}
```

## Resources

- `luckperms_group` — Create and manage group identity (display name, weight, prefix, suffix)
- `luckperms_group_nodes` — Manage permissions and inheritance nodes for a group
- `luckperms_track` — Manage promotion/demotion tracks

## Data Sources

- `luckperms_group` / `luckperms_groups` — Read group metadata and nodes
- `luckperms_track` / `luckperms_tracks` — Read track data

## Requirements

- Terraform >= 1.9
- Go >= 1.23 (for building)
- [LuckPerms REST API](https://github.com/LuckPerms/rest-api) running and accessible

## Provider Configuration

```hcl
provider "luckperms" {
  base_url = "http://localhost:8080"  # LuckPerms REST API URL
  api_key  = ""                       # Optional: API key if auth is enabled
}
```

Environment variables: `LUCKPERMS_BASE_URL`, `LUCKPERMS_API_KEY`.

## Development

### Build

```bash
make build        # Build provider binary
make test         # Run unit tests
make testacc      # Run acceptance tests (requires Docker)
make generate     # Generate docs
make fmt          # Format code
```

### Local Testing Setup

Use dev_overrides in `~/.terraformrc` to test against a local binary:

```hcl
provider_installation {
  dev_overrides {
    "digitaldrugstech/luckperms" = "/path/to/go/bin"
  }
  direct {}
}
```

Build and place the provider:

```bash
make build
# Binary output at: ./bin/terraform-provider-luckperms
# Copy to: $(go env GOPATH)/bin/
```

### Docker Compose

Run LuckPerms API locally:

```bash
docker-compose up -d
```

This starts:
- LuckPerms REST API on `http://localhost:8080`
- PostgreSQL on `localhost:5432`

### Generate Tool

Bootstrap Terraform configuration from existing LuckPerms state:

```bash
go run ./tools/generate \
  --url http://localhost:8080 \
  --api-key optional-key \
  --output ./generated/
```

This creates `.tf` files with all current groups, permissions, and tracks.

## Resources

See [docs/](docs/) for detailed resource and data source documentation.

## License

MPL-2.0
