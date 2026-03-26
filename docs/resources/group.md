# luckperms_group Resource

Manages a LuckPerms group with its meta attributes (display name, weight, prefix, suffix).

Use the `luckperms_group_nodes` resource to manage permission nodes separately.

## Example Usage

### Basic Group

```terraform
resource "luckperms_group" "member" {
  name   = "member"
  weight = 10
}
```

### Group with Display Name and Prefix

```terraform
resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Администрация"
  weight       = 100
  prefix       = "100.<#f1c40f>⭐"
  suffix       = "5.</dark_gray>[Admin]"
}
```

### Group with Only Display Name

```terraform
resource "luckperms_group" "moderator" {
  name         = "moderator"
  display_name = "Модератор"
}
```

## Argument Reference

- `name` (String, Required, Forces new) - Group name. Must be lowercase alphanumeric with underscores (^[a-z0-9_]+$). Immutable after creation.
- `display_name` (String, Optional) - Human-readable display name. Stored as a `displayname.{value}` meta node in LuckPerms. Can be omitted to remove the display name.
- `weight` (Number, Optional, Default: 0) - Group weight/priority. Higher weights take precedence. Stored as a `weight.{value}` meta node. Default: 0.
- `prefix` (String, Optional) - Chat prefix in MiniMessage format with priority prefix. Format: `{priority}.{text}`. Example: `100.<#f1c40f>⭐`. Can be omitted to remove the prefix.
- `suffix` (String, Optional) - Chat suffix in MiniMessage format with priority prefix. Format: `{priority}.{text}`. Can be omitted to remove the suffix.

## Attributes Reference

- `name` (String) - The group name.
- `display_name` (String) - The display name (empty if not set).
- `weight` (Number) - The group weight (default 0).
- `prefix` (String) - The chat prefix (empty if not set).
- `suffix` (String) - The chat suffix (empty if not set).

## Meta Node Storage

Meta attributes are stored as special permission nodes in LuckPerms:

- `display_name` → `displayname.{value}`
- `weight` → `weight.{value}`
- `prefix` → `prefix.{value}` (where value is the full string you set, e.g., `100.<#f1c40f>⭐`)
- `suffix` → `suffix.{value}`

The `luckperms_group_nodes` resource will not include these meta nodes. They are managed exclusively by the `luckperms_group` resource.

## Import

Groups can be imported by name:

```bash
terraform import luckperms_group.admin admin
```

## Special Behavior

**Default Group**: The default group cannot be deleted. Terraform will skip deletion if the group name is "default".

~> **Warning:** When removing the `default` group from Terraform config, Terraform will remove it from state but the group will continue to exist in LuckPerms. The next `terraform plan` will show it as needing creation.

**Coordination with group_nodes**: The `luckperms_group` and `luckperms_group_nodes` resources work together:
- `luckperms_group` manages meta nodes (displayname, weight, prefix, suffix)
- `luckperms_group_nodes` manages permission and inheritance nodes

When using both resources for the same group, Terraform maintains separation by preserving meta nodes during `luckperms_group_nodes` updates and preserving permission nodes during `luckperms_group` updates.

~> **Note:** Both resources use a read-merge-write pattern against the same API endpoint. If external tools modify nodes between the read and write, those changes may be lost. Within a single Terraform apply this is safe because Terraform serializes resource operations.
