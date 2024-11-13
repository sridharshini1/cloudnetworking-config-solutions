/**
 * Copyright 2024 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

output "cluster_details" {
  description = "Display cluster name and details like cluster id, network configuration and state of the AlloyDB cluster created."
  value = {
    for name, cluster in module.alloy_db : name => {
      "cluster_id" : cluster.cluster.cluster_id,
      "cluster_display_name" : cluster.cluster.display_name,
      "database_version" : cluster.cluster.database_version,
      "network_config" : {
        "network" : try(cluster.cluster.network_config[0].network, null),
        "allocated_ip_range" : try(cluster.cluster.network_config[0].allocated_ip_range, null),
        "psc_config" : {
          "psc_enabled" : local.alloydb_network_config[name].psc_config != null ? true : false,
          "configured_allowed_consumer_projects" : local.alloydb_network_config[name].psc_config != null ? local.alloydb_network_config[name].psc_config.psc_allowed_consumer_projects : []
        }
      },
      "cluster_status" : cluster.cluster.state,
      "connectivity_options" : local.alloydb_network_config[name].network_self_link != null ? "PSA" : (local.alloydb_network_config[name].psc_config != null ? "PSC" : "UNKNOWN")
    }
  }
}