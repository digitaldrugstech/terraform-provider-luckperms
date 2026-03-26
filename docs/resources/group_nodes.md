# luckperms_group_nodes Resource

Manages permission and inheritance nodes for a LuckPerms group. Meta nodes (displayname, weight, prefix, suffix) must be managed through the `luckperms_group` resource.

This resource uses full replacement semantics: Terraform owns ALL permission nodes. If you declare nodes in this resource, they will be the only permission nodes on the group after apply.

## Example Usage

### Basic Permissions

```terraform
resource "luckperms_group" "admin" {
  name   = "admin"
  weight = 100
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name

  node {
    key   = "*"
    value = true
  }

  node {
    key   = "luckperms.*"
    value = true
  }
}
```

### Group Inheritance

```terraform
resource "luckperms_group" "member" {
  name   = "member"
  weight = 10
}

resource "luckperms_group" "moderator" {
  name   = "moderator"
  weight = 50
}

resource "luckperms_group_nodes" "moderator" {
  group = luckperms_group.moderator.name

  node {
    key   = "group.${luckperms_group.member.name}"
    value = true
  }

  node {
    key   = "moderation.*"
    value = true
  }
}
```

### Context-Specific Permissions

```terraform
resource "luckperms_group_nodes" "builder" {
  group = "builder"

  node {
    key   = "worldedit.selection.*"
    value = true

    context {
      key   = "world"
      value = "creative"
    }
  }

  node {
    key   = "command.claim"
    value = true

    context {
      key   = "world"
      value = "survival"
    }
  }
}
```

### Temporary Permissions

```terraform
resource "luckperms_group_nodes" "vip" {
  group = luckperms_group.vip.name

  node {
    key    = "vip.commands"
    value  = true
    expiry = 1704067200  # 2024-01-01T00:00:00Z
  }
}
```

### Negated Permissions

```terraform
resource "luckperms_group_nodes" "restricted" {
  group = luckperms_group.restricted.name

  node {
    key   = "command.ban"
    value = false
  }

  node {
    key   = "command.mute"
    value = false
  }
}
```

## Argument Reference

- `group` (String, Required, Forces new) - The group name. Must reference an existing group created with `luckperms_group` or imported.

### Node Block

Each `node` block defines a single permission or inheritance node:

- `key` (String, Required) - Permission key. Examples: `*`, `worldedit.*`, `group.member`, `luckperms.admin`.
- `value` (Boolean, Optional, Default: true) - Permission value. `true` = grant, `false` = negated/deny.
- `expiry` (Number, Optional) - Unix timestamp for temporary nodes. Omit for permanent nodes.

#### Context Block

Each `context` block (within a node) defines a server/world context:

- `key` (String, Required) - Context key. Examples: `server`, `world`, `dimension`.
- `value` (String, Required) - Context value. Examples: `survival`, `creative`, `main-world`.

Multiple contexts with the same key within a single node represent OR semantics (e.g., server=creative-build OR server=creative-infrastructure). Different context keys represent AND semantics (e.g., server=survival AND world=nether — both must match).

## Attributes Reference

- `id` (String) - Resource identifier (same as the group name).
- `group` (String) - The group name.
- `node` (Set) - All permission and inheritance nodes (excludes meta nodes).

## Restrictions

**Meta Nodes Not Allowed**: The following node keys are reserved for the `luckperms_group` resource and cannot be used here:
- Keys starting with `displayname.`
- Keys starting with `weight.`
- Keys starting with `prefix.`
- Keys starting with `suffix.`

If you try to use a meta node key, Terraform will reject it with an error message.

## Import

Groups can be imported by group name:

```bash
terraform import luckperms_group_nodes.admin admin
```

This imports all current permission/inheritance nodes (meta nodes are excluded).

## Behavior Notes

**Full Replacement**: When you apply this resource, all current permission nodes are replaced with the nodes specified in the Terraform configuration. This is not an additive operation.

**Coordination with group**: The `luckperms_group` and `luckperms_group_nodes` resources coordinate automatically:
- When `luckperms_group` updates meta nodes, permission nodes are preserved
- When `luckperms_group_nodes` updates permission nodes, meta nodes are preserved
- You can safely use both resources for the same group

**Duplicate Detection**: Terraform will reject duplicate nodes (same key, value, and contexts).

**Deletion**: When this resource is deleted, all permission nodes are removed. Meta nodes (from `luckperms_group`) are preserved.

## Important: Use Resource References

Always write `group = luckperms_group.admin.name`, not `group = "admin"`. This establishes an implicit dependency in Terraform that ensures:

1. The group exists before nodes are created
2. Terraform tracks the relationship automatically
3. Changes to the group trigger node re-evaluation

**Example: Correct**

```terraform
resource "luckperms_group" "admin" {
  name = "admin"
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name  # ✓ Resource reference

  node {
    key   = "*"
    value = true
  }
}
```

**Example: Avoid**

```terraform
resource "luckperms_group_nodes" "admin" {
  group = "admin"  # ✗ String literal — no implicit dependency

  node {
    key   = "*"
    value = true
  }
}
```

When using string literals, you lose Terraform's dependency tracking. If the group doesn't exist, the apply fails without a clear relationship error.
