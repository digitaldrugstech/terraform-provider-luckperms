resource "luckperms_group" "member" {
  name   = "member"
  weight = 1
}

resource "luckperms_group" "moderator" {
  name   = "moderator"
  weight = 50
}

resource "luckperms_group" "senior_mod" {
  name   = "senior_mod"
  weight = 75
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
    luckperms_group.senior_mod.name,
    luckperms_group.admin.name,
  ]
}

resource "luckperms_track" "vip" {
  name   = "vip"
  groups = ["member", "vip", "vip_plus"]
}
