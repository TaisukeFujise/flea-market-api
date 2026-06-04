# main.tf : terraform・providerブロック
terraform {
  required_version = "~> 1.15"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.33"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  backend "gcs" {
    bucket = "term9-taisuke-fujise-tfstate"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
