# luckperms_tracks Data Source

Lists all track names in LuckPerms.

## Example Usage

### List All Tracks

```terraform
data "luckperms_tracks" "all" {}

output "all_tracks" {
  value = data.luckperms_tracks.all.names
}
```

### Validate Track Existence

```terraform
data "luckperms_tracks" "all" {}

locals {
  all_track_names = toset(data.luckperms_tracks.all.names)
}

# Use in locals to reference tracks
resource "local_file" "track_report" {
  content = join("\n", [
    "Available tracks: ${join(", ", data.luckperms_tracks.all.names)}"
  ])
  filename = "tracks.txt"
}
```

## Argument Reference

This data source takes no arguments.

## Attributes Reference

- `names` (List of Strings) - All track names in LuckPerms.

## Example Output

```
names = [
  "staff",
  "vip",
  "moderator_track"
]
```
