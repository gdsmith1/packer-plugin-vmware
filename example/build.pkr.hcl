# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0
# From this folder, run: rm -rf builds/debian_aarch64 && packer init . && packer build -var-file=pkrvars/debian/fusion-13.pkrvars.hcl .

packer {
  required_version = ">= 1.7.0"
  required_plugins {
    vmware = {
      version = "~> 1.0.7"
      source  = "github.com/hashicorp/vmware"
    }
    vagrant = {
        version = "~> 1"
        source  = "github.com/hashicorp/vagrant"
    }
  }
}

build {
  sources = ["source.vmware-iso.debian"]
  provisioner "shell" {
    inline = [
    "sudo apt-get update > /home/vagrant/update.log",
    "sudo apt-get upgrade -y > /home/vagrant/upgrade.log",
    "sudo apt-get dist-upgrade -y > /home/vagrant/dist-upgrade.log",
    "echo 'Hello, World!' > hello.txt",
    "mkdir /home/vagrant/.ssh",
    "touch /home/vagrant/.ssh/authorized_keys",
    "wget --no-check-certificate https://raw.githubusercontent.com/hashicorp/vagrant/main/keys/vagrant.pub -O /home/vagrant/.ssh/authorized_keys",
    "chmod 700 /home/vagrant/.ssh",
    "chmod 600 /home/vagrant/.ssh/authorized_keys",
    "echo 'vagrant ALL=(ALL) NOPASSWD: ALL' | sudo tee /etc/sudoers.d/vagrant",
    "sudo chmod 440 /etc/sudoers.d/vagrant",
    "echo 'UseDNS no' | sudo tee -a /etc/ssh/sshd_config",
    "sudo service ssh restart",
    "sudo reboot",
    ]
    expect_disconnect = true
  }
  post-processor "vagrant" {
      output = "output-vmware-iso/package.box"
      keep_input_artifact = true # keeps the build folder
  }

  post-processor "vagrant" {
    output = "output-vmware-iso/package.vmx"
    keep_input_artifact = true 
  }

  post-processors {
    post-processor "vagrant" {
        keep_input_artifact = true # needed to pass to the next post-processor
    }
    post-processor "vagrant-cloud" {
        access_token = "YOUR TOKEN HERE"
        box_tag = "gibsmith619/debian11"
        version = "1.0.1"
        architecture = "arm64"
    }
  }

}
