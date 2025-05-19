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

output "nlb_forwarding_rule_addresses" {
  description = "Map of Network Load Balancer names to their forwarding rule names and IP addresses."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.forwarding_rule_addresses
  }
}

output "nlb_forwarding_rule_self_links" {
  description = "Map of Network Load Balancer names to their forwarding rule names and self links."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.forwarding_rule_self_links
  }
}

output "nlb_backend_service_self_links" {
  description = "Map of Network Load Balancer names to their backend service self links."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.backend_service_self_link
  }
}

output "nlb_health_check_self_links" {
  description = "Map of Network Load Balancer names to their auto-created health check self links (if applicable)."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.health_check_self_link
  }
}

output "nlb_forwarding_rules" {
  description = "Detailed forwarding rule resources for each NLB."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.forwarding_rules
  }
}

output "nlb_backend_services" {
  description = "Detailed backend service resources for each NLB."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.backend_service
  }
}

output "nlb_health_checks" {
  description = "Detailed auto-created health check resources for each NLB (if applicable)."
  value = {
    for lb_name, lb in module.nlb_passthrough_ext :
    lb_name => lb.health_check
  }
}