provider "google" {
  credentials = file("service-account.json")

  project = "merge-chance-time"
  region = "asia-northeast1"
  zone = "asia-northeast1-a"
}

resource "google_app_engine_application" "app" {
  project = "merge-chance-time"
  location_id = "asia-northeast1"
}
