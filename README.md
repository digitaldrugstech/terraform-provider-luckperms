# terraform-provider-luckperms

Terraform provider for managing [LuckPerms](https://luckperms.net/) permissions, groups, and tracks via the [LuckPerms REST API](https://github.com/LuckPerms/rest-api).

> **Status:** In development. Not yet published to the Terraform Registry.

## Overview

Manage your Minecraft server permissions as Infrastructure as Code:

```hcl
resource "luckperms_group" "admin" {
  name = "admin"
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name

  node {
    key   = "*"
    value = true
  }

  node {
    key   = "prefix.100.<#f1c40f>★"
    value = true
  }

  node {
    key   = "displayname.Admin"
    value = true
  }

  node {
    key   = "weight.500"
    value = true
  }
}

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["helper", "admin", "superadmin"]
}
```

## Resources

- `luckperms_group` — Create and manage groups
- `luckperms_group_nodes` — Manage permissions, inheritance, prefixes, and metadata for a group
- `luckperms_track` — Manage promotion/demotion tracks

## Data Sources

- `luckperms_group` / `luckperms_groups` — Read group data
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

```bash
make build        # Build provider binary
make test         # Run unit tests
make testacc      # Run acceptance tests (requires Docker)
make generate     # Generate docs
make fmt          # Format code
```

See [SPEC.md](SPEC.md) for the full technical specification.

## License

MPL-2.0
