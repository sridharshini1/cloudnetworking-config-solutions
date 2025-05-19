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

variable "description" {
  description = "Optional description used for resources."
  type        = string
  default     = "Terraform managed External Passthrough Network Load Balancer."
}

variable "labels" {
  description = "Labels set on resources."
  type        = map(string)
  default     = {}
}

# Default backend service configuration values
variable "backend_protocol" {
  description = "Default protocol used by the backend service. Common for NLB: TCP, UDP, SCTP. Default: TCP."
  type        = string
  default     = "TCP"
}

variable "backend_port_name" {
  description = "Default name of the port used by the backend service (for session affinity/connection tracking)."
  type        = string
  default     = "tcp" # Default to TCP port name
}

variable "backend_timeout_sec" {
  description = "Default timeout in seconds for backend connections."
  type        = number
  default     = 10
}

variable "backend_log_sample_rate" {
  description = "Default backend service log sample rate (0.0 to 1.0)."
  type        = number
  default     = 1.0
}

variable "backend_service_connection_draining_timeout_sec" {
  description = "Default time in seconds to wait for connections to terminate before removing a backend instance. Set to null to use the service default or not configure."
  type        = number
  default     = null
  nullable    = true
}

variable "backend_service_locality_lb_policy" {
  description = "Default locality load balancing policy for the backend service. Options: MAGLEV, WEIGHTED_MAGLEV."
  type        = string
  default     = "MAGLEV"
}

# Optional defaults for backend service connection tracking
variable "backend_service_connection_tracking_idle_timeout_sec" {
  description = "Default connection tracking idle timeout in seconds."
  type        = number
  default     = null
  nullable    = true
}

variable "backend_service_connection_tracking_persist_conn_on_unhealthy" {
  description = "Default behavior for persisting connections on unhealthy backends. Options: NEVER_PERSIST, ALWAYS_PERSIST, DEFAULT_FOR_PROTOCOL."
  type        = string
  default     = null
  nullable    = true
}

variable "backend_service_connection_tracking_track_per_session" {
  description = "Default flag to track connections per session."
  type        = bool
  default     = null
  nullable    = true
}


# Optional defaults for backend service failover policy
variable "backend_service_failover_disable_conn_drain" {
  description = "Default flag to disable connection draining on failover."
  type        = bool
  default     = null
  nullable    = true
}

variable "backend_service_failover_drop_traffic_if_unhealthy" {
  description = "Default flag to drop traffic if all backends are unhealthy."
  type        = bool
  default     = null
  nullable    = true
}

variable "backend_service_failover_ratio" {
  description = "Default failover ratio for the backend service."
  type        = number
  default     = null
  nullable    = true
}

variable "backend_service_session_affinity" {
  description = "Default session affinity for the backend service."
  type        = string
  default     = "NONE" # NONE, CLIENT_IP, CLIENT_IP_NO_DESTINATION, CLIENT_IP_PORT_PROTO, CLIENT_IP_PROTO
}


# Default health check configuration values (used if health_check_name is null)
variable "health_check_check_interval_sec" {
  description = "Default health check interval in seconds."
  type        = number
  default     = 5
}

variable "health_check_timeout_sec" {
  description = "Default health check timeout in seconds."
  type        = number
  default     = 5
}

variable "health_check_healthy_threshold" {
  description = "Default number of consecutive successful health checks required for a backend to be considered healthy."
  type        = number
  default     = 2
}

variable "health_check_unhealthy_threshold" {
  description = "Default number of consecutive failed health checks required for a backend to be considered unhealthy."
  type        = number
  default     = 2
}

variable "health_check_enable_logging" {
  description = "Default flag to enable logging for the health check."
  type        = bool
  default     = false
}

# Default configuration for auto-created TCP health checks (if tcp block is not specified in YAML)
variable "health_check_tcp_port" {
  description = "Default port for auto-created TCP health checks."
  type        = number
  default     = null # Use null to rely on module's default or serving port
  nullable    = true
}

variable "health_check_tcp_port_specification" {
  description = "Default port specification for auto-created TCP health checks (USE_FIXED_PORT, USE_NAMED_PORT, USE_SERVING_PORT)."
  type        = string
  default     = "USE_SERVING_PORT"
}


# Default forwarding rule configuration values
variable "forwarding_rule_protocol" {
  description = "Default protocol for the forwarding rule (listener). Common for NLB: TCP, UDP, SCTP. Default: TCP."
  type        = string
  default     = "TCP"
}

variable "forwarding_rule_ports" {
  description = "Default list of ports for the forwarding rule. Set to null to listen on all ports (only for TCP/UDP)."
  type        = list(string)
  default     = null
  nullable    = true
}

variable "forwarding_rule_address" {
  description = "Default IP address (name or self_link) for the forwarding rule. Set to null for an ephemeral IP."
  type        = string
  default     = null
  nullable    = true
}

variable "forwarding_rule_description" {
  description = "Default description for forwarding rules if not specified in YAML."
  type        = string
  default     = null # Set a default description string here or keep null
  nullable    = true
}

