resource "google_storage_bucket" "assets" {
  name                        = "term9-taisuke-fujise-assets"
  location                    = var.region
  uniform_bucket_level_access = true

  cors {
    origin          = ["*"]
    method          = ["GET"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }

  depends_on = [google_project_service.apis]
}

# 商品画像・3Dモデルは誰でも閲覧可能
resource "google_storage_bucket_iam_member" "public_read" {
  bucket = google_storage_bucket.assets.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}
