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
  config_folder_path = var.config_folder_path
  raw_app_configs = [
    for file in fileset(local.config_folder_path, "[^_]*.yaml") : {
      filename = file
      config   = yamldecode(file("${local.config_folder_path}/${file}"))
    }
  ]
  app_engine_instances = {
    for config_item in local.raw_app_configs :
    trimsuffix(config_item.filename, ".yaml") => {
      project                       = config_item.config.project_id
      service                       = config_item.config.service
      runtime                       = config_item.config.runtime
      location_id                   = try(config_item.config.app_engine_application.location_id, var.location_id)
      version_id                    = try(config_item.config.version_id, var.version_id)
      deployment                    = try(config_item.config.deployment, var.deployment)
      entrypoint                    = try(config_item.config.entrypoint, var.entrypoint)
      app_engine_apis               = try(config_item.config.app_engine_apis, var.app_engine_apis)
      runtime_api_version           = try(config_item.config.runtime_api_version, var.runtime_api_version)
      service_account               = try(config_item.config.service_account, var.service_account)
      threadsafe                    = try(config_item.config.threadsafe, var.threadsafe)
      inbound_services              = try(config_item.config.inbound_services, var.inbound_services)
      instance_class                = try(config_item.config.instance_class, var.instance_class)
      labels                        = try(config_item.config.labels, var.labels)
      delete_service_on_destroy     = try(config_item.config.delete_service_on_destroy, var.delete_service_on_destroy)
      noop_on_destroy               = try(config_item.config.noop_on_destroy, var.noop_on_destroy)
      env_variables                 = try(config_item.config.env_variables, var.env_variables)
      handlers                      = try(config_item.config.handlers, var.handlers)
      libraries                     = try(config_item.config.libraries, var.libraries)
      automatic_scaling             = try(config_item.config.automatic_scaling, var.automatic_scaling)
      basic_scaling                 = try(config_item.config.basic_scaling, var.basic_scaling)
      manual_scaling                = try(config_item.config.manual_scaling, var.manual_scaling)
      create_vpc_connector          = try(config_item.config.create_vpc_connector, var.create_vpc_connector)
      vpc_access_connector          = try(config_item.config.vpc_access_connector, var.vpc_access_connector)
      vpc_connector_details         = try(config_item.config.vpc_connector_details, var.vpc_connector_details)
      create_network_settings       = try(config_item.config.create_network_settings, var.create_network_settings)
      network_settings              = try(config_item.config.network_settings, var.network_settings)
      create_split_traffic          = try(config_item.config.create_split_traffic, var.create_split_traffic)
      split_traffic                 = try(config_item.config.split_traffic, var.split_traffic)
      create_app_engine_application = try(config_item.config.create_app_engine_application, var.create_app_engine_application)
      app_engine_application        = try(config_item.config.app_engine_application, var.app_engine_application) # Default empty map
      create_dispatch_rules         = try(config_item.config.create_dispatch_rules, var.create_dispatch_rules)
      dispatch_rules                = try(config_item.config.dispatch_rules, var.dispatch_rules)
      create_domain_mappings        = try(config_item.config.create_domain_mappings, var.create_domain_mappings)
      domain_mappings               = try(config_item.config.domain_mappings, var.domain_mappings)
      create_firewall_rules         = try(config_item.config.create_firewall_rules, var.create_firewall_rules)
      firewall_rules                = try(config_item.config.firewall_rules, var.firewall_rules)
      create_app_version            = try(config_item.config.create_app_version, var.create_app_version)
      auth_domain                   = try(config_item.config.app_engine_application.auth_domain, var.auth_domain)
      database_type                 = try(config_item.config.app_engine_application.database_type, var.database_type)
      serving_status                = try(config_item.config.app_engine_application.serving_status, var.serving_status)
      feature_settings              = try(config_item.config.app_engine_application.feature_settings, var.feature_settings)
      iap                           = try(config_item.config.app_engine_application.iap, var.iap)
    } if config_item.config != null
  }
}
