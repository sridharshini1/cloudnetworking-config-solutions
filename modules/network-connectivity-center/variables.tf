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

variable "project_id" {
  description = "Project ID for NCC Hub resources."
  type        = string
}

variable "create_new_hub" {
  type        = bool
  description = "Indicates if a new hub should be created."
  default     = false
}

variable "existing_hub_uri" {
  type        = string
  description = "URI of an existing NCC hub to use, if null a new one is created."
  default     = null
}

variable "ncc_hub_name" {
  description = "The Name of the NCC Hub."
  type        = string
}

variable "ncc_hub_description" {
  description = "This can be used to provide additional context or details about the purpose or usage of the hub."
  type        = string
  default     = "Network Connectivity Center hub for managing and connecting multiple network resources."
}

variable "group_decription" {
  description = "Description for the network connectivity group."
  type        = string
  default     = "Used for auto-accepting projects"
}

variable "ncc_hub_labels" {
  description = "Labels to be attached to network connectivity center hub resource."
  type        = map(string)
  default = {
    environment = "prod"
    owner       = "network-team"
  }
}

variable "spoke_labels" {
  description = "Default labels to be merged with spoke-specific labels."
  type        = map(string)
  default = {
    environment = "prod"
    owner       = "network-team"
  }
}

variable "export_psc" {
  description = "Whether Private Service Connect transitivity is enabled for the hub."
  type        = bool
  default     = false
}

variable "group_name" {
  description = "Name of the network connectivity group."
  type        = string
  default     = "default"
}

variable "policy_mode" {
  description = "Policy mode for the NCC hub."
  type        = string
  default     = "PRESET"
}

variable "preset_topology" {
  description = "Preset topology for the NCC hub."
  type        = string
  default     = "MESH"
}

variable "auto_accept_projects" {
  description = "List of projects to auto-accept."
  type        = list(string)
  default     = []
}

variable "vpc_spokes" {
  description = "A map of VPC spokes to be created. The key should be the spoke name."
  type = map(object({
    project_id            = string
    uri                   = string
    description           = optional(string)
    labels                = optional(map(string))
    exclude_export_ranges = optional(list(string))
    include_export_ranges = optional(list(string))
  }))
  default = {}
}

variable "producer_vpc_spokes" {
  description = "A map of Producer VPC spokes to be created."
  type = map(object({
    project_id            = string
    location              = optional(string, "global")
    description           = optional(string)
    uri                   = string # In this context, it's the network URI
    peering               = string
    labels                = optional(map(string))
    exclude_export_ranges = optional(list(string))
    include_export_ranges = optional(list(string))
  }))
  default = {}
}

variable "hybrid_spokes" {
  description = "A map of Hybrid spokes (VPN/Interconnect) to be created."
  type = map(object({
    project_id                 = string
    location                   = optional(string, "global")
    description                = optional(string)
    spoke_type                 = string # "vpn" or "interconnect"
    uris                       = list(string)
    site_to_site_data_transfer = optional(bool, false)
    labels                     = optional(map(string))
  }))
  default = {}
}

variable "router_appliance_spokes" {
  description = "A map of Router Appliance spokes to be created."
  type = map(object({
    location    = string
    description = optional(string)
    instances = list(object({
      virtual_machine = string
      ip_address      = string
    }))
    site_to_site_data_transfer = bool
    labels                     = optional(map(string))
  }))
  default = {}
}