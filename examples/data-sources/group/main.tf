data "luckperms_group" "admin" {
  name = "admin"
}

output "admin_display_name" {
  value = data.luckperms_group.admin.display_name
}

output "admin_weight" {
  value = data.luckperms_group.admin.weight
}

output "admin_prefix" {
  value = data.luckperms_group.admin.prefix
}

output "admin_suffix" {
  value = data.luckperms_group.admin.suffix
}

output "admin_nodes_count" {
  value = length(data.luckperms_group.admin.nodes)
}

data "luckperms_groups" "all" {}

output "all_groups" {
  value = data.luckperms_groups.all.names
}

data "luckperms_track" "staff" {
  name = "staff"
}

output "staff_progression" {
  value = data.luckperms_track.staff.groups
}

data "luckperms_tracks" "all" {}

output "all_tracks" {
  value = data.luckperms_tracks.all.names
}
