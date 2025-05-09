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
  # Determine if ANY service requires connector creation.
  create_connector = anytrue([for s in var.services : try(s.create_vpc_connector, false)])

  # Get the FIRST service that defines vpc_connector_details, for defaults.
  first_connector_details = try(flatten([
    for s in var.services : [
      {
        name              = try(s.vpc_connector_details.name, "appengine-connector-${try(s.vpc_connector_details.region, var.location_id)}")
        host_project_id   = try(s.vpc_connector_details.host_project_id, var.project_id)
        region            = try(s.vpc_connector_details.region, var.location_id)
        network           = try(s.vpc_connector_details.network, null)
        ip_cidr_range     = try(s.vpc_connector_details.ip_cidr_range, "10.8.0.0/28")
        subnet_name       = try(s.vpc_connector_details.subnet_name, null)
        subnet_project_id = try(s.vpc_connector_details.subnet_project_id, null)
        machine_type      = try(s.vpc_connector_details.machine_type, "e2-micro")
        min_instances     = try(s.vpc_connector_details.min_instances, 2)
        max_instances     = try(s.vpc_connector_details.max_instances, 10)
        min_throughput    = try(s.vpc_connector_details.min_throughput, 200)
        max_throughput    = try(s.vpc_connector_details.max_throughput, 300)
        egress_setting    = try(s.vpc_connector_details.egress_setting, null)

      }
    ] if s.vpc_connector_details != null && s.create_vpc_connector == true
    ])[0], {
    name            = "appengine-connector-${var.location_id}" # Provide a default name
    host_project_id = var.project_id                           # Provide a default project
    region          = var.location_id                          # Provide a default region
  })


  connector_name    = try(local.first_connector_details.name, "appengine-connector-${var.location_id}")
  connector_project = try(local.first_connector_details.host_project_id, var.project_id)
  connector_region  = try(local.first_connector_details.region, var.location_id)
}