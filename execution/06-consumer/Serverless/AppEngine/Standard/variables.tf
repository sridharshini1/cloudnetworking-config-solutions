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

variable "create_app_engine_application" {
  type        = bool
  default     = false
  description = "Whether to create the google_app_engine_application resource."
}

variable "location_id" {
  type        = string
  default     = ""
  description = "Location / region where nthe app engine applcaition is to be deployed"
}

variable "create_dispatch_rules" {
  type        = bool
  default     = false
  description = "Whether to create AppEngine dispatch rules."
}

variable "create_domain_mappings" {
  type        = bool
  default     = false
  description = "Whether to create AppEngine domain mappings."
}

variable "create_firewall_rules" {
  type        = bool
  default     = false
  description = "Whether to create AppEngine firewall rules."
}

variable "create_network_settings" {
  type        = bool
  default     = false
  description = "Whether to create AppEngine service network settings."
}

variable "create_split_traffic" {
  type        = bool
  default     = false
  description = "Whether to create AppEngine service split traffic settings."
}

variable "create_app_version" {
  type        = bool
  default     = true
  description = "Whether to create appengine standard app version"
}

variable "create_vpc_connector" {
  type        = bool
  default     = false
  description = "Whether to create a vpc access connector."
}
variable "app_engine_apis" {
  type        = bool
  description = "Enable AppEngine APIs for new services"
  default     = null
}
variable "runtime_api_version" {
  type        = string
  description = "AppEngine runtime API version"
  default     = null
}
variable "service_account" {
  type        = string
  description = "Service account to be used by the AppEngine version(defaultstotheAppEnginedefaultserviceaccount)."
  default     = null
}

variable "threadsafe" {
  type        = bool
  description = "Whether the application is threadsafe(default:true).Deprecatedinnewerruntimes."
  default     = null
}
variable "inbound_services" {
  type        = list(string)
  description = "(Optional)A list of inbound services that this service will receive traffic from. An empty list means it will not receive traffic from other services"
  default     = null
}

variable "instance_class" {
  type        = string
  description = "Instance class for the AppEngine version(default:F1)."
  default     = "F2"
}

variable "delete_service_on_destroy" {
  type        = bool
  description = "Whether to delete the service when destroying the resource(default:true)."
  default     = true
}

variable "env_variables" {
  type        = map(string)
  description = "Environment variables for the AppEngine version(default:{})."
  default     = null
}
#entrypoint
variable "entrypoint_shell" {
  type        = string
  description = "entrypointshell"
  default     = null
}

#libraries
variable "libraries_name" {
  type        = string
  description = "libraries name"
  default     = null
}
variable "libraries_version" {
  type        = string
  description = "libraries version"
  default     = null
}

variable "noop_on_destroy" {
  type        = bool
  description = "If set to true,then the application version will not be deleted"
  default     = null
}
variable "labels" {
  type        = map(string)
  description = "Labels to apply to the AppEngine version"
  default     = {}
}

variable "handlers" {
  description = "Default value for 'handlers' if missing in YAML."
  type        = list(any) # Use list(any) for flexibility, or define exact object type
  default     = []
}

variable "libraries" {
  description = "Default value for 'libraries' if missing in YAML."
  type        = list(any) # Use list(any) for flexibility
  default     = []
}

variable "automatic_scaling" {
  description = "Default value for 'automatic_scaling' block if missing in YAML."
  type        = any # Use 'any' or define the exact object type if needed
  default     = null
}

variable "basic_scaling" {
  description = "Default value for 'basic_scaling' block if missing in YAML."
  type        = any
  default     = null
}

variable "manual_scaling" {
  description = "Default value for 'manual_scaling' block if missing in YAML."
  type        = any
  default     = null
}

variable "vpc_access_connector" {
  description = "Default value for 'vpc_access_connector' block if missing in YAML."
  type        = any
  default     = null
}

variable "vpc_connector_details" {
  description = "Default value for 'vpc_connector_details' block if missing in YAML."
  type        = any
  default     = null
}

variable "network_settings" {
  description = "Default value for 'network_settings' block if missing in YAML."
  type        = any
  default     = null
}

variable "split_traffic" {
  description = "Default value for 'split_traffic' block if missing in YAML."
  type        = any
  default     = null
}

variable "app_engine_application" {
  description = "Default value for 'app_engine_application' block if missing in YAML."
  type        = map(any) # Use map(any) or define exact object type
  default     = {}
}

variable "dispatch_rules" {
  description = "Default value for 'dispatch_rules' list if missing in YAML."
  type        = list(any)
  default     = []
}

variable "domain_mappings" {
  description = "Default value for 'domain_mappings' list if missing in YAML."
  type        = list(any)
  default     = []
}

variable "firewall_rules" {
  description = "Default value for 'firewall_rules' list if missing in YAML."
  type        = list(any)
  default     = []
}

variable "version_id" {
  type        = string
  description = "Default value for the version_id of missing in the YAML"
  default     = "v1"
}

variable "deployment" {
  description = "Default value for 'deployment' block if missing in YAML."
  type        = any # Use 'any' type as the structure can vary and default is null
  default     = null
}

variable "entrypoint" {
  description = "Default value for 'entrypoint' block if missing in YAML."
  type        = any # Use 'any' type as the structure can vary and default is null
  default     = null
}

variable "auth_domain" {
  description = "Default App Engine auth_domain if app_engine_application.auth_domain is missing in YAML."
  type        = string
  default     = null
}

variable "database_type" {
  description = "Default App Engine database_type if app_engine_application.database_type is missing in YAML."
  type        = string
  default     = null # Or e.g., "CLOUD_DATASTORE_COMPATIBILITY"
}

variable "serving_status" {
  description = "Default App Engine serving_status if app_engine_application.serving_status is missing in YAML."
  type        = string
  default     = null # Or e.g., "SERVING"
}

variable "feature_settings" {
  description = "Default App Engine feature_settings block if app_engine_application.feature_settings is missing in YAML."
  type        = any # Use 'any' or define the specific object type
  default     = null
}

variable "iap" {
  description = "Default App Engine iap block if app_engine_application.iap is missing in YAML."
  type        = any # Use 'any' or define the specific object type
  default     = null
}

variable "config_folder_path" {
  description = "Path to the folder containing AppEngine YAML configuration files"
  type        = string
  default     = "../../../../../configuration/consumer/Serverless/AppEngine/Standard/config"
}
