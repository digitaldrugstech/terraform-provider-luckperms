# luckperms_track Data Source

Reads information about a LuckPerms track.

## Example Usage

### Read a Track

```terraform
data "luckperms_track" "staff" {
  name = "staff"
}

output "staff_groups" {
  value = data.luckperms_track.staff.groups
}
```

### Use Track Data in Other Resources

```terraform
data "luckperms_track" "staff" {
  name = "staff"
}

# Add a new group to the end of the track
resource "luckperms_track" "staff_updated" {
  name = "staff"
  groups = concat(
    data.luckperms_track.staff.groups,
    ["senior_admin"]
  )
}
```

## Argument Reference

- `name` (String, Required) - The track name to read.

## Attributes Reference

- `name` (String) - The track name.
- `groups` (List of Strings) - The ordered list of group names in the track.

## Example Output

```
name = "staff"
groups = [
  "member",
  "moderator",
  "admin"
]
```
