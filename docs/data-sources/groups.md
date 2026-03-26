# luckperms_groups Data Source

Lists all group names in LuckPerms.

## Example Usage

### List All Groups

```terraform
data "luckperms_groups" "all" {}

output "all_groups" {
  value = data.luckperms_groups.all.names
}
```

### Use in Validation

```terraform
data "luckperms_groups" "all" {}

locals {
  all_group_names = toset(data.luckperms_groups.all.names)
}

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["member", "moderator", "admin"]

  depends_on = [
    # Ensure all referenced groups exist
    data.luckperms_groups.all,
  ]
}
```

## Argument Reference

This data source takes no arguments.

## Attributes Reference

- `names` (List of Strings) - All group names in LuckPerms.

## Example Output

```
names = [
  "admin",
  "member",
  "moderator",
  "mute",
  "vip",
  "default"
]
```
