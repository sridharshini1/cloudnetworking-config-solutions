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
  instances          = [for file in fileset(local.config_folder_path, "[^_]*.yaml") : yamldecode(file("${local.config_folder_path}/${file}"))]
  instance_list = flatten([
    for instance in try(local.instances, []) : {
      project_id      = instance.project_id
      name            = instance.name
      region          = instance.region
      network         = try(instance.network, var.network)
      subnet_name     = try(instance.subnet_name, var.subnet_name)
      ip_cidr_range   = try(instance.ip_cidr_range, var.ip_cidr_range)
      host_project_id = try(instance.host_project_id, var.host_project_id)
      machine_type    = try(instance.machine_type, var.machine_type)
      min_instances   = try(instance.min_instances, var.min_instances)
      max_instances   = try(instance.max_instances, var.max_instances)
      max_throughput  = try(instance.max_throughput, var.max_throughput)
      min_throughput  = try(instance.min_throughput, var.min_throughput)
    }
  ])
}