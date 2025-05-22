/**
 * Copyright 2024-25 Google LLC
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

output "ncc_hub" {
  description = "The NCC Hub object"
  value       = google_network_connectivity_hub.hub
}

output "vpc_spokes" {
  description = "All created vpc spoke resource objects"
  value       = google_network_connectivity_spoke.vpc_spoke
}

output "hybrid_spokes" {
  description = "All created hybrid spoke resource objects"
  value       = google_network_connectivity_spoke.hybrid_spoke
}

output "router_appliance_spokes" {
  description = "All created router appliance spoke resource objects"
  value       = google_network_connectivity_spoke.router_appliance_spoke
}

output "producer_vpc_spokes" {
  description = "All created producer VPC spoke resource objects"
  value       = google_network_connectivity_spoke.producer_vpc_spoke
}