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

variable "organization_id" {
  description = "Organization id to be used to deploy the resources."
  type        = string
  default     = null
}

variable "create_firewall_endpoint" {
  description = "Control condition to create firewall endpoint."
  type        = bool
  default     = false
}

variable "firewall_endpoint_name" {
  description = "Firewall endpoint name"
  type        = string
  default     = null
}

variable "location" {
  description = "Location (zone) for the endpoint to be deployed."
  type        = string
  default     = null
}

variable "billing_project_id" {
  description = "Project id to be billed for the resources not deployed in any specific project"
  type        = string
  default     = null
}

variable "firewall_endpoint_labels" {
  description = "Labels for a firewall endpoint."
  type        = map(string)
  default     = {}
}

variable "create_firewall_endpoint_association" {
  description = "Control variable for creating an association for firewall endpoint."
  type        = bool
  default     = false
}

variable "association_name" {
  description = "Name for an association for firewall endpoint."
  type        = string
  default     = null
}

variable "association_project_id" {
  description = "Project ID for the firewall endpoint association."
  type        = string
  default     = null
}
variable "vpc_id" {
  description = "VPC network name for the firewall endpoint association."
  type        = string
  default     = null
}
variable "tls_inspection_policy_id" {
  description = "TLS Inspection Policy name."
  type        = string
  default     = null
}

variable "association_labels" {
  description = "Labels for an association."
  type        = map(string)
  default     = {}
}

variable "association_disabled" {
  description = "Disabled State for an association."
  type        = bool
  default     = false
}

variable "existing_firewall_endpoint_id" {
  description = "Existing firewall endpoint ID to be linked to the association."
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Path to the folder containing the YAML configuration files for firewall endpoints."
  type        = string
  default     = "../../../configuration/networking/FirewallEndpoint/config/"
}