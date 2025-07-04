# Terraform Google App Engine Standard Module

This module creates a Google App Engine Standard environment application, including services, versions, firewall rules, domain mappings, and dispatch rules.

## Usage

```terraform
module "app_engine" {
  source = "./modules/app_engine_standard"

  project_id  = "your-project-id"
  location_id = "us-central"

  services = {
    default = {
      service    = "default"
      version_id = "v1"
      runtime    = "python39"
      deployment = {
        zip = {
          source_url = "gs://your-bucket/app.zip"
        }
      }
      env_variables = {
        MY_VARIABLE = "my_value"
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
| [google-beta_google_vpc_access_connector.connector](https://registry.terraform.io/providers/hashicorp/google-beta/latest/docs/resources/google_vpc_access_connector) | resource |
| [google_app_engine_application.app](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_application) | resource |
| [google_app_engine_application_url_dispatch_rules.dispatch](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_application_url_dispatch_rules) | resource |
| [google_app_engine_domain_mapping.mapping](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_domain_mapping) | resource |
| [google_app_engine_firewall_rule.firewall](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_firewall_rule) | resource |
| [google_app_engine_service_network_settings.network_settings](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_service_network_settings) | resource |
| [google_app_engine_service_split_traffic.split_traffic](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_service_split_traffic) | resource |
| [google_app_engine_standard_app_version.standard](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/app_engine_standard_app_version) | resource |
| [google_project_iam_member.app_engine_service_account](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_iam_member) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_app_engine_apis"></a> [app\_engine\_apis](#input\_app\_engine\_apis) | Enable AppEngine APIs for new services | `bool` | `null` | no |
| <a name="input_auth_domain"></a> [auth\_domain](#input\_auth\_domain) | The domain to authenticate users with. | `string` | `null` | no |
| <a name="input_automatic_scaling_max_concurrent_requests"></a> [automatic\_scaling\_max\_concurrent\_requests](#input\_automatic\_scaling\_max\_concurrent\_requests) | Maximum concurrent requests for automatic scaling(default:50). | `number` | `50` | no |
| <a name="input_automatic_scaling_max_idle_instances"></a> [automatic\_scaling\_max\_idle\_instances](#input\_automatic\_scaling\_max\_idle\_instances) | Maximum idle instances for automatic scaling(default:1). | `number` | `1` | no |
| <a name="input_automatic_scaling_max_pending_latency"></a> [automatic\_scaling\_max\_pending\_latency](#input\_automatic\_scaling\_max\_pending\_latency) | Maximum pending latency for automatic scaling(default:30s). | `string` | `"30s"` | no |
| <a name="input_automatic_scaling_min_idle_instances"></a> [automatic\_scaling\_min\_idle\_instances](#input\_automatic\_scaling\_min\_idle\_instances) | Minimum idle instances for automatic scaling(default:0). | `number` | `0` | no |
| <a name="input_automatic_scaling_min_pending_latency"></a> [automatic\_scaling\_min\_pending\_latency](#input\_automatic\_scaling\_min\_pending\_latency) | Minimum pending latency for automatic scaling(default:30ms). | `string` | `"1s"` | no |
| <a name="input_automatic_scaling_standard_scheduler_settings_max_instances"></a> [automatic\_scaling\_standard\_scheduler\_settings\_max\_instances](#input\_automatic\_scaling\_standard\_scheduler\_settings\_max\_instances) | Maximum instances for standard scheduler settings(default:100). | `number` | `100` | no |
| <a name="input_automatic_scaling_standard_scheduler_settings_min_instances"></a> [automatic\_scaling\_standard\_scheduler\_settings\_min\_instances](#input\_automatic\_scaling\_standard\_scheduler\_settings\_min\_instances) | Minimum instances for standard scheduler settings(default:0). | `number` | `0` | no |
| <a name="input_automatic_scaling_standard_scheduler_settings_target_cpu_utilization"></a> [automatic\_scaling\_standard\_scheduler\_settings\_target\_cpu\_utilization](#input\_automatic\_scaling\_standard\_scheduler\_settings\_target\_cpu\_utilization) | Target CPU utilization for standard scheduler settings(default:0.6). | `number` | `0.6` | no |
| <a name="input_automatic_scaling_standard_scheduler_settings_target_throughput_utilization"></a> [automatic\_scaling\_standard\_scheduler\_settings\_target\_throughput\_utilization](#input\_automatic\_scaling\_standard\_scheduler\_settings\_target\_throughput\_utilization) | Target throughput utilization for standard scheduler settings(default:0.6). | `number` | `0.6` | no |
| <a name="input_basic_scaling_idle_timeout"></a> [basic\_scaling\_idle\_timeout](#input\_basic\_scaling\_idle\_timeout) | Idle timeout for basic scaling. | `string` | `null` | no |
| <a name="input_basic_scaling_max_instances"></a> [basic\_scaling\_max\_instances](#input\_basic\_scaling\_max\_instances) | Maximum instances for basic scaling. | `number` | `null` | no |
| <a name="input_connector_name"></a> [connector\_name](#input\_connector\_name) | VPC connector name | `string` | `""` | no |
| <a name="input_create_app_engine_application"></a> [create\_app\_engine\_application](#input\_create\_app\_engine\_application) | Whether to create the google\_app\_engine\_application resource. | `bool` | `false` | no |
| <a name="input_create_app_version"></a> [create\_app\_version](#input\_create\_app\_version) | Whether to create appengine standard app version | `bool` | `true` | no |
| <a name="input_create_dispatch_rules"></a> [create\_dispatch\_rules](#input\_create\_dispatch\_rules) | Whether to create AppEngine dispatch rules. | `bool` | `false` | no |
| <a name="input_create_domain_mappings"></a> [create\_domain\_mappings](#input\_create\_domain\_mappings) | Whether to create AppEngine domain mappings. | `bool` | `false` | no |
| <a name="input_create_firewall_rules"></a> [create\_firewall\_rules](#input\_create\_firewall\_rules) | Whether to create AppEngine firewall rules. | `bool` | `false` | no |
| <a name="input_create_network_settings"></a> [create\_network\_settings](#input\_create\_network\_settings) | Whether to create AppEngine service network settings. | `bool` | `false` | no |
| <a name="input_create_split_traffic"></a> [create\_split\_traffic](#input\_create\_split\_traffic) | Whether to create AppEngine service split traffic settings. | `bool` | `false` | no |
| <a name="input_database_type"></a> [database\_type](#input\_database\_type) | The type of database to use. | `string` | `null` | no |
| <a name="input_delete_service_on_destroy"></a> [delete\_service\_on\_destroy](#input\_delete\_service\_on\_destroy) | Whether to delete the service when destroying the resource(default:true). | `bool` | `false` | no |
| <a name="input_deployment_files_name"></a> [deployment\_files\_name](#input\_deployment\_files\_name) | deployment filesname | `string` | `null` | no |
| <a name="input_deployment_files_sha1_sum"></a> [deployment\_files\_sha1\_sum](#input\_deployment\_files\_sha1\_sum) | deployment files sha1\_sum | `string` | `null` | no |
| <a name="input_deployment_files_source_url"></a> [deployment\_files\_source\_url](#input\_deployment\_files\_source\_url) | deployment files source\_url | `string` | `null` | no |
| <a name="input_deployment_zip_files_count"></a> [deployment\_zip\_files\_count](#input\_deployment\_zip\_files\_count) | Number of files in the deployment ZIP file. | `number` | `null` | no |
| <a name="input_deployment_zip_source_url"></a> [deployment\_zip\_source\_url](#input\_deployment\_zip\_source\_url) | Source URL for the deployment ZIP file. | `string` | `null` | no |
| <a name="input_dispatch_rules"></a> [dispatch\_rules](#input\_dispatch\_rules) | dispatch rules for the appengine | <pre>list(object({<br/>    domain  = string<br/>    path    = string<br/>    service = string<br/>  }))</pre> | `[]` | no |
| <a name="input_domain_mappings"></a> [domain\_mappings](#input\_domain\_mappings) | Domain mappings for the AppEngine application. | <pre>list(object({<br/>    domain_name       = string<br/>    override_strategy = optional(string, "STRICT")<br/>    ssl_settings = optional(object({<br/>      certificate_id      = optional(string)<br/>      ssl_management_type = string<br/>    }))<br/>  }))</pre> | `[]` | no |
| <a name="input_entrypoint_shell"></a> [entrypoint\_shell](#input\_entrypoint\_shell) | entry point shell | `string` | `null` | no |
| <a name="input_env_variables"></a> [env\_variables](#input\_env\_variables) | Environment variables for the AppEngine version(default:{}). | `map(string)` | `null` | no |
| <a name="input_feature_settings"></a> [feature\_settings](#input\_feature\_settings) | Feature settings for the AppEngine application. | <pre>object({<br/>    split_health_checks = optional(bool, true)<br/>  })</pre> | `null` | no |
| <a name="input_firewall_rules"></a> [firewall\_rules](#input\_firewall\_rules) | Firewall rules for the  AppEngine application. | <pre>list(object({<br/>    source_range = string<br/>    action       = string<br/>    priority     = optional(number)<br/>    description  = optional(string)<br/>  }))</pre> | `[]` | no |
| <a name="input_handlers_auth_fail_action"></a> [handlers\_auth\_fail\_action](#input\_handlers\_auth\_fail\_action) | handlers auth fail action | `string` | `null` | no |
| <a name="input_handlers_login"></a> [handlers\_login](#input\_handlers\_login) | handlers login | `string` | `null` | no |
| <a name="input_handlers_redirect_http_response_code"></a> [handlers\_redirect\_http\_response\_code](#input\_handlers\_redirect\_http\_response\_code) | handlers redirect http responsecode | `string` | `null` | no |
| <a name="input_handlers_script_script_path"></a> [handlers\_script\_script\_path](#input\_handlers\_script\_script\_path) | script path | `string` | `"auto"` | no |
| <a name="input_handlers_security_level"></a> [handlers\_security\_level](#input\_handlers\_security\_level) | handlers security level | `string` | `null` | no |
| <a name="input_handlers_static_files_application_readable"></a> [handlers\_static\_files\_application\_readable](#input\_handlers\_static\_files\_application\_readable) | handlers static files application readable | `bool` | `null` | no |
| <a name="input_handlers_static_files_expiration"></a> [handlers\_static\_files\_expiration](#input\_handlers\_static\_files\_expiration) | handlers static files expiration | `string` | `null` | no |
| <a name="input_handlers_static_files_http_headers"></a> [handlers\_static\_files\_http\_headers](#input\_handlers\_static\_files\_http\_headers) | handlers static files http headers | `map(string)` | `null` | no |
| <a name="input_handlers_static_files_mime_type"></a> [handlers\_static\_files\_mime\_type](#input\_handlers\_static\_files\_mime\_type) | handlers static files mime type | `string` | `null` | no |
| <a name="input_handlers_static_files_path"></a> [handlers\_static\_files\_path](#input\_handlers\_static\_files\_path) | handlers static files path | `string` | `null` | no |
| <a name="input_handlers_static_files_require_matching_file"></a> [handlers\_static\_files\_require\_matching\_file](#input\_handlers\_static\_files\_require\_matching\_file) | handlers static files require matching file | `bool` | `null` | no |
| <a name="input_handlers_static_files_upload_path_regex"></a> [handlers\_static\_files\_upload\_path\_regex](#input\_handlers\_static\_files\_upload\_path\_regex) | handlers static files upload path regex | `string` | `null` | no |
| <a name="input_handlers_url_regex"></a> [handlers\_url\_regex](#input\_handlers\_url\_regex) | handlers url regex | `string` | `"/.*"` | no |
| <a name="input_iap"></a> [iap](#input\_iap) | Identity-AwareProxysettings. | <pre>object({<br/>    enabled              = optional(bool, false) #Makeenabledoptional<br/>    oauth2_client_id     = optional(string)      #Makeoptional<br/>    oauth2_client_secret = optional(string)      #Makeoptional<br/>  })</pre> | `null` | no |
| <a name="input_inbound_services"></a> [inbound\_services](#input\_inbound\_services) | (Optional)A list of inbound services that this service will receive traffic from.An empty list means it will not receive traffic from other services | `list(string)` | `null` | no |
| <a name="input_instance_class"></a> [instance\_class](#input\_instance\_class) | Instance class for the AppEngine version(default:F1). | `string` | `"F2"` | no |
| <a name="input_labels"></a> [labels](#input\_labels) | A map of labels to apply to the service | `map(string)` | `{}` | no |
| <a name="input_libraries_name"></a> [libraries\_name](#input\_libraries\_name) | libraries name | `string` | `null` | no |
| <a name="input_libraries_version"></a> [libraries\_version](#input\_libraries\_version) | libraries version | `string` | `null` | no |
| <a name="input_location_id"></a> [location\_id](#input\_location\_id) | The location to deploy the AppEngine application. | `string` | `"us-central1"` | no |
| <a name="input_manual_scaling_instances"></a> [manual\_scaling\_instances](#input\_manual\_scaling\_instances) | Number of instances for manualscaling. | `number` | `null` | no |
| <a name="input_noop_on_destroy"></a> [noop\_on\_destroy](#input\_noop\_on\_destroy) | If set to true,then the application version will not be deleted | `bool` | `null` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The ID of the project in which to deploy the AppEngine application. | `string` | n/a | yes |
| <a name="input_runtime_api_version"></a> [runtime\_api\_version](#input\_runtime\_api\_version) | AppEngine runtime API version | `string` | `null` | no |
| <a name="input_service_account"></a> [service\_account](#input\_service\_account) | Service account to be used by the AppEngine version(defaults to the AppEngine default service account). | `string` | `null` | no |
| <a name="input_services"></a> [services](#input\_services) | n/a | <pre>map(object({<br/>    service             = string<br/>    runtime             = string<br/>    version_id          = optional(string)<br/>    app_engine_apis     = optional(bool, false)<br/>    runtime_api_version = optional(string)<br/>    service_account     = optional(string)<br/>    threadsafe          = optional(bool)<br/>    inbound_services    = optional(list(string))<br/>    instance_class      = optional(string)<br/>    labels              = optional(map(string))<br/>    automatic_scaling = optional(object({<br/>      max_concurrent_requests = optional(number)<br/>      max_idle_instances      = optional(number)<br/>      max_pending_latency     = optional(string)<br/>      min_idle_instances      = optional(number)<br/>      min_pending_latency     = optional(string)<br/>      standard_scheduler_settings = optional(object({<br/>        target_cpu_utilization        = optional(number)<br/>        target_throughput_utilization = optional(number)<br/>        min_instances                 = optional(number)<br/>        max_instances                 = optional(number)<br/>      }))<br/>    }))<br/>    basic_scaling = optional(object({<br/>      max_instances = optional(number)<br/>      idle_timeout  = optional(string)<br/>    }))<br/>    manual_scaling = optional(object({<br/>      instances = optional(number)<br/>    }))<br/>    network_settings = optional(object({<br/>      ingress_traffic_allowed = optional(string)<br/>    }))<br/>    delete_service_on_destroy = optional(bool, false)<br/>    deployment = optional(object({<br/>      zip = optional(object({<br/>        source_url  = string<br/>        files_count = optional(number)<br/>      }))<br/>      files = optional(object({ // SINGLE OBJECT<br/>        name       = string<br/>        source_url = string<br/>        sha1_sum   = optional(string)<br/>      }))<br/>    }))<br/>    env_variables = optional(map(string))<br/>    entrypoint = optional(object({<br/>      shell = string<br/>    }))<br/>    handlers = optional(list(object({<br/>      auth_fail_action            = optional(string)<br/>      login                       = optional(string)<br/>      redirect_http_response_code = optional(string)<br/>      script                      = optional(object({ script_path = string }))<br/>      security_level              = optional(string)<br/>      url_regex                   = optional(string)<br/>      static_files = optional(object({<br/>        path                  = optional(string)<br/>        upload_path_regex     = optional(string)<br/>        http_headers          = optional(map(string))<br/>        mime_type             = optional(string)<br/>        expiration            = optional(string)<br/>        require_matching_file = optional(bool)<br/>        application_readable  = optional(bool)<br/>      }))<br/>    })))<br/>    libraries = optional(list(object({<br/>      name    = optional(string)<br/>      version = optional(string)<br/>    })))<br/>    vpc_access_connector = optional(object({ #Existingconnectordetails<br/>      name = string<br/>    }))<br/>    vpc_connector_details = optional(object({<br/>      name              = string<br/>      host_project_id   = optional(string) # Project ID for the new connector<br/>      region            = optional(string) # Region for the new connector<br/>      ip_cidr_range     = optional(string)<br/>      subnet_name       = optional(string)<br/>      subnet_project_id = optional(string)<br/>      network           = optional(string)<br/>      machine_type      = optional(string)<br/>      min_instances     = optional(number)<br/>      max_instances     = optional(number)<br/>      min_throughput    = optional(number)<br/>      max_throughput    = optional(number)<br/>      egress_setting    = optional(string)<br/><br/>    }))<br/>    create_vpc_connector = optional(bool, false)<br/>    split_traffic = optional(object({<br/>      migrate_traffic = optional(bool, false)<br/>      allocations     = map(number)<br/>      shard_by        = optional(string)<br/>    }))<br/>    create_split_traffic    = optional(bool, false)<br/>    create_dispatch_rules   = optional(bool, false)<br/>    create_network_settings = optional(bool, false)<br/>    noop_on_destroy         = optional(bool, false)<br/>  }))</pre> | `{}` | no |
| <a name="input_serving_status"></a> [serving\_status](#input\_serving\_status) | The serving status of the application. | `string` | `null` | no |
| <a name="input_threadsafe"></a> [threadsafe](#input\_threadsafe) | Whether the application is thread safe(default:true).Deprecated in newer runtimes. | `bool` | `null` | no |
| <a name="input_vpc_access_connector_egress_setting"></a> [vpc\_access\_connector\_egress\_setting](#input\_vpc\_access\_connector\_egress\_setting) | vpc connector egress setting | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_app_engine_standard"></a> [app\_engine\_standard](#output\_app\_engine\_standard) | The configuration details for the standard app engine instances deployed. |
| <a name="output_application_url"></a> [application\_url](#output\_application\_url) | The default URL of the App Engine application. |
| <a name="output_service_urls"></a> [service\_urls](#output\_service\_urls) | A map of service names to their URLs. |
<!-- END_TF_DOCS -->