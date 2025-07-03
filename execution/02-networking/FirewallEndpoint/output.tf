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

output "firewall_endpoints" {
  description = "A map of all created firewall endpoints with their details."
  value = {
    for key, endpoint in module.firewall_endpoints : key => {
      id   = endpoint.firewall_endpoint_id
      name = endpoint.firewall_endpoint_name
    } if endpoint.firewall_endpoint_id != null
  }
}

output "firewall_endpoint_associations" {
  description = "A map of all created firewall endpoint associations with their details."
  value = {
    for key, associations in module.firewall_endpoints : key => {
      id   = associations.association_id
      name = associations.association_name
    } if associations.association_id != null
  }
}