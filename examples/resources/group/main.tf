resource "luckperms_group" "member" {
  name   = "member"
  weight = 1
}

resource "luckperms_group" "moderator" {
  name         = "moderator"
  display_name = "Модератор"
  weight       = 50
  prefix       = "50.<#3498db>⚔"
}

resource "luckperms_group" "admin" {
  name         = "admin"
  display_name = "Администрация"
  weight       = 100
  prefix       = "100.<#f1c40f>⭐"
  suffix       = "5.</dark_gray>[Admin]"
}

resource "luckperms_group" "vip" {
  name         = "vip"
  display_name = "VIP Участник"
  weight       = 25
  prefix       = "25.<gradient:yellow:gold>✦"
}
