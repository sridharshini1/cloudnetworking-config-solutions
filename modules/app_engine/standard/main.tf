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
  count       = var.create_app_engine_application ? 1 : 0
  project     = var.project_id
  location_id = var.location_id

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
    for_each = var.iap != null ? [1] : []
    content {
      enabled              = try(var.iap.enabled, false)
      oauth2_client_id     = var.iap.oauth2_client_id
      oauth2_client_secret = var.iap.oauth2_client_secret
    }
  }
}

resource "google_app_engine_application_url_dispatch_rules" "dispatch" {
  count   = var.create_dispatch_rules && length(var.dispatch_rules) > 0 ? 1 : 0
  project = var.project_id
  dynamic "dispatch_rules" {
    for_each = var.dispatch_rules
    content {
      domain  = dispatch_rules.value.domain
      path    = dispatch_rules.value.path
      service = dispatch_rules.value.service
    }
  }
  depends_on = [google_app_engine_standard_app_version.standard]
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
  depends_on = [google_app_engine_application.app]
}

resource "google_app_engine_firewall_rule" "firewall" {
  for_each     = var.create_firewall_rules ? { for index, config in var.firewall_rules : index => config } : {}
  project      = var.project_id
  source_range = each.value.source_range
  action       = each.value.action
  priority     = try(each.value.priority, null)
  description  = try(each.value.description, null)
  depends_on   = [google_app_engine_application.app]
}

resource "google_app_engine_service_network_settings" "network_settings" {
  for_each = { for k, v in var.services : k => v if v.create_network_settings }
  service  = each.key
  project  = var.project_id
  dynamic "network_settings" {
    for_each = try(each.value.network_settings, {}) != {} ? [each.value.network_settings] : []
    content {
      ingress_traffic_allowed = try(network_settings.value.ingress_traffic_allowed, null)
    }
  }
  depends_on = [google_app_engine_standard_app_version.standard]
}

resource "google_app_engine_service_split_traffic" "split_traffic" {
  for_each        = { for k, v in var.services : k => v if v.create_split_traffic } # Key change!
  service         = each.key
  project         = var.project_id
  migrate_traffic = try(each.value.split_traffic.migrate_traffic, false)

  dynamic "split" {
    for_each = try(each.value.split_traffic.allocations, null) != null ? [each.value.split_traffic] : []
    content {
      shard_by    = try(split.value.shard_by, "IP")
      allocations = try(split.value.allocations, [])
    }
  }
  depends_on = [google_app_engine_standard_app_version.standard]
}

