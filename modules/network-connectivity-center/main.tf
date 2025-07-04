/**
 * Copyright 2024-25 Google LLC
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

locals {
  hub_id = (
    var.existing_hub_uri != null && var.existing_hub_uri != ""
    ? var.existing_hub_uri
    : google_network_connectivity_hub.hub[0].id
  )
}

resource "google_network_connectivity_hub" "hub" {
  count           = var.create_new_hub ? 1 : 0
  name            = var.ncc_hub_name
  project         = var.project_id
  description     = var.ncc_hub_description
  labels          = var.ncc_hub_labels
  export_psc      = var.export_psc
  policy_mode     = var.policy_mode
  preset_topology = var.preset_topology
}

resource "google_network_connectivity_group" "default" {
  name        = var.group_name
  hub         = local.hub_id
  project     = var.project_id
  description = var.group_decription
  labels      = var.ncc_hub_labels

  auto_accept {
    auto_accept_projects = var.auto_accept_projects
  }
}

resource "google_network_connectivity_spoke" "vpc_spoke" {
  for_each    = var.vpc_spokes
  project     = each.value.project_id
  name        = each.key
  location    = "global"
  description = each.value.description
  hub         = local.hub_id
  labels      = merge(var.spoke_labels, lookup(each.value, "labels", {}))

  linked_vpc_network {
    uri                   = each.value.uri
    exclude_export_ranges = lookup(each.value, "exclude_export_ranges", [])
    include_export_ranges = lookup(each.value, "include_export_ranges", [])
  }
}

resource "google_network_connectivity_spoke" "producer_vpc_spoke" {
  for_each    = var.producer_vpc_spokes
  project     = each.value.project_id
  name        = each.key
  location    = each.value.location
  description = each.value.description
  hub         = local.hub_id
  labels      = merge(var.spoke_labels, lookup(each.value, "labels", {}))

  linked_producer_vpc_network {
    network               = each.value.uri
    peering               = each.value.peering
    exclude_export_ranges = lookup(each.value, "exclude_export_ranges", [])
    include_export_ranges = lookup(each.value, "include_export_ranges", [])
  }

  depends_on = [
    google_network_connectivity_spoke.vpc_spoke
  ]
}

resource "google_network_connectivity_spoke" "hybrid_spoke" {
  for_each    = merge(var.linked_vpn_tunnels, var.linked_interconnect_attachments)
  project     = each.value.project_id
  name        = each.key
  location    = each.value.location
  description = each.value.description
  hub         = local.hub_id
  labels      = merge(var.spoke_labels, lookup(each.value, "labels", {}))

  dynamic "linked_interconnect_attachments" {
    for_each = each.value.type == "linked_interconnect_attachments" ? [1] : []
    content {
      uris                       = each.value.uris
      site_to_site_data_transfer = each.value.site_to_site_data_transfer
    }
  }

  dynamic "linked_vpn_tunnels" {
    for_each = each.value.type == "linked_vpn_tunnels" ? [1] : []
    content {
      uris                       = each.value.uris
      site_to_site_data_transfer = each.value.site_to_site_data_transfer
    }
  }

  depends_on = [
    google_network_connectivity_spoke.vpc_spoke
  ]
}

resource "google_network_connectivity_spoke" "router_appliance_spoke" {
  for_each    = var.router_appliance_spokes
  project     = var.project_id
  name        = each.key
  location    = each.value.location
  description = each.value.description
  hub         = local.hub_id
  labels      = merge(var.spoke_labels, lookup(each.value, "labels", {}))

  linked_router_appliance_instances {
    dynamic "instances" {
      for_each = each.value.instances
      iterator = instance_list
      content {
        virtual_machine = instance_list.value.virtual_machine
        ip_address      = instance_list.value.ip_address
      }
    }
    site_to_site_data_transfer = each.value.site_to_site_data_transfer
  }
  depends_on = [
    google_network_connectivity_spoke.hybrid_spoke
  ]
}
