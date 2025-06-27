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

output "firewall_endpoint_id" {
  description = "The full resource ID of the created Firewall Endpoint."
  value       = length(google_network_security_firewall_endpoint.firewall_endpoint) > 0 ? google_network_security_firewall_endpoint.firewall_endpoint[0].id : null
}

output "firewall_endpoint_name" {
  description = "The name of the created Firewall Endpoint."
  value       = length(google_network_security_firewall_endpoint.firewall_endpoint) > 0 ? google_network_security_firewall_endpoint.firewall_endpoint[0].name : null
}

output "firewall_endpoint_self_link" {
  description = "The self-link of the created Firewall Endpoint."
  value       = length(google_network_security_firewall_endpoint.firewall_endpoint) > 0 ? google_network_security_firewall_endpoint.firewall_endpoint[0].self_link : null
}

output "association_id" {
  description = "The full resource ID of the created Firewall Endpoint Association."
  value       = length(google_network_security_firewall_endpoint_association.firewall_endpoint_association) > 0 ? google_network_security_firewall_endpoint_association.firewall_endpoint_association[0].id : null
}

output "association_name" {
  description = "The name of the created Firewall Endpoint Association."
  value       = length(google_network_security_firewall_endpoint_association.firewall_endpoint_association) > 0 ? google_network_security_firewall_endpoint_association.firewall_endpoint_association[0].name : null
}