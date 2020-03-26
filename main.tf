provider "google" {
  credentials = file("service-account.json")

  project = "merge-chance-time"
  region  = "asia-northeast1"
  zone    = "asia-northeast1-a"
}

data "google_project" "current" {}

resource "google_app_engine_application" "app" {
  project     = "merge-chance-time"
  location_id = "asia-northeast1"
}

resource "google_service_account" "github_actions_service_account" {
  project      = "merge-chance-time"
  account_id   = "github-actions"
  display_name = "GitHub Actions"
}

resource "google_project_iam_binding" "github_actions_appengine_admin" {
  project = "merge-chance-time"
  role    = "roles/appengine.appAdmin"
  members = [
    "serviceAccount:${google_service_account.github_actions_service_account.email}"
  ]
}

resource "google_project_iam_binding" "github_actions_cloud_build" {
  project = "merge-chance-time"
  role    = "roles/cloudbuild.builds.builder"
  members = [
    "serviceAccount:${google_service_account.github_actions_service_account.email}"
  ]
}

// refs. https://cloud.google.com/cloud-build/docs/securing-builds/set-service-account-permissions
resource "google_project_iam_binding" "cloud_build_service_account" {
  project = "merge-chance-time"
  role    = "roles/cloudbuild.builds.builder"
  members = [
    "serviceAccount:${data.google_project.current.number}@cloudbuild.gserviceaccount.com"
  ]
}
