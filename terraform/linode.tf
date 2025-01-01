terraform {
  required_providers {
    linode = {
      source = "linode/linode"
      version = "2.31.1"
    }
  }
}

# Configure the Linode Provider
provider "linode" {
  # token = "..."
}

# Create a Linode
resource "linode_instance" "test_instance" {
    label = "pepita"
    region = "us-southeast"
    type   = "g6-nanode-1"
    private_ip = true
}