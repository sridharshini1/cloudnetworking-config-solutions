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

variable "version_id" {
  type        = string
  default     = "v1"
  description = "Default version id for the app engine application"
}
variable "domain_mappings" {
  type = list(object({
    domain_name = string
    ssl_settings = optional(object({
      certificate_id      = optional(string)
      ssl_management_type = string
    }))
  }))
  default     = []
  description = "Domain mappings for the App Engine application."
}
variable "location_id" {
  type        = string
  default     = ""
  description = "Default location for the App Engine application."
}
variable "create_application" {
  type        = bool
  default     = false
  description = "Default value for whether to create the App Engine application."
}
variable "create_dispatch_rules" {
  type        = bool
  default     = false
  description = "Default value for whether to create dispatch rules."
}
variable "create_domain_mappings" {
  type        = bool
  default     = false
  description = "Default value for whether to create domain mappings."
}
variable "create_firewall_rules" {
  type        = bool
  default     = false
  description = "Default value for whether to create firewall rules."
}
variable "create_network_settings" {
  type        = bool
  default     = false
  description = "Default value for whether to create network settings."
}
variable "create_split_traffic" {
  type        = bool
  default     = false
  description = "Default value for whether to create split traffic settings."
}
variable "auth_domain" {
  type        = string
  default     = null
  description = "Default authentication domain."
}
variable "database_type" {
  type        = string
  default     = null
  description = "Default database type."
}
variable "serving_status" {
  type        = string
  default     = null
  description = "Default serving status."
}
variable "feature_settings" {
  type = object({
    split_health_checks = optional(bool, true)
  })
  default     = null
  description = "Default feature settings (if none in YAML)."
}
variable "iap_settings" {
  type = object({
    enabled              = bool
    oauth2_client_id     = string
    oauth2_client_secret = string
  })
  default     = null
  description = "Default IAP settings (if none in YAML)."
}
variable "dispatch_rules" {
  type = list(object({
    domain  = string
    path    = string
    service = string
  }))
  default     = []
  description = "Default dispatch rules (if none in YAML)."
}
variable "firewall_rules" {
  type = list(object({
    source_range = string
    action       = string
    priority     = optional(number)
    description  = optional(string)
  }))
  default     = []
  description = "Default firewall rules (if none in YAML)."
}
variable "instance_class" {
  description = "Default instance class if not specified in service YAML."
  type        = string
  default     = null
}
variable "flexible_runtime_settings" {
  description = "Default flexible_runtime_settings block if not specified in service YAML."
  type        = any
  default     = null
}
variable "network" {
  description = "Default network block (for service VM) if not specified in service YAML."
  type        = any
  default     = null
}
variable "resources" {
  description = "Default resources block if not specified in service YAML."
  type        = any
  default     = null
}
variable "entrypoint" {
  description = "Default entrypoint block if not specified in service YAML."
  type        = any
  default     = null
}
variable "automatic_scaling" {
  description = "Default automatic_scaling block if not specified in service YAML."
  type        = any
  default     = null
}
variable "manual_scaling" {
  description = "Default manual_scaling block if not specified in service YAML."
  type        = any
  default     = null
}
variable "env_variables" {
  description = "Default env_variables map if not specified in service YAML."
  type        = map(string)
  default     = {}
}
variable "deployment" {
  description = "Default deployment block if not specified in service YAML."
  type        = any
  default     = null
}
variable "liveness_check" {
  description = "Default liveness_check block if not specified in service YAML."
  type        = any
  default     = "/"
}
variable "readiness_check" {
  description = "Default readiness_check block if not specified in service YAML."
  type        = any
  default     = "/"
}
variable "service_account" {
  description = "Default service_account if not specified in service YAML. If null, App Engine default SA is often used."
  type        = string
  default     = null
}
variable "endpoints_api_service" {
  description = "Default endpoints_api_service block if not specified in service YAML."
  type        = any
  default     = null
}
variable "nobuild_files_regex" {
  description = "Default nobuild_files_regex if not specified in service YAML."
  type        = string
  default     = null
}
variable "beta_settings" {
  description = "Default beta_settings map if not specified in service YAML."
  type        = map(string)
  default     = {}
}
variable "inbound_services" {
  description = "Default inbound_services list if not specified in service YAML."
  type        = list(string)
  default     = null
}
variable "labels" {
  description = "Default labels map if not specified in service YAML."
  type        = map(string)
  default     = {}
}
variable "service_serving_status" {
  description = "Default serving_status for a specific service if not specified in its YAML (distinct from app-level default)."
  type        = string
  default     = null
}
variable "runtime_api_version" {
  description = "Default runtime_api_version if not specified in service YAML."
  type        = string
  default     = "1"
}
variable "runtime_channel" {
  description = "Default runtime_channel if not specified in service YAML."
  type        = string
  default     = null
}
variable "runtime_main_executable_path" {
  description = "Default runtime_main_executable_path if not specified in service YAML."
  type        = string
  default     = null
}
variable "delete_service_on_destroy" {
  description = "Default delete_service_on_destroy flag if not specified in service YAML."
  type        = bool
  default     = false
}
variable "noop_on_destroy" {
  description = "Default noop_on_destroy flag if not specified in service YAML."
  type        = bool
  default     = false
}
variable "network_settings_block" {
  description = "Default network_settings block (for service network settings resource) if not specified in service YAML."
  type        = any
  default     = null
}
variable "split_traffic_block" {
  description = "Default split_traffic block (for service split traffic resource) if not specified in service YAML."
  type        = any
  default     = null
}
variable "config_folder_path" {
  description = "Path to the folder containing App Engine Flexible YAML configuration files"
  type        = string
  default     = "../../../../../configuration/consumer/Serverless/AppEngine/Flexible/config"
}