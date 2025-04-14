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

resource "google_vpc_access_connector" "connector" {
  count    = local.create_connector ? 1 : 0 # Create at most ONE connector.
  provider = google-beta
  name     = local.connector_name # Use a consistent name.
  project  = local.connector_project
  region   = local.connector_region
  # Use a default ip_cidr_range if not provided in YAML.
  ip_cidr_range = try(local.first_connector_details.ip_cidr_range, "10.8.0.0/28")

  dynamic "subnet" {
    for_each = try(local.first_connector_details.subnet_name, null) != null ? [1] : []
    content {
      name       = local.first_connector_details.subnet_name
      project_id = try(local.first_connector_details.subnet_project_id, var.project_id)
    }
  }

  network = try(local.first_connector_details.network, null)

  machine_type   = try(local.first_connector_details.machine_type, "e2-micro")
  min_instances  = try(local.first_connector_details.min_instances, 2)
  max_instances  = try(local.first_connector_details.max_instances, 10)
  min_throughput = try(local.first_connector_details.min_throughput, 200)
  max_throughput = try(local.first_connector_details.max_throughput, 300)

  lifecycle {
    ignore_changes = [
      min_throughput, max_throughput
    ]
  }
}