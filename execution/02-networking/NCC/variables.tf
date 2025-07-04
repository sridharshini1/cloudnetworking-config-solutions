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

variable "config_folder_path" {
  description = "Location of YAML files holding NCC configuration values."
  type        = string
  default     = "../../../configuration/networking/ncc/config"
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

variable "export_psc" {
  type        = bool
  description = "Whether Private Service Connect transitivity is enabled for the hub."
  default     = false
}

variable "spoke_labels" {
  description = "Labels to be attached to network connectivity center spoke resource."
  type        = map(string)
  default     = null
}

variable "ncc_hub_description" {
  description = "This can be used to provide additional context or details about the purpose or usage of the hub."
  type        = string
  default     = "Network Connectivity Center hub for managing and connecting multiple network resources."
}

variable "ncc_hub_labels" {
  description = "Labels to be attached to network connectivity center hub resource."
  type        = map(string)
  default     = null
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
variable "group_name" {
  description = "Name of the network connectivity group."
  type        = string
  default     = "default"
}

variable "group_decription" {
  description = "Description for the network connectivity group."
  type        = string
  default     = "Used for auto-accepting projects"
}