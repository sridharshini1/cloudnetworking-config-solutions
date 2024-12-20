
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable "bootstrap_project_id" {
  type        = string
  description = "Google Cloud Project ID which will be used to create the service account and Google Cloud storage buckets."
}

variable "folder_id" {
  type        = string
  description = "Google Cloud folder ID designating the parent folder for both the networking host project and the service project."
}

variable "network_hostproject_id" {
  type        = string
  description = "Google Cloud Project ID for the networking host project to be used to create networking and security resources."
}

variable "network_serviceproject_id" {
  type        = string
  description = "Google Cloud Project ID to be used to create Google Cloud resources like consumer and producer services."
}

variable "gcs_bucket_name" {
  type        = string
  description = "Name of the Google Cloud storage bucket."
  default     = "terraform-state"
}

variable "versioning" {
  type        = bool
  description = "The Goocle Cloud storage bucket versioning."
  default     = true
}

variable "gcs_bucket_location" {
  description = "Location of the Google Cloud storage bucket."
  type        = string
  default     = "EU"
}

variable "organization_sa_name" {
  type        = string
  description = "Name of the service account to create for organization stage."
  default     = "organization-sa"
}

variable "organization_administrator" {
  type        = list(string)
  description = "List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "networking_sa_name" {
  type        = string
  description = "Name of the service account to create for networking stage."
  default     = "networking-sa"
}

variable "networking_administrator" {
  type        = list(string)
  description = "List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "security_sa_name" {
  type        = string
  description = "Name of the service account to create for security stage."
  default     = "security-sa"
}

variable "security_administrator" {
  type        = list(string)
  description = "List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "producer_cloudsql_sa_name" {
  type        = string
  description = "Name of the service account to create for CloudSQL's producer stage."
  default     = "producer-cloudsql-sa"
}

variable "producer_cloudsql_administrator" {
  type        = list(string)
  description = "List of Cloud SQL administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "producer_alloydb_sa_name" {
  type        = string
  description = "Name of the service account to create for AlloyDB's producer stage."
  default     = "producer-alloydb-sa"
}

variable "producer_alloydb_administrator" {
  type        = list(string)
  description = "List of AlloyDB administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "producer_mrc_sa_name" {
  type        = string
  description = "Name of the service account to create for MRC's producer stage."
  default     = "producer-mrc-sa"
}

variable "producer_mrc_administrator" {
  type        = list(string)
  description = "List of MRC administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "producer_vertex_sa_name" {
  type        = string
  description = "Name of the service account to create for Vertex AI's producer stage."
  default     = "producer-vertex-sa"
}

variable "producer_vertex_administrator" {
  type        = list(string)
  description = "List of Vertex AI administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "producer_gke_sa_name" {
  type        = string
  description = "Name of the service account to create for GKE's producer stage."
  default     = "producer-gke-sa"
}

variable "producer_gke_administrator" {
  type        = list(string)
  description = "List of GKE administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "networking_manual_sa_name" {
  type        = string
  description = "Name of the service account to create for networking manual stage."
  default     = "networking-manual-sa"
}

variable "networking_manual_administrator" {
  type        = list(string)
  description = "List of Members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "consumer_gce_sa_name" {
  type        = string
  description = "Name of the service account to create for GCE consumer stage."
  default     = "consumer-gce-sa"
}

variable "consumer_gce_administrator" {
  type        = list(string)
  description = "List of GCE administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}

variable "consumer_cloudrun_sa_name" {
  type        = string
  description = "Name of the service account to create for Cloud Run consumer stage."
  default     = "consumer-cloudrun-sa"
}

variable "consumer_cloudrun_administrator" {
  type        = list(string)
  description = "List of Cloud Run administrative members to be granted an IAM role. e.g. (group:my-group@example.com),(user:my-user@example.com)"
  default     = [""]
}
