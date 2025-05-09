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
  description = "Location of YAML files holding VPC Access Connector configuration values."
  type        = string
  default     = "../../../../configuration/consumer/Serverless/VPCAccessConnector/config"
}

variable "network" {
  description = "Default VPC network name if not specified in YAML."
  type        = string
  default     = null
}

variable "subnet_name" {
  description = "Default subnet name if not specified in YAML."
  type        = string
  default     = null
}

variable "ip_cidr_range" {
  description = "Default IP CIDR range if not specified in YAML."
  type        = string
  default     = null
}

variable "host_project_id" {
  description = "Default host project ID for the subnet if not specified in YAML (for Shared VPC)."
  type        = string
  default     = null
}

variable "machine_type" {
  description = "Default machine type for the connector instances."
  type        = string
  default     = "e2-standard-4"
}

variable "min_instances" {
  description = "Default minimum number of instances (Note: not used by vpc-serverless-connector-beta module)."
  type        = number
  default     = null
}

variable "max_instances" {
  description = "Default maximum number of instances (Note: not used by vpc-serverless-connector-beta module)."
  type        = number
  default     = null
}

variable "max_throughput" {
  description = "Default maximum throughput in Mbps."
  type        = number
  default     = null
}

variable "min_throughput" {
  description = "Default minimum throughput in Mbps."
  type        = number
  default     = null
}

