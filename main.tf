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

variable "netlify_token" {}

provider "google" {
  credentials = base64decode(var.google_service_account)

  project = "merge-chance-time"
  region  = "asia-northeast1"
  zone    = "asia-northeast1-a"
}

provider "netlify" {
  token = var.netlify_token
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

resource "google_project_iam_binding" "pubsub" {
  project = data.google_project.current.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  members = [
    "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
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

resource "google_pubsub_topic" "update_chance_topic" {
  name = "update-chance"
}

resource "google_pubsub_subscription" "update_chance_subscription" {
  name  = "update-chance"
  topic = google_pubsub_topic.update_chance_topic.name
  push_config {
    push_endpoint = "https://${google_app_engine_application.app.default_hostname}/cron"
    oidc_token {
      service_account_email = "${google_app_engine_application.app.id}@appspot.gserviceaccount.com"
    }
  }
}

resource "google_cloud_scheduler_job" "update_chance" {
  name     = "update-chance"
  schedule = "0 * * * *"
  pubsub_target {
    topic_name = google_pubsub_topic.update_chance_topic.id
    data       = base64encode(jsonencode({ "from" = "cloud-scheduler" }))
  }
}

resource "netlify_site" "admin" {
  name          = "merge-chance-time"
  custom_domain = trimsuffix(google_dns_record_set.root.name, ".")

  repo {
    provider    = "github"
    repo_path   = "aereal/merge-chance-time"
    repo_branch = "master"
    command     = "yarn build"
    dir         = "./front/web/build"
  }
}

resource "google_dns_managed_zone" "mergechancetime_app" {
  name     = "mergechancetime-app"
  dns_name = "mergechancetime.app."
}

resource "google_dns_record_set" "root" {
  name         = google_dns_managed_zone.mergechancetime_app.dns_name
  type         = "A"
  ttl          = 300
  managed_zone = google_dns_managed_zone.mergechancetime_app.name
  rrdatas      = ["104.198.14.52"]
}

resource "google_dns_record_set" "www" {
  name         = "www.${google_dns_managed_zone.mergechancetime_app.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = google_dns_managed_zone.mergechancetime_app.name
  rrdatas      = ["104.198.14.52"]
}
