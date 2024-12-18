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

module "network_connectivity_center" {
  count          = var.create_ncc ? 1 : 0
  source         = "../../modules/network-connectivity-center"
  project_id     = var.project_id
  ncc_hub_name   = var.ncc_hub_name
  ncc_hub_labels = var.ncc_hub_labels
  spoke_labels   = var.spoke_labels
  vpc_spokes     = local.vpc_spokes
  depends_on     = [module.vpc_network]
}
