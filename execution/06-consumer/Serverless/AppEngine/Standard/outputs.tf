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
  description = "Map of instance keys to the default URL of the App Engine application associated with that instance."
  value = {
    for key, instance in module.appengine_standard_instance : key => instance.application_url
  }
}

output "instance_service_urls" {
  description = "Map of instance keys to the map of service names to their URLs for that instance."
  value = {
    for key, instance in module.appengine_standard_instance : key => instance.service_urls
  }
}

output "instance_domain_mapping_details" {
  description = "Map of instance keys to the domain mapping resource records created by that instance."
  value = {
    for key, instance in module.appengine_standard_instance : key => instance.domain_mapping_resource_records
  }
}

output "instance_vpc_connector_names" {
  description = "Map of instance keys to the name of the VPC Access Connector created by that instance (if any)."
  value = {
    for key, instance in module.appengine_standard_instance : key => instance.vpc_connector_name
  }
}

output "instance_vpc_connector_self_links" {
  description = "Map of instance keys to the self-link of the VPC Access Connector created by that instance (if any)."
  value = {
    for key, instance in module.appengine_standard_instance : key => instance.vpc_connector_self_link
  }
}

output "all_instance_details" {
  description = "Map of instance keys to an object containing key outputs for that instance."
  value = {
    for key, instance in module.appengine_standard_instance : key => {
      application_url = instance.application_url
      service_urls    = instance.service_urls
      vpc_connector   = instance.vpc_connector_name
      domain_mappings = instance.domain_mapping_resource_records # Contains the full object might be large
    }
  }
}