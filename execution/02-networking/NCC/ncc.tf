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

module "network_connectivity_center" {
  for_each                        = local.instance_list
  source                          = "../../../modules/network-connectivity-center"
  project_id                      = each.value.project_id
  create_new_hub                  = each.value.create_new_hub
  existing_hub_uri                = each.value.existing_hub_uri
  spoke_labels                    = each.value.spoke_labels
  export_psc                      = each.value.export_psc
  ncc_hub_name                    = each.value.ncc_hub_name # Hub configuration
  ncc_hub_description             = each.value.ncc_hub_description
  ncc_hub_labels                  = each.value.ncc_hub_labels
  policy_mode                     = each.value.policy_mode
  preset_topology                 = each.value.preset_topology
  auto_accept_projects            = each.value.auto_accept_projects
  group_name                      = each.value.group_name
  group_decription                = each.value.group_decription
  vpc_spokes                      = each.value.vpc_spokes                      # VPC spoke configuration
  producer_vpc_spokes             = each.value.producer_vpc_spokes             # Producer VPC spoke configuration
  linked_interconnect_attachments = each.value.linked_interconnect_attachments # Hybrid spoke (Interconnect) configuration
  linked_vpn_tunnels              = each.value.linked_vpn_tunnels              # Hybrid spoke (vpn  tunnel) configuration
  router_appliance_spokes         = each.value.router_appliance_spokes         # Router Applicance configuration
}
