terraform {
  required_providers {
    corex = {
      source  = "registry.terraform.io/ne4u/corex"
      version = "~> 0.1"
    }
  }
}

provider "corex" {
  host     = var.corex_host
  username = var.corex_username
  password = var.corex_password
}

variable "corex_host" {
  type    = string
  default = "https://corex.example.com"
}

variable "corex_username" {
  type    = string
  default = "admin"
}

variable "corex_password" {
  type      = string
  sensitive = true
}
