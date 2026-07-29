terraform {
  required_version = ">= 1.7.0"

  required_providers {
    stackit = {
      source  = "stackitcloud/stackit"
      version = "~> 0.104" # pin to a minor version
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.31"
    }

    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }

  backend "s3" {
    # never use local state for shared infra
    bucket = "paas-tf-state"
    key    = "ske/pass/terraform.tfstate"

    region = "us-east-1"
    endpoints = {
      s3 = "https://object.storage.eu01.onstackit.cloud"
    }
    encrypt = true

    use_lockfile   = true
    use_path_style = true

    skip_region_validation      = true
    skip_credentials_validation = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
  }
}

provider "stackit" {
  default_region           = "eu01"
  service_account_key      = var.service_account_key
}

provider "helm" {
  kubernetes {
    host = data.stackit_ske_kubeconfig.cluster.host

    client_certificate = base64decode(
      data.stackit_ske_kubeconfig.cluster.client_certificate
    )

    client_key = base64decode(
      data.stackit_ske_kubeconfig.cluster.client_key
    )

    cluster_ca_certificate = base64decode(
      data.stackit_ske_kubeconfig.cluster.cluster_ca_certificate
    )
  }
}
