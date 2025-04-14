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
  type        = string
  description = "The ID of the project in which to deploy the AppEngine application."
}

variable "location_id" {
  type        = string
  description = "The location to deploy the AppEngine application."
  default     = "us-central1" #Goodpracticetohaveadefault.
}

variable "auth_domain" {
  type        = string
  default     = null
  description = "The domain to authenticate users with."
}

variable "database_type" {
  type        = string
  default     = null
  description = "The type of database to use."
}

variable "serving_status" {
  type        = string
  default     = null
  description = "The serving status of the application."
}

variable "feature_settings" {
  type = object({
    split_health_checks = optional(bool, true)
  })
  default     = null
  description = "Feature settings for the AppEngine application."
}

variable "iap" {
  type = object({
    enabled              = optional(bool, false) #Makeenabledoptional
    oauth2_client_id     = optional(string)      #Makeoptional
    oauth2_client_secret = optional(string)      #Makeoptional
  })
  default     = null #Correctdefault
  description = "Identity-AwareProxysettings."
}
variable "dispatch_rules" {
  type = list(object({
    domain  = string
    path    = string
    service = string
  }))
  default     = []
  description = "dispatch rules for the appengine"
}

variable "domain_mappings" {
  type = list(object({
    domain_name       = string
    override_strategy = optional(string, "STRICT")
    ssl_settings = optional(object({
      certificate_id      = optional(string)
      ssl_management_type = string
    }))
  }))
  default     = []
  description = "Domain mappings for the AppEngine application."
}

variable "firewall_rules" {
  type = list(object({
    source_range = string
    action       = string
    priority     = optional(number)
    description  = optional(string)
  }))
  default     = []
  description = "Firewall rules for the  AppEngine application."
}

variable "services" {
  type = map(object({
    service             = string
    runtime             = string
    version_id          = optional(string)
    app_engine_apis     = optional(bool, false)
    runtime_api_version = optional(string)
    service_account     = optional(string)
    threadsafe          = optional(bool)
    inbound_services    = optional(list(string))
    instance_class      = optional(string)
    labels              = optional(map(string))
    automatic_scaling = optional(object({
      max_concurrent_requests = optional(number)
      max_idle_instances      = optional(number)
      max_pending_latency     = optional(string)
      min_idle_instances      = optional(number)
      min_pending_latency     = optional(string)
      standard_scheduler_settings = optional(object({
        target_cpu_utilization        = optional(number)
        target_throughput_utilization = optional(number)
        min_instances                 = optional(number)
        max_instances                 = optional(number)
      }))
    }))
    basic_scaling = optional(object({
      max_instances = optional(number)
      idle_timeout  = optional(string)
    }))
    manual_scaling = optional(object({
      instances = optional(number)
    }))
    network_settings = optional(object({
      ingress_traffic_allowed = optional(string)
    }))
    delete_service_on_destroy = optional(bool, false)
    deployment = optional(object({
      zip = optional(object({
        source_url  = string
        files_count = optional(number)
      }))
      files = optional(object({ // SINGLE OBJECT
        name       = string
        source_url = string
        sha1_sum   = optional(string)
      }))
    }))
    env_variables = optional(map(string))
    entrypoint = optional(object({
      shell = string
    }))
    handlers = optional(list(object({
      auth_fail_action            = optional(string)
      login                       = optional(string)
      redirect_http_response_code = optional(string)
      script                      = optional(object({ script_path = string }))
      security_level              = optional(string)
      url_regex                   = optional(string)
      static_files = optional(object({
        path                  = optional(string)
        upload_path_regex     = optional(string)
        http_headers          = optional(map(string))
        mime_type             = optional(string)
        expiration            = optional(string)
        require_matching_file = optional(bool)
        application_readable  = optional(bool)
      }))
    })))
    libraries = optional(list(object({
      name    = optional(string)
      version = optional(string)
    })))
    vpc_access_connector = optional(object({ #Existingconnectordetails
      name = string
    }))
    vpc_connector_details = optional(object({
      name              = string
      host_project_id   = optional(string) # Project ID for the new connector
      region            = optional(string) # Region for the new connector
      ip_cidr_range     = optional(string)
      subnet_name       = optional(string)
      subnet_project_id = optional(string)
      network           = optional(string)
      machine_type      = optional(string)
      min_instances     = optional(number)
      max_instances     = optional(number)
      min_throughput    = optional(number)
      max_throughput    = optional(number)
      egress_setting    = optional(string)

    }))
    create_vpc_connector = optional(bool, false)
    split_traffic = optional(object({
      migrate_traffic = optional(bool, false)
      allocations     = map(number)
      shard_by        = optional(string)
    }))
    create_split_traffic    = optional(bool, false)
    create_dispatch_rules   = optional(bool, false)
    create_network_settings = optional(bool, false)
    noop_on_destroy         = optional(bool, false)
  }))
  default = {}
}

