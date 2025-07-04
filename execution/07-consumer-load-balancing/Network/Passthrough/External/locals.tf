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

locals {
  # Read all YAML files in the specified config folder (excluding files starting with _)
  lb_configs_raw = [
    for file in fileset(var.config_folder_path, "[^_]*.yaml") :
    yamldecode(file("${var.config_folder_path}/${file}"))
  ]

  lb_configs_processed = [
    for lb_config in local.lb_configs_raw : {
      # Keep all original lb_config data
      lb_config = lb_config

      # --- Health Check Configuration ---
      raw_hc_config = try(lb_config.health_check, null)

      # Determine if the health_check block was provided in YAML
      health_check_config_provided = try(lb_config.health_check, null) != null
      # Determine if any protocol config is explicitly provided in YAML raw_hc_config
      health_check_protocol_provided = (
        try(lb_config.health_check, null) != null && (
          try(lb_config.health_check.grpc, null) != null ||
          try(lb_config.health_check.http, null) != null ||
          try(lb_config.health_check.http2, null) != null ||
          try(lb_config.health_check.https, null) != null ||
          try(lb_config.health_check.ssl, null) != null ||
          try(lb_config.health_check.tcp, null) != null
        )
      )
      # Determine if we should apply the default TCP config (if no existing HC name and no protocol in YAML)
      apply_default_tcp = (
        try(lb_config.health_check.name, null) == null && # Not using an existing HC name from raw_hc_config
        !(                                                # Check if ANY protocol config was explicitly provided in YAML
          try(lb_config.health_check.grpc, null) != null ||
          try(lb_config.health_check.http, null) != null ||
          try(lb_config.health_check.http2, null) != null ||
          try(lb_config.health_check.https, null) != null ||
          try(lb_config.health_check.ssl, null) != null ||
          try(lb_config.health_check.tcp, null) != null
        )
      )
      # --- End Health Check Configuration ---
    }
  ]


  # Flatten the list and apply defaults/transformations using the processed configs
  # This creates a list of simplified configuration objects
  lb_config_list = flatten([
    for processed_config in local.lb_configs_processed : {
      # Access the original lb_config data
      lb_config = processed_config.lb_config

      # Access the calculated intermediate health check flags
      raw_hc_config                  = processed_config.raw_hc_config
      health_check_config_provided   = processed_config.health_check_config_provided
      health_check_protocol_provided = processed_config.health_check_protocol_provided
      apply_default_tcp              = processed_config.apply_default_tcp

      # Required fields from YAML
      name                = processed_config.lb_config.name
      project_id          = processed_config.lb_config.project_id
      region              = processed_config.lb_config.region
      backend_definitions = processed_config.lb_config.backends

      # Optional fields with defaults from variables
      description = try(processed_config.lb_config.description, var.description) # Top-level LB description
      labels      = try(processed_config.lb_config.labels, var.labels)

      # Backend Service Configuration - Map to the module's expected object structure
      # Attributes here are for the backend *service*, not individual backend items
      backend_service_config = {
        # Map flat attributes with defaults from root variables
        protocol                        = try(processed_config.lb_config.backend_service.protocol, var.backend_protocol)
        port_name                       = try(processed_config.lb_config.backend_service.port_name, var.backend_port_name)
        timeout_sec                     = try(processed_config.lb_config.backend_service.timeout_sec, var.backend_timeout_sec)
        connection_draining_timeout_sec = try(processed_config.lb_config.backend_service.connection_draining_timeout_sec, var.backend_service_connection_draining_timeout_sec)
        log_sample_rate                 = try(processed_config.lb_config.backend_service.log_sample_rate, var.backend_log_sample_rate)
        locality_lb_policy              = try(processed_config.lb_config.backend_service.locality_lb_policy, var.backend_service_locality_lb_policy)
        session_affinity                = try(processed_config.lb_config.backend_service.session_affinity, var.backend_service_session_affinity)

        connection_tracking = try(processed_config.lb_config.backend_service.connection_tracking, var.backend_service_connection_tracking)
        failover_config     = try(processed_config.lb_config.backend_service.failover_config, var.backend_service_failover_config)
      }

      # --- Health Check Configuration ---
      # Map to the module's health_check_config object using the calculated intermediate values.
      health_check_config = {
        # Map top-level health check attributes with defaults from root variables
        check_interval_sec = try(processed_config.raw_hc_config.check_interval_sec, var.health_check_check_interval_sec)
        description        = try(processed_config.raw_hc_config.description, var.description) # Can use main description or add a specific var if needed
        enable_logging     = try(processed_config.raw_hc_config.enable_logging, var.health_check_enable_logging)
        healthy_threshold  = try(processed_config.raw_hc_config.healthy_threshold, var.health_check_healthy_threshold)
        # name - module handles the name for auto-created HC based on main name variable
        timeout_sec         = try(processed_config.raw_hc_config.timeout_sec, var.health_check_timeout_sec)
        unhealthy_threshold = try(processed_config.raw_hc_config.unhealthy_threshold, var.health_check_unhealthy_threshold)

        # Map nested protocol-specific configurations
        # Use YAML value if present, otherwise use null. Access via raw_hc_config.
        grpc  = try(processed_config.raw_hc_config.grpc, var.hc_grpc_config)
        http  = try(processed_config.raw_hc_config.http, var.hc_http_config)
        http2 = try(processed_config.raw_hc_config.http2, var.hc_http2_config)
        https = try(processed_config.raw_hc_config.https, var.hc_https_config)
        ssl   = try(processed_config.raw_hc_config.ssl, var.hc_ssl_config)

        # TCP mapping - If default should be applied, use variables. Otherwise, use YAML value if present.
        tcp = processed_config.apply_default_tcp ? {
          # Default TCP config using root variables (when no protocol specified in YAML)
          port               = var.health_check_tcp_port
          port_specification = try(var.health_check_tcp_port_specification, var.health_check_tcp_port_spec_fallback)
          request            = var.health_check_tcp_request
          response           = var.health_check_tcp_response
          proxy_header       = var.health_check_tcp_proxy_header
        } : try(processed_config.raw_hc_config.tcp, var.hc_tcp_config) # Use variable for default null
      }

      health_check_name = try(processed_config.raw_hc_config.name, var.health_check_name_override)

      # --- End Health Check Configuration ---

      # Forwarding Rules Configuration - Map to the module's expected map(object) structure
      # Iterate over the map of rules provided in YAML (defaulting to the map defined in a variable if none)
      forwarding_rules_config = {
        for rule_key, rule_config in try(processed_config.lb_config.forwarding_rules, var.forwarding_rules_map) : rule_key => {
          # Map individual attributes for each rule with defaults
          address     = try(rule_config.address, var.forwarding_rule_address)
          description = try(rule_config.description, var.forwarding_rule_description)
          ipv6        = try(rule_config.ipv6, try(var.forwarding_rule_ipv6, var.forwarding_rule_ipv6_fallback))
          name        = try(rule_config.name, var.forwarding_rule_name_override)
          ports       = try(rule_config.ports, var.forwarding_rule_ports)
          protocol    = try(rule_config.protocol, try(processed_config.lb_config.forwarding_rule_protocol, var.forwarding_rule_protocol))
          subnetwork  = try(rule_config.subnetwork, var.forwarding_rule_subnetwork)
        }
      }
    }
  ])

  # Convert the list of config objects to a map keyed by the load balancer name.
  # This is the final structure passed to the module's for_each.
  lb_map = {
    for lb in local.lb_config_list :
    lb.name => merge(
      lb,
      {
        backends = [
          for backend_def in lb.backend_definitions : {
            group = (
              try(backend_def.group_zone, null) != null ?
              format("projects/%s/zones/%s/instanceGroups/%s",
                lb.project_id,
                backend_def.group_zone,
                backend_def.group_name
              ) :
              format(var.backend_group_self_link_format,
                lb.project_id,
                try(backend_def.group_region, lb.region),
                backend_def.group_name
              )
            ),
            failover    = try(backend_def.failover, var.backend_item_failover),
            description = try(backend_def.description, var.backend_item_description)
          }
        ]
      }
    )
  }
}