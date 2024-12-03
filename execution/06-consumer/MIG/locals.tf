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
  config_folder_path = var.config_folder_path

  # Reading YAML files for MIG configurations
  migs_config = [
    for file in fileset(local.config_folder_path, "[^_]*.yaml") :
    yamldecode(file("${local.config_folder_path}/${file}"))
  ]

  # Flattening the MIG configurations into a list
  mig_list = flatten([
    for mig in try(local.migs_config, []) : {
      name        = mig.name
      project_id  = mig.project_id
      location    = mig.location
      target_size = try(mig.target_size, null)

      # Extracting VPC and Subnetwork names from YAML
      vpc_name        = try(mig.vpc_name)
      subnetwork_name = try(mig.subnetwork_name)

      auto_healing_policies = {
        health_check      = try(mig.auto_healing_policies.health_check, null)
        initial_delay_sec = try(mig.auto_healing_policies.initial_delay_sec, 30)
      }

      autoscaler_config = {
        max_replicas    = try(mig.autoscaler_config.max_replicas, 3)
        min_replicas    = try(mig.autoscaler_config.min_replicas, 1)
        cooldown_period = try(mig.autoscaler_config.cooldown_period, null)
        scaling_signals = {
          cpu_utilization = {
            target                = try(mig.autoscaler_config.scaling_signals.cpu_utilization.target, 0.65)
            optimize_availability = try(mig.autoscaler_config.scaling_signals.cpu_utilization.optimize_availability, false)
          }
        }
      }

      health_check_config = {
        enable_logging = try(mig.health_check_config.enable_logging, true)
        tcp            = try(mig.health_check_config.tcp, null)   # TCP health check settings
        http           = try(mig.health_check_config.http, null)  # HTTP health check settings
        https          = try(mig.health_check_config.https, null) # HTTPS health check settings
        http2          = try(mig.health_check_config.http2, null) # HTTP2 health check settings
        grpc           = try(mig.health_check_config.grpc, null)  # gRPC health check settings
        ssl            = try(mig.health_check_config.ssl, null)   # SSL health check settings
      }

      description         = try(mig.description, "Terraform managed.")
      distribution_policy = try(mig.distribution_policy, null)
    }
  ])

  # Creating a map for easy access to MIG configurations by name
  mig_map = { for mig in local.mig_list : mig.name => mig }

  # Constructing self-links for VPC and Subnetwork based on each MIG's configuration
  vpc_self_links        = { for mig in local.mig_list : mig.name => "projects/${mig.project_id}/global/networks/${mig.vpc_name}" }
  subnetwork_self_links = { for mig in local.mig_list : mig.name => "projects/${mig.project_id}/regions/${var.region}/subnetworks/${mig.subnetwork_name}" }
}