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

module "firewall_endpoints" {
  for_each                             = local.firewall_endpoints_map
  source                               = "../../../modules/firewall_endpoint"
  location                             = each.value.location
  create_firewall_endpoint             = each.value.firewall_endpoint.create_firewall_endpoint
  organization_id                      = each.value.firewall_endpoint.organization_id
  billing_project_id                   = each.value.firewall_endpoint.billing_project_id
  firewall_endpoint_name               = each.value.firewall_endpoint.name
  firewall_endpoint_labels             = each.value.firewall_endpoint.labels
  create_firewall_endpoint_association = each.value.firewall_endpoint_association.create_firewall_endpoint_association
  association_name                     = each.value.firewall_endpoint_association.name
  association_project_id               = each.value.firewall_endpoint_association.association_project_id
  vpc_id                               = each.value.firewall_endpoint_association.vpc_id
  tls_inspection_policy_id             = each.value.firewall_endpoint_association.tls_inspection_policy_id
  association_labels                   = each.value.firewall_endpoint_association.labels
  association_disabled                 = each.value.firewall_endpoint_association.disabled
  existing_firewall_endpoint_id        = each.value.firewall_endpoint_association.existing_firewall_endpoint_id
}