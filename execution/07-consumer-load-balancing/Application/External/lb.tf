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

module "lb_http" {
  for_each = local.lb_map
  source   = "../../../../modules/lb_http/"

  # Basic Load Balancer Configuration
  name    = each.value.name
  project = each.value.project

  # Backend Configuration
  backends = {
    default = {
      protocol    = each.value.backends.default.protocol
      port        = each.value.backends.default.port
      port_name   = each.value.backends.default.port_name
      timeout_sec = each.value.backends.default.timeout_sec
      enable_cdn  = each.value.backends.default.enable_cdn

      health_check = each.value.backends.default.health_check
      log_config   = each.value.backends.default.log_config

      groups = [
        {
          group = try(local.lb_map[each.key].backends.default.groups[0].group, var.instance_group)
        },
      ]
      iap_config = each.value.backends.default.iap_config
    }
  }
}