#---Defaultvalues(canbeoverriddenatthewrapperlevel)---

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
  description = "Service account to be used by the AppEngine version(defaults to the AppEngine default service account)."
  default     = null
}
variable "threadsafe" {
  type        = bool
  description = "Whether the application is thread safe(default:true).Deprecated in newer runtimes."
  default     = null
}
variable "inbound_services" {
  type        = list(string)
  description = "(Optional)A list of inbound services that this service will receive traffic from.An empty list means it will not receive traffic from other services"
  default     = null
}
variable "instance_class" {
  type        = string
  description = "Instance class for the AppEngine version(default:F1)."
  default     = "F2"
}
variable "labels" {
  type        = map(string)
  description = "A map of labels to apply to the service"
  default     = {}
}
variable "automatic_scaling_max_concurrent_requests" {
  type        = number
  description = "Maximum concurrent requests for automatic scaling(default:50)."
  default     = 50
}

variable "automatic_scaling_max_idle_instances" {
  type        = number
  description = "Maximum idle instances for automatic scaling(default:1)."
  default     = 1
}

variable "automatic_scaling_max_pending_latency" {
  type        = string
  description = "Maximum pending latency for automatic scaling(default:30s)."
  default     = "30s"
}

variable "automatic_scaling_min_idle_instances" {
  type        = number
  description = "Minimum idle instances for automatic scaling(default:0)."
  default     = 0
}

variable "automatic_scaling_min_pending_latency" {
  type        = string
  description = "Minimum pending latency for automatic scaling(default:30ms)."
  default     = "1s"
}
variable "automatic_scaling_standard_scheduler_settings_target_cpu_utilization" {
  type        = number
  description = "Target CPU utilization for standard scheduler settings(default:0.6)."
  default     = 0.6
}

variable "automatic_scaling_standard_scheduler_settings_target_throughput_utilization" {
  type        = number
  description = "Target throughput utilization for standard scheduler settings(default:0.6)."
  default     = 0.6
}

variable "automatic_scaling_standard_scheduler_settings_min_instances" {
  type        = number
  description = "Minimum instances for standard scheduler settings(default:0)."
  default     = 0
}

variable "automatic_scaling_standard_scheduler_settings_max_instances" {
  type        = number
  description = "Maximum instances for standard scheduler settings(default:100)."
  default     = 100
}
variable "basic_scaling_max_instances" {
  type        = number
  description = "Maximum instances for basic scaling."
  default     = null
}

variable "basic_scaling_idle_timeout" {
  type        = string
  description = "Idle timeout for basic scaling."
  default     = null
}
variable "manual_scaling_instances" {
  type        = number
  description = "Number of instances for manualscaling."
  default     = null
}

variable "delete_service_on_destroy" {
  type        = bool
  description = "Whether to delete the service when destroying the resource(default:true)."
  default     = false
}
#deployment
variable "deployment_zip_source_url" {
  type        = string
  description = "Source URL for the deployment ZIP file."
  default     = null
}

variable "deployment_zip_files_count" {
  type        = number
  description = "Number of files in the deployment ZIP file."
  default     = null
}
variable "deployment_files_name" {
  type        = string
  description = "deployment filesname"
  default     = null
}
variable "deployment_files_sha1_sum" {
  type        = string
  description = "deployment files sha1_sum"
  default     = null
}
variable "deployment_files_source_url" {
  type        = string
  description = "deployment files source_url"
  default     = null
}

variable "env_variables" {
  type        = map(string)
  description = "Environment variables for the AppEngine version(default:{})."
  default     = null
}
#entrypoint
variable "entrypoint_shell" {
  type        = string
  description = "entry point shell"
  default     = null
}

#handlers

variable "handlers_auth_fail_action" {
  type        = string
  description = "handlers auth fail action"
  default     = null
}
variable "handlers_login" {
  type        = string
  description = "handlers login"
  default     = null
}
variable "handlers_redirect_http_response_code" {
  type        = string
  description = "handlers redirect http responsecode"
  default     = null
}
#handlersscript
variable "handlers_script_script_path" {
  type        = string
  description = "script path"
  default     = "auto"
}
variable "handlers_security_level" {
  type        = string
  description = "handlers security level"
  default     = null
}
variable "handlers_url_regex" {
  type        = string
  description = "handlers url regex"
  default     = "/.*"
}
#handlers_static_files
variable "handlers_static_files_path" {
  type        = string
  description = "handlers static files path"
  default     = null
}
variable "handlers_static_files_upload_path_regex" {
  type        = string
  description = "handlers static files upload path regex"
  default     = null
}
variable "handlers_static_files_http_headers" {
  type        = map(string)
  description = "handlers static files http headers"
  default     = null
}
variable "handlers_static_files_mime_type" {
  type        = string
  description = "handlers static files mime type"
  default     = null
}
variable "handlers_static_files_expiration" {
  type        = string
  description = "handlers static files expiration"
  default     = null
}
variable "handlers_static_files_require_matching_file" {
  type        = bool
  description = "handlers static files require matching file"
  default     = null
}
variable "handlers_static_files_application_readable" {
  type        = bool
  description = "handlers static files application readable"
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

variable "vpc_access_connector_egress_setting" {
  type        = string
  description = "vpc connector egress setting"
  default     = null
}
variable "noop_on_destroy" {
  type        = bool
  description = "If set to true,then the application version will not be deleted"
  default     = null
}

variable "connector_name" {
  type        = string
  default     = ""
  description = "VPC connector name"
}

variable "create_app_engine_application" {
  type        = bool
  default     = false
  description = "Whether to create the google_app_engine_application resource."
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
  default     = true # Default doesn't matter as it wasn't used
  description = "Whether to create appengine standard app version"
}
