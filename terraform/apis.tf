resource "google_project_service" "apis" {
  for_each = toset([
    # Cloud Run
    "run.googleapis.com",

    # Cloud SQL（ComputeはCloud SQLの内部依存）
    "sqladmin.googleapis.com",
    "compute.googleapis.com",

    # Cloud Storage
    "storage.googleapis.com",

    # Artifact Registry（Cloud RunのDockerイメージ管理）
    "artifactregistry.googleapis.com",

    # Firebase Authentication
    "identitytoolkit.googleapis.com",

    # Vertex AI（Gemini + Multimodal Embedding）
    "aiplatform.googleapis.com",

    # Secret Manager
    "secretmanager.googleapis.com",

    # Cloud Build
    "cloudbuild.googleapis.com",

    # Workload Identity Federation（CI/CD用）
    "iamcredentials.googleapis.com",

  ])

  project = var.project_id
  service = each.value

  disable_on_destroy = false
}
