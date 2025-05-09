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

resource "google_app_engine_application" "app" {
  count          = var.create_application ? 1 : 0
  project        = var.project_id
  location_id    = var.location_id
  auth_domain    = var.auth_domain
  database_type  = var.database_type
  serving_status = var.serving_status
  dynamic "feature_settings" {
    for_each = var.feature_settings != null ? [var.feature_settings] : []
    content {
      split_health_checks = feature_settings.value.split_health_checks
    }
  }
  dynamic "iap" {
    for_each = var.iap != null ? [var.iap] : []
    content {
      oauth2_client_id     = iap.value.oauth2_client_id
      oauth2_client_secret = iap.value.oauth2_client_secret
      enabled              = iap.value.enabled
    }
  }
}
resource "google_app_engine_application_url_dispatch_rules" "dispatch" {
  count   = var.create_dispatch_rules ? 1 : 0
  project = var.project_id
  dynamic "dispatch_rules" {
    for_each = var.dispatch_rules
    content {
      domain  = dispatch_rules.value.domain
      path    = dispatch_rules.value.path
      service = dispatch_rules.value.service
    }
  }
  depends_on = [google_app_engine_flexible_app_version.flexible]
}
resource "google_app_engine_domain_mapping" "mapping" {
  for_each          = var.create_domain_mappings ? { for index, config in var.domain_mappings : index => config } : {}
  project           = var.project_id
  domain_name       = each.value.domain_name
  override_strategy = try(each.value.override_strategy, "STRICT")
  dynamic "ssl_settings" {
    for_each = each.value.ssl_settings != null && can(each.value.ssl_settings.ssl_management_type) ? [each.value.ssl_settings] : []
    content {
      certificate_id      = try(ssl_settings.value.certificate_id, null)
      ssl_management_type = ssl_settings.value.ssl_management_type
    }
  }
  depends_on = [google_app_engine_flexible_app_version.flexible]
}
resource "google_app_engine_firewall_rule" "firewall" {
  for_each     = var.create_firewall_rules ? { for index, config in var.firewall_rules : index => config } : {}
  project      = var.project_id
  source_range = each.value.source_range
  action       = each.value.action
  priority     = try(each.value.priority, null)
  description  = try(each.value.description, null)
  depends_on   = [google_app_engine_application.app[0]]
}
resource "google_app_engine_flexible_app_version" "flexible" {
  provider                     = google-beta
  for_each                     = var.services
  project                      = var.project_id
  service                      = each.value.service
  runtime                      = each.value.runtime
  version_id                   = each.value.version_id
  instance_class               = try(each.value.instance_class, null)
  runtime_api_version          = try(each.value.runtime_api_version, null)
  runtime_channel              = try(each.value.runtime_channel, null)
  runtime_main_executable_path = try(each.value.runtime_main_executable_path, null)
  service_account              = try(each.value.service_account, null)
  serving_status               = try(each.value.serving_status, null)
  nobuild_files_regex          = try(each.value.nobuild_files_regex, "\\.git$")
  delete_service_on_destroy    = try(each.value.delete_service_on_destroy, false)
  noop_on_destroy              = try(each.value.noop_on_destroy, false)
  beta_settings                = try(each.value.beta_settings, {})
  inbound_services             = try(each.value.inbound_services, null)
  dynamic "entrypoint" {
    for_each = each.value.entrypoint != null ? [each.value.entrypoint] : []
    content {
      shell = entrypoint.value.shell
    }
  }
  dynamic "liveness_check" {
    for_each = each.value.liveness_check != null ? [each.value.liveness_check] : []
    content {
      path              = liveness_check.value.path
      host              = try(liveness_check.value.host, null)
      failure_threshold = try(liveness_check.value.failure_threshold, 3)
      success_threshold = try(liveness_check.value.success_threshold, 2)
      check_interval    = try(liveness_check.value.check_interval, "30s")
      timeout           = try(liveness_check.value.timeout, "4s")
      initial_delay     = try(liveness_check.value.initial_delay, "300s")
    }
  }
  dynamic "readiness_check" {
    for_each = each.value.readiness_check != null ? [each.value.readiness_check] : []
    content {
      path              = readiness_check.value.path
      host              = try(readiness_check.value.host, null)
      failure_threshold = try(readiness_check.value.failure_threshold, 3)
      success_threshold = try(readiness_check.value.success_threshold, 2)
      check_interval    = try(readiness_check.value.check_interval, "10s")
      timeout           = try(readiness_check.value.timeout, "4s")
      app_start_timeout = try(readiness_check.value.app_start_timeout, "600s")
    }
  }
  dynamic "network" {
    for_each = each.value.network != null ? [each.value.network] : []
    content {
      name             = network.value.name
      subnetwork       = try(network.value.subnetwork, null)
      forwarded_ports  = try(network.value.forwarded_ports, [])
      instance_tag     = try(network.value.instance_tag, null)
      session_affinity = try(network.value.session_affinity, null)
      instance_ip_mode = try(network.value.instance_ip_mode, null)
    }
  }
  dynamic "resources" {
    for_each = each.value.resources != null ? [each.value.resources] : []
    content {
      cpu       = try(resources.value.cpu, 2)
      disk_gb   = try(resources.value.disk_gb, 20)
      memory_gb = try(resources.value.memory_gb, 4)
      dynamic "volumes" {
        for_each = resources.value.volumes != null ? resources.value.volumes : []
        content {
          name        = try(volumes.value.name, "my-volume")
          volume_type = try(volumes.value.volume_type, "tmpfs")
          size_gb     = try(volumes.value.size_gb, 3)
        }
      }
    }
  }
  dynamic "flexible_runtime_settings" {
    for_each = each.value.flexible_runtime_settings != null ? [each.value.flexible_runtime_settings] : []
    content {
      operating_system = try(flexible_runtime_settings.value.operating_system, null)
      runtime_version  = try(flexible_runtime_settings.value.runtime_version, null)
    }
  }
  dynamic "automatic_scaling" {
    for_each = each.value.automatic_scaling != null ? [each.value.automatic_scaling] : []
    content {
      cool_down_period        = try(automatic_scaling.value.cool_down_period, null)
      max_concurrent_requests = try(automatic_scaling.value.max_concurrent_requests, null)
      max_total_instances     = try(automatic_scaling.value.max_total_instances, null)
      min_total_instances     = try(automatic_scaling.value.min_total_instances, null)
      max_idle_instances      = try(automatic_scaling.value.max_idle_instances, null)
      min_idle_instances      = try(automatic_scaling.value.min_idle_instances, null)
      max_pending_latency     = try(automatic_scaling.value.max_pending_latency, null)
      min_pending_latency     = try(automatic_scaling.value.min_pending_latency, null)

      dynamic "cpu_utilization" {
        for_each = automatic_scaling.value.cpu_utilization != null ? [automatic_scaling.value.cpu_utilization] : []
        content {
          target_utilization        = cpu_utilization.value.target_utilization
          aggregation_window_length = try(cpu_utilization.value.aggregation_window_length, null)
        }
      }
      dynamic "disk_utilization" {
        for_each = automatic_scaling.value.disk_utilization != null ? [automatic_scaling.value.disk_utilization] : []
        content {
          target_read_bytes_per_second  = try(disk_utilization.value.target_read_bytes_per_second, null)
          target_read_ops_per_second    = try(disk_utilization.value.target_read_ops_per_second, null)
          target_write_bytes_per_second = try(disk_utilization.value.target_write_bytes_per_second, null)
          target_write_ops_per_second   = try(disk_utilization.value.target_write_ops_per_second, null)
        }
      }
      dynamic "network_utilization" {
        for_each = automatic_scaling.value.network_utilization != null ? [automatic_scaling.value.network_utilization] : []
        content {
          target_received_bytes_per_second   = try(network_utilization.value.target_received_bytes_per_second, null)
          target_received_packets_per_second = try(network_utilization.value.target_received_packets_per_second, null)
          target_sent_bytes_per_second       = try(network_utilization.value.target_sent_bytes_per_second, null)
          target_sent_packets_per_second     = try(network_utilization.value.target_sent_packets_per_second, null)
        }
      }
      dynamic "request_utilization" {
        for_each = automatic_scaling.value.request_utilization != null ? [automatic_scaling.value.request_utilization] : []
        content {
          target_concurrent_requests      = try(request_utilization.value.target_concurrent_requests, null)
          target_request_count_per_second = try(request_utilization.value.target_request_count_per_second, null)
        }
      }
    }
  }
  dynamic "manual_scaling" {
    for_each = each.value.manual_scaling != null ? [each.value.manual_scaling] : []
    content {
      instances = manual_scaling.value.instances
    }
  }
  dynamic "endpoints_api_service" {
    for_each = each.value.endpoints_api_service != null ? [each.value.endpoints_api_service] : []
    content {
      name                   = endpoints_api_service.value.name
      config_id              = try(endpoints_api_service.value.config_id, null)
      rollout_strategy       = try(endpoints_api_service.value.rollout_strategy, null)
      disable_trace_sampling = try(endpoints_api_service.value.disable_trace_sampling, null)
    }
  }
  dynamic "deployment" {
    for_each = each.value.deployment != null ? [each.value.deployment] : []
    content {
      dynamic "container" {
        for_each = deployment.value.container != null ? [deployment.value.container] : []
        content {
          image = deployment.value.container.image
        }
      }
      dynamic "files" {
        for_each = deployment.value.files != null ? [deployment.value.files] : []
        content {
          name       = files.value.name
          source_url = files.value.source_url
          sha1_sum   = try(files.value.sha1_sum, null)
        }
      }
      dynamic "zip" {
        for_each = deployment.value.zip != null ? [deployment.value.zip] : []
        content {
          source_url  = deployment.value.zip.source_url
          files_count = try(deployment.value.zip.files_count, null)
        }
      }
      dynamic "cloud_build_options" {
        for_each = deployment.value.cloud_build_options != null ? [deployment.value.cloud_build_options] : []
        content {
          app_yaml_path       = deployment.value.cloud_build_options.app_yaml_path
          cloud_build_timeout = try(deployment.value.cloud_build_options.cloud_build_timeout, null)
        }
      }
    }
  }
  env_variables = try(each.value.env_variables, {})
}
resource "google_app_engine_service_network_settings" "network_settings" {
  for_each = var.create_network_settings ? { for k, v in var.services : k => v } : {}
  project  = var.project_id
  service  = each.key
  dynamic "network_settings" {
    for_each = each.value.network_settings != null ? [each.value.network_settings] : []
    content {
      ingress_traffic_allowed = try(network_settings.value.ingress_traffic_allowed, null)
    }
  }
  depends_on = [google_app_engine_flexible_app_version.flexible]
}
resource "google_app_engine_service_split_traffic" "split_traffic" {
  for_each        = var.create_split_traffic ? { for k, v in var.services : k => v } : {}
  project         = var.project_id
  service         = each.key
  migrate_traffic = try(each.value.split_traffic.migrate_traffic, false)
  dynamic "split" {
    for_each = each.value.split_traffic != null ? [each.value.split_traffic] : []
    content {
      shard_by    = try(split.value.shard_by, null)
      allocations = try(split.value.allocations, null)
    }
  }
  depends_on = [google_app_engine_flexible_app_version.flexible]
}