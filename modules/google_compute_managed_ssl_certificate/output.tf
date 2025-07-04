/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

output "certificate_id" {
  description = "The unique numerical identifier for the managed SSL certificate resource, assigned by Google Cloud."
  value       = google_compute_managed_ssl_certificate.ssl_cert.certificate_id
}

output "creation_timestamp" {
  description = "The timestamp in RFC3339 format indicating when this SSL certificate resource was created."
  value       = google_compute_managed_ssl_certificate.ssl_cert.creation_timestamp
}

output "expire_time" {
  description = "The timestamp in RFC3339 format indicating when this SSL certificate will expire. For 'MANAGED' type certificates, Google automatically renews the certificate before this time."
  value       = google_compute_managed_ssl_certificate.ssl_cert.expire_time
}

output "id" {
  description = "The fully qualified identifier (ID) of the managed SSL certificate resource, typically in the format 'projects/PROJECT_ID/global/sslCertificates/CERTIFICATE_NAME'."
  value       = google_compute_managed_ssl_certificate.ssl_cert.id
}

output "project" {
  description = "The ID of the Google Cloud project in which the managed SSL certificate was created."
  value       = google_compute_managed_ssl_certificate.ssl_cert.project
}

output "self_link" {
  description = "The self-referential URI of the created managed SSL certificate resource."
  value       = google_compute_managed_ssl_certificate.ssl_cert.self_link
}

output "subject_alternative_names" {
  description = "A list of Subject Alternative Names (SANs) that are secured by this SSL certificate. These usually correspond to the domains provided in the 'managed' block."
  value       = google_compute_managed_ssl_certificate.ssl_cert.subject_alternative_names
}

output "ssl_cert" {
  description = "All attributes of the created `google_compute_managed_ssl_certificate` resource. This output provides the full object, allowing access to any of its properties."
  value       = google_compute_managed_ssl_certificate.ssl_cert
}