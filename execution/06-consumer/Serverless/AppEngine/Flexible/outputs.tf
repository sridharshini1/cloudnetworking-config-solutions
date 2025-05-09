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

output "instance_application_urls" {
  description = "Map of instance keys (e.g., 'project_service') to the corresponding App Engine application URL output by that module instance. Should ideally be the same URL for instances targeting the same project."
  value = {
    for k, instance in module.flexible_app_engine_instance :
    k => instance.application_url
  }
  sensitive = false
}
output "instance_service_urls" {
  description = "Map of instance keys to the map of service URLs provided by that instance. Each inner map usually contains the URL for the single service defined in the corresponding YAML."
  value = {
    for k, instance in module.flexible_app_engine_instance :
    k => instance.service_urls
  }
}
output "instance_domain_mapping_resource_records" {
  description = "Map of instance keys to the domain mapping resource records output by that specific module instance."
  value = {
    for k, instance in module.flexible_app_engine_instance :
    k => instance.domain_mapping_resource_records
  }
  sensitive = true
}