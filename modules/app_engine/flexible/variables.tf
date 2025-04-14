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
  description = "The ID of the project in which to deploy the App Engine application."
}
variable "location_id" {
  type        = string
  description = "The location to deploy the App Engine application."
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
  description = "Feature settings for the App Engine application."
}
variable "iap" {
  type = object({
    enabled              = bool
    oauth2_client_id     = string
    oauth2_client_secret = string
  })
  default     = null
  description = "Identity-Aware Proxy settings."
}
variable "dispatch_rules" {
  type = list(object({
    domain  = string
    path    = string
    service = string
  }))
  default     = []
  description = "dispatch rules for the app engine"
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
  description = "Domain mappings for the App Engine application."
}
variable "firewall_rules" {
  type = list(object({
    source_range = string
    action       = string
    priority     = optional(number)
    description  = optional(string)
  }))
  default     = []
  description = "Firewall rules for the App Engine application."
}
variable "services" {
  type = map(object({
    service                      = string
    runtime                      = string
    version_id                   = string
    instance_class               = optional(string)
    runtime_api_version          = optional(string)
    runtime_channel              = optional(string)
    runtime_main_executable_path = optional(string)
    service_account              = optional(string)
    serving_status               = optional(string)
    nobuild_files_regex          = optional(string)
    delete_service_on_destroy    = optional(bool, false)
    noop_on_destroy              = optional(bool, false)
    beta_settings                = optional(map(string))
    inbound_services             = optional(list(string))
    labels                       = optional(map(string))
    entrypoint = optional(object({
      shell = string
    }))
    liveness_check = optional(object({
      path              = string
      host              = optional(string)
      failure_threshold = optional(number)
      success_threshold = optional(number)
      check_interval    = optional(string)
      timeout           = optional(string)
      initial_delay     = optional(string)
    }))
    readiness_check = optional(object({
      path              = string
      host              = optional(string)
      failure_threshold = optional(number)
      success_threshold = optional(number)
      check_interval    = optional(string)
      timeout           = optional(string)
      app_start_timeout = optional(string)
    }))
    network = optional(object({
      name             = string
      subnetwork       = optional(string)
      forwarded_ports  = optional(list(string))
      instance_tag     = optional(string)
      session_affinity = optional(bool)
      instance_ip_mode = optional(string)
    }))
    resources = optional(object({
      cpu       = optional(number)
      disk_gb   = optional(number)
      memory_gb = optional(number)
      volumes = optional(list(object({
        name        = string
        volume_type = string
        size_gb     = number
      })))
    }))
    flexible_runtime_settings = optional(object({
      operating_system = optional(string)
      runtime_version  = optional(string)
    }))
    automatic_scaling = optional(object({
      cool_down_period        = optional(string)
      max_concurrent_requests = optional(number)
      max_total_instances     = optional(number)
      min_total_instances     = optional(number)
      max_idle_instances      = optional(number)
      min_idle_instances      = optional(number)
      max_pending_latency     = optional(string)
      min_pending_latency     = optional(string)
      cpu_utilization = optional(object({
        target_utilization        = number
        aggregation_window_length = optional(string)
      }))
      disk_utilization = optional(object({
        target_read_bytes_per_second  = optional(number)
        target_read_ops_per_second    = optional(number)
        target_write_bytes_per_second = optional(number)
        target_write_ops_per_second   = optional(number)
      }))
      network_utilization = optional(object({
        target_received_bytes_per_second   = optional(number)
        target_received_packets_per_second = optional(number)
        target_sent_bytes_per_second       = optional(number)
        target_sent_packets_per_second     = optional(number)
      }))
      request_utilization = optional(object({
        target_concurrent_requests      = optional(number)
        target_request_count_per_second = optional(number)
      }))
    }))
    manual_scaling = optional(object({
      instances = number
    }))
    endpoints_api_service = optional(object({
      name                   = string
      config_id              = optional(string)
      rollout_strategy       = optional(string)
      disable_trace_sampling = optional(bool)
    }))
    deployment = optional(object({
      container = optional(object({
        image = string
      }))
      files = optional(object({
        name       = string
        source_url = string
        sha1_sum   = optional(string)
      }))
      zip = optional(object({
        source_url  = string
        files_count = optional(number)
      }))
      cloud_build_options = optional(object({
        app_yaml_path       = string
        cloud_build_timeout = optional(string)
      }))
    }))
    env_variables = optional(map(string), {})
    network_settings = optional(object({
      ingress_traffic_allowed = optional(string, "INGRESS_TRAFFIC_ALLOWED_ALL")
    }))
    split_traffic = optional(object({
      shard_by        = optional(string, "IP")
      allocations     = map(number)
      migrate_traffic = optional(bool, false)
    }))
  }))
  description = "A map of service configurations."
}
variable "create_application" {
  type        = bool
  default     = true
  description = "Whether to create the App Engine application. Defaults to true."
}
variable "create_dispatch_rules" {
  type        = bool
  default     = true
  description = "Whether to create the dispatch rules. Defaults to true."
}
variable "create_domain_mappings" {
  type        = bool
  default     = true
  description = "Whether to create the domain mappings. Defaults to true."
}
variable "create_firewall_rules" {
  type        = bool
  default     = true
  description = "Whether to create the firewall rules. Defaults to true."
}
variable "create_network_settings" {
  type        = bool
  default     = true
  description = "Whether to create network settings. Defaults to true."
}
variable "create_split_traffic" {
  type        = bool
  default     = true
  description = "Whether to create split traffic settings. Defaults to true."
}