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
  description = "The GCP Organization ID where the Firewall Endpoint will be created."
  type        = string
  default     = null
}

variable "billing_project_id" {
  description = "The Project ID to be billed for the Firewall Endpoint's usage."
  type        = string
  default     = null
}

variable "location" {
  description = "The location (zone) for the Firewall Endpoint and its Association. E.g., 'us-central1-a'."
  type        = string
}

# --- Firewall Endpoint Variables ---

variable "create_firewall_endpoint" {
  description = "Set to true to create a Firewall Endpoint."
  type        = bool
  default     = false
}

variable "firewall_endpoint_name" {
  description = "The name of the Firewall Endpoint."
  type        = string
  default     = null
}

variable "firewall_endpoint_labels" {
  description = "A map of labels to add to the Firewall Endpoint."
  type        = map(string)
  default     = {}
}

# --- Firewall Endpoint Association Variables ---

variable "create_firewall_endpoint_association" {
  description = "Set to true to create a Firewall Endpoint Association."
  type        = bool
  default     = false
}

variable "association_name" {
  description = "The name of the Firewall Endpoint Association."
  type        = string
  default     = null
}

variable "association_project_id" {
  description = "The Project ID where the association will be created and where the VPC network resides."
  type        = string
  default     = null
}

variable "vpc_id" {
  description = "The name of the VPC network to associate with the Firewall Endpoint."
  type        = string
  default     = null
}

variable "tls_inspection_policy_id" {
  description = "The name (not the full path) of an optional TLS Inspection Policy to attach to the association."
  type        = string
  default     = null
}

variable "association_labels" {
  description = "A map of labels to add to the Firewall Endpoint Association."
  type        = map(string)
  default     = {}
}

variable "association_disabled" {
  description = "If true, the association is created but will not intercept traffic."
  type        = bool
  default     = false
}

variable "existing_firewall_endpoint_id" {
  description = "The full resource ID of an existing Firewall Endpoint to use for an association if `create_firewall_endpoint` is false."
  type        = string
  default     = null
}