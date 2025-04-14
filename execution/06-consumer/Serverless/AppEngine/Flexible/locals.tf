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
  raw_instances = [
    for file in fileset(local.config_folder_path, "[^_]*.yaml") :
    yamldecode(file("${local.config_folder_path}/${file}"))
  ]
  processed_instance_list = [
    for instance_config in local.raw_instances : {
      project_id         = instance_config.project_id
      service_key        = instance_config.service # Used for the key in instance_data_map
      location_id        = try(instance_config.app_engine_application.location_id, var.location_id)
      auth_domain        = try(instance_config.app_engine_application.auth_domain, var.auth_domain)
      database_type      = try(instance_config.app_engine_application.database_type, var.database_type)
      app_serving_status = try(instance_config.app_engine_application.serving_status, var.serving_status)
      feature_settings   = try(instance_config.app_engine_application.feature_settings, var.feature_settings)
      iap                = try(instance_config.app_engine_application.iap, var.iap_settings)
      dispatch_rules     = try(instance_config.dispatch_rules, var.dispatch_rules)
      domain_mappings    = try(instance_config.domain_mappings, var.domain_mappings)
      firewall_rules     = try(instance_config.firewall_rules, var.firewall_rules)
      services_for_module = {
        (instance_config.service) = {
          service                      = instance_config.service
          runtime                      = instance_config.runtime
          version_id                   = try(instance_config.version_id, var.version_id)
          instance_class               = try(instance_config.instance_class, var.instance_class)
          flexible_runtime_settings    = try(instance_config.flexible_runtime_settings, var.flexible_runtime_settings)
          network                      = try(instance_config.network, var.network)
          resources                    = try(instance_config.resources, var.resources)
          entrypoint                   = try(instance_config.entrypoint, var.entrypoint)
          automatic_scaling            = try(instance_config.automatic_scaling, var.automatic_scaling)
          manual_scaling               = try(instance_config.manual_scaling, var.manual_scaling)
          env_variables                = try(instance_config.env_variables, var.env_variables)
          deployment                   = try(instance_config.deployment, var.deployment)
          liveness_check               = instance_config.liveness_check
          readiness_check              = instance_config.readiness_check
          service_account              = try(instance_config.service_account, var.service_account)
          endpoints_api_service        = try(instance_config.endpoints_api_service, var.endpoints_api_service)
          nobuild_files_regex          = try(instance_config.nobuild_files_regex, var.nobuild_files_regex)
          beta_settings                = try(instance_config.beta_settings, var.beta_settings)
          inbound_services             = try(instance_config.inbound_services, var.inbound_services)
          labels                       = try(instance_config.labels, var.labels)
          serving_status               = try(instance_config.serving_status, var.service_serving_status)
          runtime_api_version          = try(instance_config.runtime_api_version, var.runtime_api_version)
          runtime_channel              = try(instance_config.runtime_channel, var.runtime_channel)
          runtime_main_executable_path = try(instance_config.runtime_main_executable_path, var.runtime_main_executable_path)
          delete_service_on_destroy    = try(instance_config.delete_service_on_destroy, var.delete_service_on_destroy)
          noop_on_destroy              = try(instance_config.noop_on_destroy, var.noop_on_destroy)
          network_settings             = try(instance_config.network_settings, var.network_settings_block)
          split_traffic                = try(instance_config.split_traffic, var.split_traffic_block)
        }
      }
      create_application      = try(instance_config.create_application, var.create_application)
      create_dispatch_rules   = try(instance_config.create_dispatch_rules, var.create_dispatch_rules)
      create_domain_mappings  = try(instance_config.create_domain_mappings, var.create_domain_mappings)
      create_firewall_rules   = try(instance_config.create_firewall_rules, var.create_firewall_rules)
      create_network_settings = try(instance_config.create_network_settings, var.create_network_settings)
      create_split_traffic    = try(instance_config.create_split_traffic, var.create_split_traffic)
    }
  ]
  instance_data_map = {
    for config in local.processed_instance_list :
    "${config.project_id}_${config.service_key}" => config
  }
}