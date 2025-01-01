terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "2.46.1"
    }
  }
}

variable "do_token" {}

provider "digitalocean" {
    token = var.do_token
}

# Create a new Web Droplet in the nyc2 region
resource "digitalocean_droplet" "web" {
#   image   = "s-1vcpu-1gb" //1gb/1cpu .009/hour
  image = "s-1vcpu-2gb" // 2gb/1cpu .018/hour
  name    = "web-1"
  region  = "nyc3" //nyc1 & nyc3 available
  size    = "s-1vcpu-1gb"
#   backups = true //default is false
#   backup_policy {
#     plan    = "weekly"
#     weekday = "TUE"
#     hour    = 8
#   }
}