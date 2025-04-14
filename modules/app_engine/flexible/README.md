# Terraform Google App Engine Flexible Module

This module creates a Google App Engine Flexible environment application, including services, versions, firewall rules, domain mappings, and dispatch rules.

## Usage

```terraform
module "app_engine_flexible" {
  source = "./modules/app_engine_flexible"

  project_id  = "your-project-id"
  location_id = "us-central"

  services = {
    default = {
      service    = "default"
      version_id = "v1"
      runtime    = "nodejs" # Flexible uses container images; runtime is a hint
      entrypoint = {
        shell = "npm start"
      }
      deployment = {
        container = {
          image = "gcr.io/your-project-id/your-image:latest"
        }
      }
      network = {
        name = "default"
      }
       readiness_check = {
        path = "/ready"
      }
      liveness_check = {
        path = "/healthz"
      }
      env_variables = {
        NODE_ENV = "production"
      }
        resources = {
        cpu = 2
        memory_gb = 4
        disk_gb = 20
      }
    }
  }
    firewall_rules = [
    {
      source_range = "*"
      action       = "ALLOW"
    },
  ]
}
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.8 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 6.20.0, < 7.0.0 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | >= 6.20.0, < 7.0.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | >= 6.20.0, < 7.0.0 |
| <a name="provider_google-beta"></a> [google-beta](#provider\_google-beta) | >= 6.20.0, < 7.0.0 |

## Resources

| Name | Type |
|------|------|
| [google-beta_google_app_engine_flexible_app_version.flexible](https://registry.terraform.io/providers/hashicorp/google-beta/latest/docs/resources/google_app_engine_flexible_app_version) | resource |
| [google_app_engine_application.app](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_application) | resource |
| [google_app_engine_application_url_dispatch_rules.dispatch](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_application_url_dispatch_rules) | resource |
| [google_app_engine_domain_mapping.mapping](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_domain_mapping) | resource |
| [google_app_engine_firewall_rule.firewall](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_firewall_rule) | resource |
| [google_app_engine_service_network_settings.network_settings](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_service_network_settings) | resource |
| [google_app_engine_service_split_traffic.split_traffic](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_service_split_traffic) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auth_domain"></a> [auth\_domain](#input\_auth\_domain) | The domain to authenticate users with. | `string` | `null` | no |
| <a name="input_create_application"></a> [create\_application](#input\_create\_application) | Whether to create the App Engine application. Defaults to true. | `bool` | `true` | no |
| <a name="input_create_dispatch_rules"></a> [create\_dispatch\_rules](#input\_create\_dispatch\_rules) | Whether to create the dispatch rules. Defaults to true. | `bool` | `true` | no |
| <a name="input_create_domain_mappings"></a> [create\_domain\_mappings](#input\_create\_domain\_mappings) | Whether to create the domain mappings. Defaults to true. | `bool` | `true` | no |
| <a name="input_create_firewall_rules"></a> [create\_firewall\_rules](#input\_create\_firewall\_rules) | Whether to create the firewall rules. Defaults to true. | `bool` | `true` | no |
| <a name="input_create_network_settings"></a> [create\_network\_settings](#input\_create\_network\_settings) | Whether to create network settings. Defaults to true. | `bool` | `true` | no |
| <a name="input_create_split_traffic"></a> [create\_split\_traffic](#input\_create\_split\_traffic) | Whether to create split traffic settings. Defaults to true. | `bool` | `true` | no |
| <a name="input_database_type"></a> [database\_type](#input\_database\_type) | The type of database to use. | `string` | `null` | no |
| <a name="input_dispatch_rules"></a> [dispatch\_rules](#input\_dispatch\_rules) | dispatch rules for the app engine | <pre>list(object({<br/>    domain  = string<br/>    path    = string<br/>    service = string<br/>  }))</pre> | `[]` | no |
| <a name="input_domain_mappings"></a> [domain\_mappings](#input\_domain\_mappings) | Domain mappings for the App Engine application. | <pre>list(object({<br/>    domain_name       = string<br/>    override_strategy = optional(string, "STRICT")<br/>    ssl_settings = optional(object({<br/>      certificate_id      = optional(string)<br/>      ssl_management_type = string<br/>    }))<br/>  }))</pre> | `[]` | no |
| <a name="input_feature_settings"></a> [feature\_settings](#input\_feature\_settings) | Feature settings for the App Engine application. | <pre>object({<br/>    split_health_checks = optional(bool, true)<br/>  })</pre> | `null` | no |
| <a name="input_firewall_rules"></a> [firewall\_rules](#input\_firewall\_rules) | Firewall rules for the App Engine application. | <pre>list(object({<br/>    source_range = string<br/>    action       = string<br/>    priority     = optional(number)<br/>    description  = optional(string)<br/>  }))</pre> | `[]` | no |
| <a name="input_iap"></a> [iap](#input\_iap) | Identity-Aware Proxy settings. | <pre>object({<br/>    enabled              = bool<br/>    oauth2_client_id     = string<br/>    oauth2_client_secret = string<br/>  })</pre> | `null` | no |
| <a name="input_location_id"></a> [location\_id](#input\_location\_id) | The location to deploy the App Engine application. | `string` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The ID of the project in which to deploy the App Engine application. | `string` | n/a | yes |
| <a name="input_services"></a> [services](#input\_services) | A map of service configurations. | <pre>map(object({<br/>    service                      = string<br/>    runtime                      = string<br/>    version_id                   = string<br/>    instance_class               = optional(string)<br/>    runtime_api_version          = optional(string)<br/>    runtime_channel              = optional(string)<br/>    runtime_main_executable_path = optional(string)<br/>    service_account              = optional(string)<br/>    serving_status               = optional(string)<br/>    nobuild_files_regex          = optional(string)<br/>    delete_service_on_destroy    = optional(bool, false)<br/>    noop_on_destroy              = optional(bool, false)<br/>    beta_settings                = optional(map(string))<br/>    inbound_services             = optional(list(string))<br/>    labels                       = optional(map(string))<br/>    entrypoint = optional(object({<br/>      shell = string<br/>    }))<br/>    liveness_check = optional(object({<br/>      path              = string<br/>      host              = optional(string)<br/>      failure_threshold = optional(number)<br/>      success_threshold = optional(number)<br/>      check_interval    = optional(string)<br/>      timeout           = optional(string)<br/>      initial_delay     = optional(string)<br/>    }))<br/>    readiness_check = optional(object({<br/>      path              = string<br/>      host              = optional(string)<br/>      failure_threshold = optional(number)<br/>      success_threshold = optional(number)<br/>      check_interval    = optional(string)<br/>      timeout           = optional(string)<br/>      app_start_timeout = optional(string)<br/>    }))<br/>    network = optional(object({<br/>      name             = string<br/>      subnetwork       = optional(string)<br/>      forwarded_ports  = optional(list(string))<br/>      instance_tag     = optional(string)<br/>      session_affinity = optional(bool)<br/>      instance_ip_mode = optional(string)<br/>    }))<br/>    resources = optional(object({<br/>      cpu       = optional(number)<br/>      disk_gb   = optional(number)<br/>      memory_gb = optional(number)<br/>      volumes = optional(list(object({<br/>        name        = string<br/>        volume_type = string<br/>        size_gb     = number<br/>      })))<br/>    }))<br/>    flexible_runtime_settings = optional(object({<br/>      operating_system = optional(string)<br/>      runtime_version  = optional(string)<br/>    }))<br/>    automatic_scaling = optional(object({<br/>      cool_down_period        = optional(string)<br/>      max_concurrent_requests = optional(number)<br/>      max_total_instances     = optional(number)<br/>      min_total_instances     = optional(number)<br/>      max_idle_instances      = optional(number)<br/>      min_idle_instances      = optional(number)<br/>      max_pending_latency     = optional(string)<br/>      min_pending_latency     = optional(string)<br/>      cpu_utilization = optional(object({<br/>        target_utilization        = number<br/>        aggregation_window_length = optional(string)<br/>      }))<br/>      disk_utilization = optional(object({<br/>        target_read_bytes_per_second  = optional(number)<br/>        target_read_ops_per_second    = optional(number)<br/>        target_write_bytes_per_second = optional(number)<br/>        target_write_ops_per_second   = optional(number)<br/>      }))<br/>      network_utilization = optional(object({<br/>        target_received_bytes_per_second   = optional(number)<br/>        target_received_packets_per_second = optional(number)<br/>        target_sent_bytes_per_second       = optional(number)<br/>        target_sent_packets_per_second     = optional(number)<br/>      }))<br/>      request_utilization = optional(object({<br/>        target_concurrent_requests      = optional(number)<br/>        target_request_count_per_second = optional(number)<br/>      }))<br/>    }))<br/>    manual_scaling = optional(object({<br/>      instances = number<br/>    }))<br/>    endpoints_api_service = optional(object({<br/>      name                   = string<br/>      config_id              = optional(string)<br/>      rollout_strategy       = optional(string)<br/>      disable_trace_sampling = optional(bool)<br/>    }))<br/>    deployment = optional(object({<br/>      container = optional(object({<br/>        image = string<br/>      }))<br/>      files = optional(object({<br/>        name       = string<br/>        source_url = string<br/>        sha1_sum   = optional(string)<br/>      }))<br/>      zip = optional(object({<br/>        source_url  = string<br/>        files_count = optional(number)<br/>      }))<br/>      cloud_build_options = optional(object({<br/>        app_yaml_path       = string<br/>        cloud_build_timeout = optional(string)<br/>      }))<br/>    }))<br/>    env_variables = optional(map(string), {})<br/>    network_settings = optional(object({<br/>      ingress_traffic_allowed = optional(string, "INGRESS_TRAFFIC_ALLOWED_ALL")<br/>    }))<br/>    split_traffic = optional(object({<br/>      shard_by        = optional(string, "IP")<br/>      allocations     = map(number)<br/>      migrate_traffic = optional(bool, false)<br/>    }))<br/>  }))</pre> | n/a | yes |
| <a name="input_serving_status"></a> [serving\_status](#input\_serving\_status) | The serving status of the application. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_application_url"></a> [application\_url](#output\_application\_url) | The default URL of the App Engine application. URLs are empty strings if 'create\_application' is false. |
| <a name="output_domain_mapping_resource_records"></a> [domain\_mapping\_resource\_records](#output\_domain\_mapping\_resource\_records) | all domain mapping resource records |
| <a name="output_service_urls"></a> [service\_urls](#output\_service\_urls) | A map of service names to their URLs. URLs are empty strings if 'create\_application' is false. |
<!-- END_TF_DOCS -->
