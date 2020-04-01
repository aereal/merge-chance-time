terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "org-aereal"

    workspaces {
      name = "merge-chance-time"
    }
  }
}

variable "google_service_account" {}

provider "google" {
  credentials = base64decode(var.google_service_account)

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
  role    = "projects/${data.google_project.current.project_id}/roles/${google_project_iam_custom_role.github_actions_executor.role_id}"
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

resource "google_project_iam_custom_role" "github_actions_executor" {
  role_id = "GithubActionsExecutor"
  title   = "GitHub Actions Executor"
  permissions = [
    "appengine.applications.get",
    "appengine.applications.update",
    "appengine.operations.get",
    "appengine.services.list",
    "appengine.services.update",
    "appengine.versions.create",
    "appengine.versions.delete",
    "appengine.versions.list",
    "appengine.versions.update",
    "artifactregistry.files.get",
    "artifactregistry.files.list",
    "artifactregistry.packages.get",
    "artifactregistry.packages.list",
    "artifactregistry.repositories.downloadArtifacts",
    "artifactregistry.repositories.get",
    "artifactregistry.repositories.list",
    "artifactregistry.repositories.uploadArtifacts",
    "artifactregistry.tags.create",
    "artifactregistry.tags.get",
    "artifactregistry.tags.list",
    "artifactregistry.tags.update",
    "artifactregistry.versions.get",
    "artifactregistry.versions.list",
    "cloudbuild.builds.create",
    "cloudbuild.builds.get",
    "cloudbuild.builds.list",
    "cloudbuild.builds.update",
    "logging.logEntries.create",
    "pubsub.topics.create",
    "pubsub.topics.publish",
    "remotebuildexecution.blobs.get",
    "resourcemanager.projects.get",
    "source.repos.get",
    "source.repos.list",
    "storage.buckets.create",
    "storage.buckets.get",
    "storage.buckets.list",
    "storage.objects.create",
    "storage.objects.delete",
    "storage.objects.get",
    "storage.objects.list",
    "storage.objects.update",
  ]
}

resource "google_pubsub_topic" "cron_topic" {
  name = "cron"
}

resource "google_pubsub_subscription" "cron_subscription" {
  name  = "cron-subscription"
  topic = google_pubsub_topic.cron_topic.name
}

resource "google_cloud_scheduler_job" "invoke_endpoint" {
  name     = "invoke-endpoint"
  schedule = "*/2 * * * *"
  pubsub_target {
    topic_name = google_pubsub_topic.cron_topic.id
    data       = base64encode(jsonencode({ "from" = "cloud-scheduler" }))
  }
}
