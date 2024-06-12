# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_version = ">= 1.7.0"
  required_plugins {
    vmware = {
      version = ">= 1.0.7"
      source  = "github.com/hashicorp/vmware"
    }
  }
}

build {
  sources = ["source.vmware-iso.debian"]
  provisioner "shell" {
    inline = [
      "sudo apt-get update > /home/vagrant/update.log",
      "sudo apt-get upgrade -y > /home/vagrant/upgrade.log",
      "echo 'Hello, World!' > hello.txt",
      "mkdir /home/vagrant/.ssh",
      "touch /home/vagrant/.ssh/authorized_keys",
      "echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOroZ8aIt32D6VyFOd/QRF7+DHaIsv4N0qNB6/GwgOnh admin@admins-MBP.hsd1.ca.comcast.net' > /home/vagrant/.ssh/authorized_keys",
      "chmod 700 /home/vagrant/.ssh",
      "chmod 600 /home/vagrant/.ssh/authorized_keys",
    ]
  }
}
