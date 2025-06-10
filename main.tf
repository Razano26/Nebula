module "talos" {
  source = "./talos"

  providers = {
    proxmox = proxmox
  }

  image = {
    version = "v1.10.0"
    schematic = file("${path.module}/talos/image/schematic.yaml")
  }

  cilium = {
    install = file("${path.module}/talos/inline-manifests/cilium-install.yaml")
    values = file("${path.module}/talos/inline-manifests/cilium-values.yaml")
  }

  cluster = {
    name            = "talion"
    endpoint        = "162.38.112.152"
    gateway         = "162.38.112.254"
    private_gateway = "10.15.50.1"
    talos_version   = "v1.10.0"
    proxmox_cluster = "serpentard"
  }

  nodes = {
    "nebula-ctrl-00" = {
      host_node     = "serpentard-1"
      machine_type  = "controlplane"
      ip            = "10.15.50.100"
      mac_address   = "BC:24:11:2E:08:00"
      secondary_mac_address = "BC:24:11:2E:08:01"
      vm_id         = 800
      cpu           = 4
      ram_dedicated = 4096
    }
    "nebula-work-00" = {
      host_node     = "serpentard-2"
      machine_type  = "worker"
      ip            = "10.15.50.110"
      mac_address   = "BC:24:11:2E:08:02"
      vm_id         = 810
      cpu           = 4
      ram_dedicated = 4096
    }
    "nebula-work-01" = {
      host_node     = "serpentard-3"
      machine_type  = "worker"
      ip            = "10.15.50.111"
      mac_address   = "BC:24:11:2E:08:03"
      vm_id         = 811
      cpu           = 4
      ram_dedicated = 4096
    }
    "nebula-work-02" = {
      host_node     = "serpentard-1"
      machine_type  = "worker"
      ip            = "10.15.50.112"
      mac_address   = "BC:24:11:2E:08:04"
      vm_id         = 812
      cpu           = 4
      ram_dedicated = 4096
    }
  }
}

module "longhorn" {
  depends_on = [
    module.talos
  ]
  source = "./bootstrap/longhorn"

  providers = {
    kubernetes = kubernetes
    helm = helm
  }
}
