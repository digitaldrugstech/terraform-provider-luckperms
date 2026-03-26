terraform {
  required_providers {
    luckperms = {
      source  = "digitaldrugstech/luckperms"
      version = "~> 0.1"
    }
  }
}

provider "luckperms" {
  base_url = "http://localhost:8080"
  api_key  = "your-api-key-here"
  timeout  = 30
  insecure = false
}

# ─── Groups ───────────────────────────────────────────────────────────────────

resource "luckperms_group" "default" {
  name         = "default"
  display_name = "Default"
  weight       = 0
}

resource "luckperms_group" "player" {
  name         = "player"
  display_name = "Player"
  weight       = 10
  prefix       = "10.<gray>[Player]</gray> "
}

resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Admin"
  weight       = 100
  prefix       = "100.<red>[Admin]</red> "
  suffix       = "100. <dark_gray>✦</dark_gray>"
}

# ─── Group Nodes ──────────────────────────────────────────────────────────────

resource "luckperms_group_nodes" "default" {
  group = luckperms_group.default.name

  node {
    key   = "minecraft.command.me"
    value = true
  }

  node {
    key   = "minecraft.command.tell"
    value = true
  }
}

resource "luckperms_group_nodes" "player" {
  group = luckperms_group.player.name

  node {
    key   = "group.default"
    value = true
  }

  node {
    key   = "essentials.home"
    value = true
  }

  node {
    key   = "essentials.sethome"
    value = true
  }

  node {
    key   = "essentials.tpa"
    value = true
  }

  # Server-specific node
  node {
    key   = "worldedit.navigation.jumpto.tool"
    value = true

    context {
      key   = "server"
      value = "creative-build"
    }
  }
}

resource "luckperms_group_nodes" "admin" {
  group = luckperms_group.admin.name

  node {
    key   = "group.player"
    value = true
  }

  node {
    key   = "luckperms.*"
    value = true
  }

  node {
    key   = "essentials.*"
    value = true
  }

  node {
    key   = "worldedit.*"
    value = true
  }

  # Negated node example
  node {
    key   = "some.dangerous.permission"
    value = false
  }
}

# ─── Track ────────────────────────────────────────────────────────────────────

resource "luckperms_track" "staff" {
  name   = "staff"
  groups = ["default", "player", "admin"]

  depends_on = [
    luckperms_group.default,
    luckperms_group.player,
    luckperms_group.admin,
  ]
}

# ─── Data Sources ─────────────────────────────────────────────────────────────

data "luckperms_group" "admin_info" {
  name = luckperms_group.admin.name
}

data "luckperms_groups" "all" {}

data "luckperms_track" "staff_info" {
  name = luckperms_track.staff.name
}

data "luckperms_tracks" "all" {}

# ─── Outputs ──────────────────────────────────────────────────────────────────

output "admin_group_weight" {
  value = data.luckperms_group.admin_info.weight
}

output "all_group_names" {
  value = data.luckperms_groups.all.names
}

output "staff_track_groups" {
  value = data.luckperms_track.staff_info.groups
}
