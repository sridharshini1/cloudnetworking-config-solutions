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

output "load_balancer_ips" {
  value       = { for lb_name, lb in module.lb_http : lb_name => lb.external_ip }
  description = "A map of load balancer names to their external IP addresses."
}

output "load_balancer_ipv6s" {
  value = {
    for lb_name, lb in module.lb_http :
    lb_name => (lb.ipv6_enabled ? lb.external_ipv6_address : "undefined")
  }
  description = "A map of load balancer names to their IPv6 addresses, if enabled; else \"undefined\"."
}

output "load_balancers" {
  description = "Detailed information about each load balancer."
  value = {
    for lb_name, lb in module.lb_http :
    lb_name => {
      external_ip           = lb.external_ip
      external_ipv6_address = lb.external_ipv6_address
      ipv6_enabled          = lb.ipv6_enabled
      http_proxy            = lb.http_proxy
      https_proxy           = lb.https_proxy
      url_map               = lb.url_map
      backend_services = [
        for service in lb.backend_services :
        {
          name      = service.name
          self_link = service.self_link
        }
      ]
    }
  }
}
