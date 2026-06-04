resource "random_password" "db_password" {
  length  = 32
  special = false
}

resource "google_sql_database_instance" "main" {
  name             = "flea-market-db"
  region           = var.region
  database_version = "POSTGRES_17"

  settings {
    tier    = "db-f1-micro"
    edition = "ENTERPRISE"

    backup_configuration {
      enabled = false
    }

    ip_configuration {
      ipv4_enabled = true
    }
  }

  deletion_protection = false

  depends_on = [google_project_service.apis]
}

resource "google_sql_database" "main" {
  name     = "flea_market"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "main" {
  name     = "flea_market"
  instance = google_sql_database_instance.main.name
  password = random_password.db_password.result
}

output "db_connection_name" {
  description = "Cloud Run の cloudsql-instances アノテーションに使用"
  value       = google_sql_database_instance.main.connection_name
}

output "db_password" {
  description = "terraform output db_password で確認してSecret Managerに登録"
  value       = random_password.db_password.result
  sensitive   = true
}
