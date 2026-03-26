# luckperms_track Resource

Manages a LuckPerms track. Tracks define an ordered progression of groups used for promotions and demotions.

## Example Usage

### Basic Staff Track

```terraform
resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["member", "moderator", "admin"]
}
```

### Multiple Tracks

```terraform
resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["member", "moderator", "senior_mod", "admin"]
}

resource "luckperms_track" "vip" {
  name   = "vip"
  groups = ["member", "vip", "vip_plus"]
}
```

### Track with References to Groups

```terraform
resource "luckperms_group" "member" {
  name   = "member"
  weight = 1
}

resource "luckperms_group" "moderator" {
  name   = "moderator"
  weight = 50
}

resource "luckperms_group" "admin" {
  name   = "admin"
  weight = 100
}

resource "luckperms_track" "staff" {
  name = "staff"
  groups = [
    luckperms_group.member.name,
    luckperms_group.moderator.name,
    luckperms_group.admin.name,
  ]
}
```

## Argument Reference

- `name` (String, Required, Forces new) - Track name. Must be lowercase alphanumeric with underscores (^[a-z0-9_]+$). Immutable after creation.
- `groups` (List of Strings, Required) - Ordered list of group names in the track. Groups are used in order for promotions and demotions. All groups must exist in LuckPerms.

## Attributes Reference

- `name` (String) - The track name.
- `groups` (List of Strings) - The ordered list of group names.

## Import

Tracks can be imported by name:

```bash
terraform import luckperms_track.staff staff
```

## Behavior Notes

**Order Matters**: The order of groups in the `groups` list is significant. Groups appear in the order they would be used for progression (e.g., first group is the starting point, last is the highest).

**All Groups Must Exist**: Terraform will validate that all groups referenced in the track exist in LuckPerms.

**Track Progression**: Tracks are typically used with `/lp user promote` and `/lp user demote` commands in LuckPerms:
- Promoting moves a user up the track (to the next group)
- Demoting moves a user down the track (to the previous group)

**No Automatic Group Creation**: This resource only manages the track definition. Groups must be created separately using the `luckperms_group` resource.
