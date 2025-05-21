module "talos" {
  source = "./talos"

  providers = {
    proxmox = proxmox
  }

  image = {
    version = "v1.10.0"
    schematic = file("${path.module}/talos/image/schematic.yaml")
  }

  cluster = {
    name            = "talion"
    endpoint        = "162.38.112.152"
    public_gateway  = "162.38.112.254"
    private_gateway = "10.15.50.1"
    talos_version   = "v1.10.0"
    proxmox_cluster = "serpentard"
  }

  nodes = {
    "ctrl-00" = {
      host_node            = "serpentard-1"
      machine_type         = "controlplane"
      ip                   = "10.15.50.100"
      mac_address          = "BC:24:11:2E:C8:00"
      secondary_mac_address = "BC:24:11:2E:C8:10"
      vm_id               = 800
      cpu                 = 4
      ram_dedicated       = 4096
    }
    "ctrl-01" = {
      host_node            = "serpentard-2"
      machine_type         = "controlplane"
      ip                   = "10.15.50.101"
      mac_address          = "BC:24:11:2E:C8:01"
      secondary_mac_address = "BC:24:11:2E:C8:11"
      vm_id               = 801
      cpu                 = 4
      ram_dedicated       = 4096
    }
    "ctrl-02" = {
      host_node            = "serpentard-3"
      machine_type         = "controlplane"
      ip                   = "10.15.50.102"
      mac_address          = "BC:24:11:2E:C8:02"
      secondary_mac_address = "BC:24:11:2E:C8:12"
      vm_id               = 802
      cpu                 = 4
      ram_dedicated       = 4096
    }
    "work-00" = {
      host_node     = "serpentard-1"
      machine_type  = "worker"
      ip            = "10.15.50.110"
      mac_address   = "BC:24:11:2E:08:00"
      vm_id         = 810
      cpu           = 4
      ram_dedicated = 4096
    }
  }
}
