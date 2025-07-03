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

resource "google_network_security_firewall_endpoint" "firewall_endpoint" {
  count              = var.create_firewall_endpoint ? 1 : 0
  name               = var.firewall_endpoint_name
  parent             = "organizations/${var.organization_id}"
  location           = var.location
  labels             = var.firewall_endpoint_labels
  billing_project_id = var.billing_project_id
}

resource "google_network_security_firewall_endpoint_association" "firewall_endpoint_association" {
  count                 = var.create_firewall_endpoint_association ? 1 : 0
  name                  = var.association_name
  parent                = "projects/${var.association_project_id}"
  location              = var.location
  labels                = var.association_labels
  disabled              = var.association_disabled
  network               = var.vpc_id
  firewall_endpoint     = local.firewall_endpoint_to_associate
  tls_inspection_policy = var.tls_inspection_policy_id
}