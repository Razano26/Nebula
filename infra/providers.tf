terraform {
  required_providers {
    talos = {
      source  = "siderolabs/talos"
      version = "0.7.1"
    }
    proxmox = {
      source  = "bpg/proxmox"
      version = "0.76.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.36.0"
    }
    restapi = {
      source  = "Mastercard/restapi"
      version = "2.0.1"
    }
  }
}

provider "proxmox" {
  endpoint = var.proxmox.endpoint
  insecure = var.proxmox.insecure

  api_token = var.proxmox.api_token
  ssh {
    agent    = true
    username = var.proxmox.username
  }
}

provider "kubernetes" {
  host = module.talos.kube_config.kubernetes_client_configuration.host
  client_certificate = base64decode(module.talos.kube_config.kubernetes_client_configuration.client_certificate)
  client_key = base64decode(module.talos.kube_config.kubernetes_client_configuration.client_key)
  cluster_ca_certificate = base64decode(module.talos.kube_config.kubernetes_client_configuration.ca_certificate)
}

provider "restapi" {
  uri                  = var.proxmox.endpoint
  insecure             = var.proxmox.insecure
  write_returns_object = true

  headers = {
    "Content-Type"  = "application/json"
    "Authorization" = "PVEAPIToken=${var.proxmox.api_token}"
  }
}
