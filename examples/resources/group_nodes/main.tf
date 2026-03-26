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

resource "luckperms_group" "builder" {
  name   = "builder"
  weight = 30
}

resource "luckperms_group_nodes" "builder" {
  group = luckperms_group.builder.name

  node {
    key   = "group.member"
    value = true
  }

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

  node {
    key   = "command.ban"
    value = false
  }
}

resource "luckperms_group" "vip_temporary" {
  name   = "vip"
  weight = 25
}

resource "luckperms_group_nodes" "vip_temporary" {
  group = luckperms_group.vip_temporary.name

  node {
    key   = "group.member"
    value = true
  }

  node {
    key    = "vip.bonus_claim_blocks"
    value  = true
    # WARNING: Static timestamps in Terraform cause perpetual drift once expired.
    # Temporary permissions are better managed outside Terraform (web editor, API, commands).
    # This example is for illustration only.
    expiry = 1704067200  # 2024-01-01T00:00:00Z
  }

  node {
    key   = "vip.colored_chat"
    value = true
  }
}
