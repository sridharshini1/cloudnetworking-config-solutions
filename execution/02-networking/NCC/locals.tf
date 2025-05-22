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
  raw_instance_files = fileset(local.config_folder_path, "[^_]*.yaml")

  raw_instance_list = [
    for file in local.raw_instance_files :
    yamldecode(file("${local.config_folder_path}/${file}"))
  ]

  # Flatten the list of hubs and spokes from all the parsed YAML files
  all_hubs = flatten([for instance in local.raw_instance_list : try(instance.hubs, [])])
  all_spokes = flatten([
    for instance in local.raw_instance_list :
    [
      for spoke in try(instance.spokes, []) :
      merge(spoke, { hub_name = try(instance.hubs[0].name, null) })
    ]
  ])

  # Map hubs by name
  instance_list = {
    for hub in local.all_hubs :
    hub.name => {
      project_id           = hub.project_id
      create_new_hub       = try(hub.create_new_hub, var.create_new_hub)
      existing_hub_uri     = try(hub.existing_hub_uri, var.existing_hub_uri)
      spoke_labels         = try(hub.spoke_labels, var.spoke_labels)
      export_psc           = try(hub.export_psc, var.export_psc)
      ncc_hub_name         = hub.name
      ncc_hub_description  = try(hub.description, var.ncc_hub_description)
      ncc_hub_labels       = try(hub.labels, var.ncc_hub_labels)
      policy_mode          = try(hub.policy_mode, var.policy_mode)
      preset_topology      = try(hub.preset_topology, var.preset_topology)
      auto_accept_projects = try(hub.auto_accept_projects, var.auto_accept_projects)
      group_name           = try(hub.group_name, var.group_name)
      group_decription     = try(hub.group_decription, var.group_decription)

      vpc_spokes = {
        for spoke in local.all_spokes :
        spoke.name => spoke
        if spoke.type == "linked_vpc_network" && spoke.hub_name == hub.name
      }
      producer_vpc_spokes = {
        for spoke in local.all_spokes :
        spoke.name => spoke
        if spoke.type == "linked_producer_vpc_network" && spoke.hub_name == hub.name
      }
      hybrid_spokes = {
        for spoke in local.all_spokes :
        spoke.name => spoke
        if spoke.type == "hybrid_spoke" && spoke.hub_name == hub.name
      }
      router_appliance_spokes = {
        for spoke in local.all_spokes :
        spoke.name => spoke
        if spoke.type == "router_appliance_spoke" && spoke.hub_name == hub.name
      }
    }
  }
}