resource "google_app_engine_standard_app_version" "standard" {
  for_each = var.create_app_version ? var.services : {}

  project    = var.project_id
  service    = each.value.service
  version_id = each.value.version_id
  runtime    = each.value.runtime

  app_engine_apis     = try(each.value.app_engine_apis, null)
  runtime_api_version = try(each.value.runtime_api_version, null)
  service_account     = try(each.value.service_account, null)
  threadsafe          = try(each.value.threadsafe, null)
  inbound_services    = try(each.value.inbound_services, []) // Correct Default
  instance_class      = try(each.value.instance_class, null)

  timeouts {
    create = "30m" # Increase create timeout to 30 minutes (example)
    update = "20m" # You can also set update and delete timeouts
    delete = "20m"
  }

  dynamic "automatic_scaling" {
    for_each = each.value.automatic_scaling != null && each.value.basic_scaling == null && each.value.manual_scaling == null ? [each.value.automatic_scaling] : []
    content {
      max_concurrent_requests = try(automatic_scaling.value.max_concurrent_requests, var.automatic_scaling_max_concurrent_requests)
      max_idle_instances      = try(automatic_scaling.value.max_idle_instances, var.automatic_scaling_max_idle_instances)
      max_pending_latency     = try(automatic_scaling.value.max_pending_latency, var.automatic_scaling_max_pending_latency)
      min_idle_instances      = try(automatic_scaling.value.min_idle_instances, var.automatic_scaling_min_idle_instances)
      min_pending_latency     = try(automatic_scaling.value.min_pending_latency, var.automatic_scaling_min_pending_latency)

      dynamic "standard_scheduler_settings" {
        for_each = try(automatic_scaling.value.standard_scheduler_settings, null) != null ? [automatic_scaling.value.standard_scheduler_settings] : []
        content {
          target_cpu_utilization        = try(standard_scheduler_settings.value.target_cpu_utilization, var.automatic_scaling_standard_scheduler_settings_target_cpu_utilization)
          target_throughput_utilization = try(standard_scheduler_settings.value.target_throughput_utilization, var.automatic_scaling_standard_scheduler_settings_target_throughput_utilization)
          min_instances                 = try(standard_scheduler_settings.value.min_instances, var.automatic_scaling_standard_scheduler_settings_min_instances)
          max_instances                 = try(standard_scheduler_settings.value.max_instances, var.automatic_scaling_standard_scheduler_settings_max_instances)
        }
      }
    }
  }

  dynamic "basic_scaling" {
    for_each = each.value.basic_scaling != null && each.value.automatic_scaling == null && each.value.manual_scaling == null ? [each.value.basic_scaling] : []
    content {
      max_instances = try(basic_scaling.value.max_instances, var.basic_scaling_max_instances)
      idle_timeout  = try(basic_scaling.value.idle_timeout, var.basic_scaling_idle_timeout)
    }
  }

  dynamic "manual_scaling" {
    for_each = each.value.manual_scaling != null && each.value.automatic_scaling == null && each.value.basic_scaling == null ? [each.value.manual_scaling] : []
    content {
      instances = try(manual_scaling.value.instances, var.manual_scaling_instances)
    }
  }

  delete_service_on_destroy = try(each.value.delete_service_on_destroy, null)

  dynamic "deployment" {
    for_each = each.value.deployment != null ? [each.value.deployment] : []
    content {
      dynamic "zip" {
        for_each = try(deployment.value.zip.source_url, null) != null ? [deployment.value.zip] : []
        content {
          source_url  = zip.value.source_url
          files_count = try(zip.value.files_count, var.deployment_zip_files_count)
        }
      }

      dynamic "files" {
        for_each = try(deployment.value.files.name, null) != null ? [deployment.value.files] : []
        content {
          name       = files.value.name
          source_url = files.value.source_url
          sha1_sum   = try(files.value.sha1_sum, var.deployment_files_sha1_sum)
        }
      }
    }
  }

  env_variables = try(each.value.env_variables, {}) // Correct Default

  entrypoint {
    shell = try(each.value.entrypoint.shell, null)
  }

  dynamic "handlers" {
    for_each = try(each.value.handlers, {}) != {} ? each.value.handlers : []
    content {
      auth_fail_action            = try(handlers.value.auth_fail_action, null)
      login                       = try(handlers.value.login, null)
      redirect_http_response_code = try(handlers.value.redirect_http_response_code, null)
      dynamic "script" {
        for_each = handlers.value.script != null ? [handlers.value.script] : []
        content {
          script_path = script.value.script_path
        }
      }
      security_level = try(handlers.value.security_level, null)
      url_regex      = try(handlers.value.url_regex, null)

      dynamic "static_files" {
        for_each = handlers.value.static_files != null ? [handlers.value.static_files] : []
        content {
          path                  = try(static_files.value.path, null)
          upload_path_regex     = try(static_files.value.upload_path_regex, null)
          http_headers          = try(static_files.value.http_headers, null)
          mime_type             = try(static_files.value.mime_type, null)
          expiration            = try(static_files.value.expiration, null)
          require_matching_file = try(static_files.value.require_matching_file, null)
          application_readable  = try(static_files.value.application_readable, null)
        }
      }
    }
  }

  dynamic "libraries" {
    for_each = each.value.libraries != null ? each.value.libraries : []
    content {
      name    = try(libraries.value.name, null)
      version = try(libraries.value.version, null)
    }
  }
  dynamic "vpc_access_connector" {
    for_each = each.value.create_vpc_connector == true || each.value.vpc_access_connector != null ? [1] : []
    content {
      name           = each.value.create_vpc_connector == true ? "projects/${local.connector_project}/locations/${local.connector_region}/connectors/${local.connector_name}" : each.value.vpc_access_connector.name
      egress_setting = try(each.value.vpc_connector_details.egress_setting, null)
    }
  }
  noop_on_destroy = try(each.value.noop_on_destroy, var.noop_on_destroy)
  depends_on      = [google_app_engine_application.app, google_vpc_access_connector.connector]
}

# --- IAM Binding (Within the Module) ---

resource "google_project_iam_member" "app_engine_service_account" {
  for_each = {
    for service_name, service_config in var.services :
    service_name => service_config
    if service_config.service_account != null
  }

  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${each.value.service_account}"
}