variable "forwarding_rule_ipv6" {
  description = "Default setting for enabling IPv6 on forwarding rules if not specified in YAML."
  type        = bool
  default     = false
}

variable "forwarding_rule_name_override" {
  description = "Default name override for forwarding rules if not specified in YAML. Use with caution, module usually handles naming."
  type        = string
  default     = null
  nullable    = true
}

variable "forwarding_rule_subnetwork" {
  description = "Default subnetwork for forwarding rules if not specified in YAML. Required for IPv6."
  type        = string
  default     = null
  nullable    = true
}

variable "backend_service_connection_tracking" {
  description = "Default connection tracking policy block. If set, this policy is applied if no 'connection_tracking' block is present in the YAML. Set to null to not configure connection tracking by default."
  type = object({
    idle_timeout_sec          = optional(number)
    persist_conn_on_unhealthy = optional(string)
    track_per_session         = optional(bool)
  })
  default  = null # Default to not enabling connection tracking unless specified in YAML
  nullable = true
}

variable "backend_service_failover_config" {
  description = "Default failover policy block. If set, this policy is applied if no 'failover_config' block is present in the YAML. Set to null to not configure failover by default."
  type = object({
    disable_conn_drain        = optional(bool)
    drop_traffic_if_unhealthy = optional(bool)
    ratio                     = optional(number)
  })
  default  = null # Default to not enabling failover unless specified in YAML
  nullable = true
}

variable "backend_group_self_link_format" {
  description = "Format string used to construct the backend instance group self_link."
  type        = string
  default     = "projects/%s/regions/%s/instanceGroups/%s"
}

variable "backend_item_failover" {
  description = "Default failover setting for a backend item if not specified in YAML."
  type        = bool
  default     = false
}

variable "backend_item_description" {
  description = "Default description for a backend item if not specified in YAML."
  type        = string
  default     = "Terraform managed backend group association."
}

# Default for the entire forwarding_rules map if not specified in YAML.
# This defines the default rule key and structure.
variable "forwarding_rules_map" {
  description = "Default map of forwarding rules if none are specified in YAML. Defaults to a single rule with key ''."
  type = map(object({
    address     = optional(string)
    description = optional(string)
    ipv6        = optional(bool)
    name        = optional(string)
    ports       = optional(list(string))
    protocol    = optional(string)
    subnetwork  = optional(string)
  }))
  default = {
    "" = {} # Default to one rule with an empty object configuration
  }
}

# Fallback default for ipv6 attribute in forwarding rule if not in YAML AND var.forwarding_rule_ipv6 is null
variable "forwarding_rule_ipv6_fallback" {
  description = "Fallback default value for forwarding rule ipv6 attribute if not specified in YAML or var.forwarding_rule_ipv6."
  type        = bool
  default     = false
}

# Fallback default for health check tcp port specification if not in YAML AND var.health_check_tcp_port_specification is null
variable "health_check_tcp_port_spec_fallback" {
  description = "Fallback default value for auto-created TCP health check port_specification if not specified in YAML or var.health_check_tcp_port_specification."
  type        = string
  default     = "USE_SERVING_PORT"
}

variable "health_check_tcp_request" {
  description = "Default request string for auto-created TCP health checks."
  type        = string
  default     = null
  nullable    = true
}

variable "health_check_tcp_response" {
  description = "Default expected response string for auto-created TCP health checks."
  type        = string
  default     = null
  nullable    = true
}

variable "health_check_tcp_proxy_header" {
  description = "Default proxy header for auto-created TCP health checks (NONE, PROXY_V1)."
  type        = string
  default     = null
  nullable    = true
}

variable "health_check_config" {
  description = "Default value for the entire health_check block if not provided in the YAML. Should be null to indicate absence unless a specific default structure is desired."
  type        = any
  default     = null
}

variable "hc_grpc_config" {
  description = "Default value for the health check's GRPC configuration if not specified. Typically null."
  type        = any # Or a more specific object type if you have one defined
  default     = null
}

variable "hc_http_config" {
  description = "Default value for the health check's HTTP configuration if not specified. Typically null."
  type        = any # Or a more specific object type
  default     = null
}

variable "hc_http2_config" {
  description = "Default value for the health check's HTTP2 configuration if not specified. Typically null."
  type        = any # Or a more specific object type
  default     = null
}

variable "hc_https_config" {
  description = "Default value for the health check's HTTPS configuration if not specified. Typically null."
  type        = any # Or a more specific object type
  default     = null
}

variable "hc_ssl_config" {
  description = "Default value for the health check's SSL configuration if not specified. Typically null."
  type        = any # Or a more specific object type
  default     = null
}

variable "hc_tcp_config" {
  description = "Default value for the health check's TCP configuration if not specified (and not applying the global default TCP). Typically null."
  type        = any # Or a more specific object type
  default     = null
}

variable "health_check_name_override" {
  description = "Default value for health_check.name if an existing health check name is not provided in the YAML. Typically null to let the module auto-create or use other logic."
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Location of YAML files holding NLB configuration values."
  type        = string
  default     = "../../../../../configuration/consumer-load-balancing/Network/Passthrough/External/config/"
}