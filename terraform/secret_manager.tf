locals {
  secrets = [
    "DATABASE_URL",
    "MESHY_API_KEY",
    "MESHY_WEBHOOK_SECRET",
  ]
}

resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(local.secrets)
  secret_id = each.value

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}
