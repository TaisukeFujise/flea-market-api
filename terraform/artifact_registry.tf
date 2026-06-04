resource "google_artifact_registry_repository" "main" {
  location      = var.region
  repository_id = "flea-market"
  format        = "DOCKER"

  depends_on = [google_project_service.apis]
}

output "registry_url" {
  description = "DockerイメージのプッシュURL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/flea-market"
}
