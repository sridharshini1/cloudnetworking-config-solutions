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

locals {
  firewall_configs_data = [
    for file_path in fileset(var.config_folder_path, "*.y*ml") : {
      key     = trimsuffix(trimsuffix(basename(file_path), ".yaml"), ".yml")
      content = yamldecode(file("${var.config_folder_path}/${file_path}"))
    }
  ]
  firewall_endpoints_list = flatten([
    for item in local.firewall_configs_data : {
      key      = item.key
      location = try(item.content.location, var.location)
      firewall_endpoint = {
        create_firewall_endpoint = try(item.content.firewall_endpoint.create, var.create_firewall_endpoint)
        name                     = try(item.content.firewall_endpoint.name, var.firewall_endpoint_name)
        organization_id          = try(item.content.firewall_endpoint.organization_id, var.organization_id)
        billing_project_id       = try(item.content.firewall_endpoint.billing_project_id, var.billing_project_id)
        labels                   = try(item.content.firewall_endpoint.labels, var.firewall_endpoint_labels)
      }
      firewall_endpoint_association = {
        create_firewall_endpoint_association = try(item.content.firewall_endpoint_association.create, var.create_firewall_endpoint_association)
        name                                 = try(item.content.firewall_endpoint_association.name, var.association_name)
        association_project_id               = try(item.content.firewall_endpoint_association.association_project_id, var.association_project_id)
        vpc_id                               = try(item.content.firewall_endpoint_association.vpc_id, var.vpc_id)
        tls_inspection_policy_id             = try(item.content.firewall_endpoint_association.tls_inspection_policy_id, var.tls_inspection_policy_id)
        labels                               = try(item.content.firewall_endpoint_association.labels, var.association_labels)
        disabled                             = try(item.content.firewall_endpoint_association.disabled, var.association_disabled)
        existing_firewall_endpoint_id        = try(item.content.firewall_endpoint_association.existing_firewall_endpoint_id, var.existing_firewall_endpoint_id)
      }
    }
  ])
  firewall_endpoints_map = { for fe in local.firewall_endpoints_list : fe.key => fe if fe.key != null }
}