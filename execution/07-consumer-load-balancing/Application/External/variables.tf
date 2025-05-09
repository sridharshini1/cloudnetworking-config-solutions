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

variable "health_check" {
  description = "Health check configuration for the load balancer."
  type = object({
    request_path = string
    port         = number
  })
  default = {
    request_path = "/"
    port         = 80
  }
}

variable "log_config" {
  description = "Log configuration for the load balancer."
  type = object({
    enable      = bool
    sample_rate = number
  })
  default = {
    enable      = true
    sample_rate = 1.0
  }
}

variable "backend_protocol" {
  description = "Protocol used by the backend service."
  type        = string
  default     = "HTTP"
}

variable "backend_port" {
  description = "Port used by the backend service."
  type        = number
  default     = 80
}

variable "backend_port_name" {
  description = "Name of the port used by the backend service."
  type        = string
  default     = "http"
}

variable "backend_timeout_sec" {
  description = "Timeout in seconds for backend requests."
  type        = number
  default     = 10
}

variable "enable_cdn" {
  description = "Enable CDN for the backend service."
  type        = bool
  default     = false
}

variable "iap_config" {
  description = "IAP (Identity-Aware Proxy) configuration for the load balancer."
  type = object({
    enable = bool
  })
  default = {
    enable = false
  }
}

variable "instance_group" {
  type        = string
  default     = null
  description = "Instance group for the Load Balancer"
}


variable "config_folder_path" {
  description = "Location of YAML files holding LB configuration values."
  type        = string
  default     = "../../../../configuration/consumer-load-balancing/Application/External/config/"
}