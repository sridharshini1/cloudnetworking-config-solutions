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

output "autoscaler" {
  description = "Auto-created autoscaler resource."
  value       = { for k, v in module.mig : k => v.autoscaler }
}

output "group_manager" {
  description = "Instance group resource."
  value       = { for k, v in module.mig : k => v.group_manager }
}

output "id" {
  description = "Fully qualified group manager id."
  value       = { for k, v in module.mig : k => v.id }
}

output "instance_template" {
  description = "The self-link of the instance template used for the MIG."
  value       = { for k, v in module.mig-template : k => try(v.template.self_link, null) }
}

output "autoscaler_config" {
  description = "Configuration details of the autoscaler."
  value = {
    for k, v in local.mig_map : k => {
      max_replicas    = try(v.autoscaler_config.max_replicas, null),
      min_replicas    = try(v.autoscaler_config.min_replicas, null),
      cooldown_period = try(v.autoscaler_config.cooldown_period, null),
      scaling_signals = {
        cpu_utilization = {
          target                = try(v.autoscaler_config.scaling_signals.cpu_utilization.target, null),
          optimize_availability = try(v.autoscaler_config.scaling_signals.cpu_utilization.optimize_availability, null)
        }
      }
    }
  }
}
