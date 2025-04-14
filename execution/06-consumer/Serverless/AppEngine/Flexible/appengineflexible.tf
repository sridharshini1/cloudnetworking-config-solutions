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

module "flexible_app_engine_instance" {
  for_each                = local.instance_data_map # Use the map of processed configurations
  source                  = "../../../../../modules/app_engine/flexible"
  project_id              = each.value.project_id
  location_id             = each.value.location_id
  auth_domain             = each.value.auth_domain
  database_type           = each.value.database_type
  serving_status          = each.value.app_serving_status
  feature_settings        = each.value.feature_settings
  iap                     = each.value.iap
  dispatch_rules          = each.value.dispatch_rules
  domain_mappings         = each.value.domain_mappings
  firewall_rules          = each.value.firewall_rules
  services                = each.value.services_for_module
  create_application      = each.value.create_application
  create_dispatch_rules   = each.value.create_dispatch_rules
  create_domain_mappings  = each.value.create_domain_mappings
  create_firewall_rules   = each.value.create_firewall_rules
  create_network_settings = each.value.create_network_settings
  create_split_traffic    = each.value.create_split_traffic
}