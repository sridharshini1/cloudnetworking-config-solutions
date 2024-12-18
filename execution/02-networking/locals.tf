# Copyright 2024 Google LLC
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
  network_name    = try(module.vpc_network.name, "")
  network_id      = try(module.vpc_network.id, "")
  nat_router_name = "${var.nat_name}-route"
  // Subnets for SCP
  subnet_self_links_for_scp_policy = [
    for subnet in module.vpc_network.subnets :
    subnet.self_link
    if contains(var.subnets_for_scp_policy, subnet.name)
  ]
  vlan_attachment_project_id = var.project_id
  interconnect_project_id    = var.interconnect_project_id
  first_interconnect_name    = var.first_interconnect_name
  second_interconnect_name   = var.second_interconnect_name
  vpc_spoke1 = {
    "${var.vpc_spoke1}" = {
      uri = local.network_id
    }
  }
  vpc_spokes = merge(local.vpc_spoke1, var.existing_vpc_spoke)
}


