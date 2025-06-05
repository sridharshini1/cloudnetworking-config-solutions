# Copyright 2025 Google LLC
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

output "managed_ssl_certificate_id" {
  description = "The unique identifier for the managed SSL certificate resource."
  value       = module.ssl_certificate.certificate_id
}

output "managed_ssl_certificate_creation_timestamp" {
  description = "Creation timestamp of the managed SSL certificate."
  value       = module.ssl_certificate.creation_timestamp
}

output "managed_ssl_certificate_expire_time" {
  description = "Expire time of the managed SSL certificate."
  value       = module.ssl_certificate.expire_time
}

output "managed_ssl_certificate_full_id" {
  description = "The full ID of the managed SSL certificate."
  value       = module.ssl_certificate.id
}

output "managed_ssl_certificate_project" {
  description = "The project in which the managed SSL certificate was created."
  value       = module.ssl_certificate.project
}

output "managed_ssl_certificate_self_link" {
  description = "The self link of the managed SSL certificate."
  value       = module.ssl_certificate.self_link
}

output "managed_ssl_certificate_subject_alternative_names" {
  description = "Subject Alternative Names (SANs) of the managed SSL certificate."
  value       = module.ssl_certificate.subject_alternative_names
}

output "managed_ssl_certificate_details" {
  description = "All attributes of the created google_compute_managed_ssl_certificate."
  value       = module.ssl_certificate.ssl_cert
}