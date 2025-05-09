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
  lb_instances       = [for file in fileset(local.config_folder_path, "[^_]*.yaml") : yamldecode(file("${local.config_folder_path}/${file}"))]

  lb_instance_list = flatten([
    for lb_instance in try(local.lb_instances, []) : {
      name    = lb_instance.name
      project = lb_instance.project
      network = lb_instance.network

      backends = {
        default = {
          protocol    = try(lb_instance.backends.default.protocol, var.backend_protocol)
          port        = try(lb_instance.backends.default.port, var.backend_port)
          port_name   = try(lb_instance.backends.default.port_name, var.backend_port_name)
          timeout_sec = try(lb_instance.backends.default.timeout_sec, var.backend_timeout_sec)
          enable_cdn  = try(lb_instance.backends.default.enable_cdn, var.enable_cdn)

          health_check = try(lb_instance.backends.default.health_check, var.health_check)
          log_config   = try(lb_instance.backends.default.log_config, var.log_config)

          groups = [
            {
              group  = try(lb_instance.backends.default.groups[0].group, null)
              region = try(lb_instance.backends.default.groups[0].region, null)
            },
          ]

          iap_config = try(lb_instance.backends.default.iap_config, var.iap_config)
        }
      }
    }
  ])

  lb_map = { for lb_instance in local.lb_instance_list :
    lb_instance.name => merge(
      lb_instance,
      {
        backends = {
          default = merge(
            lb_instance.backends.default,
            {
              groups = [
                {
                  group = format("projects/%s/regions/%s/instanceGroups/%s",
                    lb_instance.project,
                    lb_instance.backends.default.groups[0].region,
                  lb_instance.backends.default.groups[0].group)
                }
              ]
            }
          )
        }
      }
    )
  }
}