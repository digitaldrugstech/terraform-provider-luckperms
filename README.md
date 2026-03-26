<h1 align="center">terraform-provider-luckperms</h1>

<p align="center">
Declarative management of Minecraft server permissions via Infrastructure as Code.
</p>

<p align="center">
  <a href="https://github.com/digitaldrugstech/terraform-provider-luckperms/actions/workflows/test.yml"><img src="https://github.com/digitaldrugstech/terraform-provider-luckperms/actions/workflows/test.yml/badge.svg" alt="Tests"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MPL--2.0-blue" alt="License: MPL-2.0"></a>
  <a href="https://github.com/digitaldrugstech/terraform-provider-luckperms"><img src="https://img.shields.io/badge/terraform-provider-purple" alt="Terraform Provider"></a>
</p>

---

Manage [LuckPerms](https://luckperms.net/) groups, permissions, tracks, and metadata through Terraform. Review permission changes in PRs, detect drift with `terraform plan`, roll back with `git revert`.

## Quick Start

```hcl
terraform {
  required_version = ">= 1.9"

  required_providers {
    luckperms = {
      source  = "digitaldrugstech/luckperms"
      version = "~> 0.1"
    }
  }
}

provider "luckperms" {
  base_url = "http://localhost:8080"
}

resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Администрация"
  weight       = 500
  prefix       = "100.<#f1c40f>⭐"
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name

  node { key = "*" }
  node { key = "luckperms.autoop" }
  node { key = "handcuffs.bypass" }
  node { key = "sayanvanish.*" }

  node {
    key   = "vulcan.bypass.*"
    value = false
  }
}

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = [
    luckperms_group.admin.name,
  ]
}
```

## Resources and Data Sources

### Resources

| Resource | Description |
|----------|-------------|
| `luckperms_group` | Group identity and meta attributes (display name, weight, prefix, suffix) |
| `luckperms_group_nodes` | Permission and inheritance nodes for a group |
| `luckperms_track` | Promotion/demotion track through groups |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `luckperms_group` | Read a single group with all metadata and nodes |
| `luckperms_groups` | List all group names |
| `luckperms_track` | Read a single track |
| `luckperms_tracks` | List all track names |

## How It Works

The provider splits LuckPerms node management between two resources:

- **`luckperms_group`** owns meta nodes: `displayname.*`, `weight.*`, `prefix.*`, `suffix.*`
- **`luckperms_group_nodes`** owns everything else: permissions, inheritance (`group.*`), contexts

Both resources coordinate through a read-merge-write pattern against the same `PUT /group/{name}/nodes` endpoint. When one updates, it preserves the other's nodes.

```
luckperms_group "admin"         luckperms_group_nodes "admin"
  display_name = "Admin"           node { key = "*" }
  weight       = 500               node { key = "group.helper" }
  prefix       = "100.<red>★"      node { key = "vulcan.bypass.*", value = false }
       |                                    |
       +------------- PUT /group/admin/nodes (merged) -----------+
```

## Provider Configuration

```hcl
provider "luckperms" {
  base_url = "http://localhost:8080"   # Required (or LUCKPERMS_BASE_URL env)
  api_key  = ""                        # Optional (or LUCKPERMS_API_KEY env)
  timeout  = 30                        # Optional, seconds
  insecure = false                     # Optional, skip TLS verify
}
```

## Migrating Existing Permissions

Bootstrap `.tf` files from your current LuckPerms state:

```bash
go run ./tools/generate \
  --url http://your-luckperms-api:8080 \
  --output ./luckperms/
```

This generates `groups.tf`, `group_nodes.tf`, and `tracks.tf` with resource references. Then import into Terraform state:

```bash
terraform import luckperms_group.admin admin
terraform import luckperms_group_nodes.admin admin
terraform import luckperms_track.staff staff
```

Verify with `terraform plan` -- should show no changes if generated config matches prod.

## Development

### Prerequisites

- Go >= 1.23
- Docker (for acceptance tests)

### Commands

```bash
make build       # Compile provider binary
make test        # Unit tests
make testacc     # Acceptance tests (requires running LuckPerms API)
make fmt         # Format code
make vet         # Lint
```

### Local LuckPerms API

```bash
docker compose up -d    # Starts LP REST API on localhost:9094 (H2 in-memory)
make testacc            # Run tests against it
docker compose down
```

### Local Provider Testing

Add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "digitaldrugstech/luckperms" = "/path/to/go/bin"
  }
  direct {}
}
```

Then `go install` and use normally with `terraform plan`/`apply`.

## Documentation

Full resource and data source documentation: [`docs/`](docs/)

## License

[MPL-2.0](LICENSE)
