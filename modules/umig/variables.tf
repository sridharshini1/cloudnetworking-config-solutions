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

variable "project_id" {
  description = "The GCP project ID."
  type        = string
}

variable "zone" {
  description = "The zone for the instance group."
  type        = string
}

variable "network" {
  description = "The name or self-link of the network to filter instances."
  type        = string
}

variable "name" {
  description = "The name of the instance group."
  type        = string
}

variable "description" {
  description = "Description for the instance group."
  type        = string
  default     = "Instance group managed by the UMIG Terraform module."
}

variable "instances" {
  description = "List of instance names."
  type        = list(string)
}

variable "named_ports" {
  description = "Map of named ports."
  type        = map(number)
  default     = {}
}