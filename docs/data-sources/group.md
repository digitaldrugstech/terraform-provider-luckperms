# luckperms_group Data Source

Reads information about a LuckPerms group, including its meta attributes and all nodes.

## Example Usage

### Read a Group

```terraform
data "luckperms_group" "admin" {
  name = "admin"
}

output "admin_weight" {
  value = data.luckperms_group.admin.weight
}

output "admin_display_name" {
  value = data.luckperms_group.admin.display_name
}
```

### Use Group Data in Resources

```terraform
data "luckperms_group" "moderator" {
  name = "moderator"
}

resource "luckperms_track" "staff" {
  name = "staff"
  groups = [
    "member",
    data.luckperms_group.moderator.name,
    "admin",
  ]
}
```

### Inspect All Nodes

```terraform
data "luckperms_group" "admin" {
  name = "admin"
}

output "all_nodes" {
  value = data.luckperms_group.admin.nodes
}
```

## Argument Reference

- `name` (String, Required) - The group name to read.

## Attributes Reference

- `name` (String) - The group name.
- `display_name` (String) - The display name (empty if not set).
- `weight` (Number) - The group weight. Default: 0.
- `prefix` (String) - The chat prefix in MiniMessage format (empty if not set).
- `suffix` (String) - The chat suffix in MiniMessage format (empty if not set).
- `nodes` (List of Objects) - All permission and inheritance nodes (meta nodes excluded).

### Node Object Attributes

Each node in the `nodes` list contains:

- `key` (String) - The permission key.
- `value` (Boolean) - The permission value (true = grant, false = negated).
- `expiry` (Number) - Unix timestamp for temporary nodes (0 if permanent).
- `context` (List of Objects) - Server/world contexts.

#### Context Object Attributes

- `key` (String) - Context key (e.g., `server`, `world`).
- `value` (String) - Context value (e.g., `survival`, `creative`).

## Example Output

```terraform
data "luckperms_group" "admin" {
  name = "admin"
}

# Output:
# display_name = "Администрация"
# weight = 100
# prefix = "100.<#f1c40f>⭐"
# suffix = "5.</dark_gray>[Admin]"
# nodes = [
#   {
#     key   = "*"
#     value = true
#   },
#   {
#     key    = "vip.commands"
#     value  = true
#     expiry = 1704067200
#     context = []
#   },
#   {
#     key   = "worldedit.*"
#     value = true
#     context = [
#       {
#         key   = "world"
#         value = "creative"
#       }
#     ]
#   }
# ]
```
