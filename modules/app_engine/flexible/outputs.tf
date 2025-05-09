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

output "application_url" {
  value       = var.create_application ? google_app_engine_application.app[0].default_hostname : null
  description = "The default URL of the App Engine application. URLs are empty strings if 'create_application' is false."
}
output "service_urls" {
  value = {
    for service_name, service_config in var.services :
    service_name => var.create_application ? "https://${service_config.version_id}-dot-${service_name}-dot-${google_app_engine_application.app[0].default_hostname}" : "https://${service_config.version_id}-dot-${service_name}-dot-${var.project_id}.appspot.com"
  }
  description = "A map of service names to their URLs. URLs are empty strings if 'create_application' is false."
}
output "domain_mapping_resource_records" {
  value       = google_app_engine_domain_mapping.mapping
  description = "all domain mapping resource records"
}