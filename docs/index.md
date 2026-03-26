# LuckPerms Provider

The LuckPerms Terraform provider enables you to manage LuckPerms permissions, groups, and tracks through Infrastructure as Code.

## Example Usage

```terraform
terraform {
  required_version = ">= 1.9"

  required_providers {
    luckperms = {
      source = "digitaldrugstech/luckperms"
    }
  }
}

provider "luckperms" {
  base_url = "http://localhost:8080"
  api_key  = var.luckperms_api_key
  timeout  = 30
}

resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Администрация"
  weight       = 100
  prefix       = "100.<#f1c40f>⭐"
}

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["moder", "admin"]
}
```

## Schema

### Required

- `base_url` (String) - Base URL of the LuckPerms REST API. Can also be set via the `LUCKPERMS_BASE_URL` environment variable.

### Optional

- `api_key` (String, Sensitive) - API key for authentication. Can also be set via the `LUCKPERMS_API_KEY` environment variable.
- `timeout` (Number) - HTTP request timeout in seconds. Default: `30`. Can also be set via the `LUCKPERMS_TIMEOUT` environment variable.
- `insecure` (Boolean) - Skip TLS certificate verification. Default: `false`. Can also be set via the `LUCKPERMS_INSECURE` environment variable.

## Configuration via Environment Variables

You can configure the provider using environment variables instead of the configuration block:

```bash
export LUCKPERMS_BASE_URL="http://localhost:8080"
export LUCKPERMS_API_KEY="your-api-key"
export LUCKPERMS_TIMEOUT="30"
export LUCKPERMS_INSECURE="false"
```

Then use the provider without explicit configuration:

```terraform
provider "luckperms" {}
```

## Connecting to LuckPerms

The provider requires the LuckPerms REST API to be running and accessible. Ensure:

1. LuckPerms is installed on your Minecraft server
2. The REST API extension is enabled
3. The API is listening on the configured `base_url`
4. The `api_key` (if required) matches the REST API configuration

The provider will validate the connection during initialization by performing a health check.

## Resource Model

The provider uses two complementary resources to manage groups:

### luckperms_group

Manages group identity and metadata:
- `display_name` — Display name in chat and UI
- `weight` — Priority/rank (higher wins)
- `prefix` — Chat prefix with priority
- `suffix` — Chat suffix with priority

```terraform
resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Administrator"
  weight       = 100
  prefix       = "100.<#f1c40f>⭐"
}
```

### luckperms_group_nodes

Manages permissions and inheritance for a group. Always reference the group using the resource attribute (not a string literal):

```terraform
resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name  # Reference, not "admin"

  node {
    key   = "*"
    value = true
  }

  node {
    key   = "group.moderator"
    value = true
  }
}
```

**Why references matter:** Implicit dependencies ensure the group exists before nodes are created. Terraform tracks these automatically.

## Migrating Existing State

### Using the Generate Tool

If you have an existing LuckPerms instance, bootstrap your Terraform configuration:

```bash
go run ./tools/generate \
  --url http://localhost:8080 \
  --api-key your-api-key \
  --output ./generated/
```

This creates `.tf` files for all groups, tracks, and their nodes.

### Import Workflow

After generating `.tf` files with the generate tool, import them into Terraform state:

1. Write the `.tf` resource blocks (from generate tool output or manually)
2. Run `terraform import` for each resource:

```bash
# Import a group and its metadata
terraform import luckperms_group.admin admin

# Import a group's permission nodes
terraform import luckperms_group_nodes.admin admin

# Import a track
terraform import luckperms_track.staff staff
```

3. Run `terraform plan` to verify no changes are needed

The resource blocks must exist in your `.tf` files before importing.

## Security Considerations

The `api_key` provider attribute is marked as sensitive and will be redacted in plan/apply output. However, it is stored in plaintext in Terraform state files. Use remote state backends with encryption (Terraform Cloud, S3+KMS, etc.) and never commit `.tfstate` files to version control.
