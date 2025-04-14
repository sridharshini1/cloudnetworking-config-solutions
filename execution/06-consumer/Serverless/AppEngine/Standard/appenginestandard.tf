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

module "appengine_standard_instance" {
  for_each                      = local.app_engine_instances # Iterate over the map of instances
  source                        = "../../../../../modules/app_engine/standard"
  project_id                    = each.value.project
  location_id                   = each.value.location_id
  dispatch_rules                = each.value.dispatch_rules
  domain_mappings               = each.value.domain_mappings
  firewall_rules                = each.value.firewall_rules
  create_app_engine_application = each.value.create_app_engine_application
  create_dispatch_rules         = each.value.create_dispatch_rules
  create_domain_mappings        = each.value.create_domain_mappings
  create_firewall_rules         = each.value.create_firewall_rules
  create_network_settings       = each.value.create_network_settings
  create_split_traffic          = each.value.create_split_traffic
  create_app_version            = each.value.create_app_version
  auth_domain                   = each.value.auth_domain
  database_type                 = each.value.database_type
  serving_status                = each.value.serving_status
  feature_settings              = each.value.feature_settings
  iap                           = each.value.iap
  services = {
    "${each.value.service}" = {
      service                   = each.value.service
      runtime                   = each.value.runtime
      version_id                = each.value.version_id
      deployment                = each.value.deployment
      entrypoint                = each.value.entrypoint
      app_engine_apis           = each.value.app_engine_apis
      runtime_api_version       = each.value.runtime_api_version
      service_account           = each.value.service_account
      threadsafe                = each.value.threadsafe
      inbound_services          = each.value.inbound_services
      instance_class            = each.value.instance_class
      labels                    = each.value.labels
      delete_service_on_destroy = each.value.delete_service_on_destroy
      noop_on_destroy           = each.value.noop_on_destroy
      env_variables             = each.value.env_variables
      handlers                  = each.value.handlers
      libraries                 = each.value.libraries
      automatic_scaling         = each.value.automatic_scaling
      basic_scaling             = each.value.basic_scaling
      manual_scaling            = each.value.manual_scaling
      create_vpc_connector      = each.value.create_vpc_connector
      vpc_access_connector      = each.value.vpc_access_connector
      vpc_connector_details     = each.value.vpc_connector_details
      create_network_settings   = each.value.create_network_settings
      network_settings          = each.value.network_settings
      create_split_traffic      = each.value.create_split_traffic
      split_traffic             = each.value.split_traffic
      create_dispatch_rules     = each.value.create_dispatch_rules
    }
  }
}
