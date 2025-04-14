## Introduction

[App Engine Flexible Environment](https://cloud.google.com/appengine/docs/flexible) is a managed serverless platform that lets you build and deploy applications with more control over the underlying infrastructure compared to the standard environment. This Terraform module automates the deployment of applications to the App Engine Flexible Environment.

## Prerequisites

Before deploying your App Engine applications with this module, ensure you have completed the following prerequisites:

1. **Prior Stages (Optional but Recommended):** While not strictly required, it's highly recommended to have completed the following stages (or their equivalent setup):
    * **01-organization:** This stage typically handles project setup, API enablement, and basic IAM configurations.  While not strictly necessary for App Engine, it's good practice to have a well-defined project and appropriate APIs enabled.

2. **API Enablement:** Ensure the following Google Cloud APIs are enabled in your project:

    * App Engine Admin API (`appengine.googleapis.com`)
    * Cloud Build API (`cloudbuild.googleapis.com`)
    * Artifact Registry API (`artifactregistry.googleapis.com`)
    * Compute Engine API (`compute.googleapis.com`)
    * Cloud Storage API (`storage-api.googleapis.com`)

3. **IAM Permissions:** Grant the following IAM roles to the user or service account that will be running Terraform:

    * **App Engine Admin** (`roles/appengine.appAdmin`): This role grants full control over App Engine applications.
    * **Cloud Build Editor** (`roles/cloudbuild.builds.editor`): This role allows for creating and managing Cloud Build builds, which you might need for building container images.
    * **Artifact Registry Reader** (`roles/artifactregistry.reader`): This role allows for reading and accessing container images from Artifact Registry.
    * **Compute Engine Network Viewer** (`roles/compute.networkViewer`): This role allows for viewing network configurations, which might be necessary for App Engine deployments.
    * **Storage Object Viewer** (`roles/storage.objectViewer`): This role allows for viewing and downloading objects from Cloud Storage buckets.

* **Note** : Service Account running the App engine must have following permission. It could either be a default compute service account or a service account created by an user.

     * **Logs Configuration Writer (roles/logging.configWriter)**

    More granular permissions can be used if desired.
4. **Gcloud CLI:** The Google Cloud SDK (`gcloud`) command-line tool installed and configured.

5. **Application Code:**  Have your application code ready for deployment.  This code should be structured according to the requirements of the App Engine Flexible Environment and the runtime you are using (e.g., custom runtimes with Docker).

## Let's Get Started! 🚀

With the prerequisites in place and your App Engine configuration files ready, you can now use Terraform to automate the creation of your App Engine Flexible Environment deployments.

### Execution Steps

1.  **Clone the repository:**

    * Clone the repository from the offcial cloud-networking config solutions git repository
    ```bash
    git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
    cd cloudnetworking-config-solution/configuration/producer/MRS/
    ```

2. **Configure YAML files:**

    * Create YAML files defining the properties of each App Engine service and version you want to deploy. These files should be stored in the `configuration/consumer/AppEngine/FlexibleAppEngine/config` folder within this repository.

    * Each YAML file can define multiple services and versions.  The structure of the YAML files is crucial and is described in the [Inputs](#inputs) section below.  Ensure all *required* settings are present in your YAML configuration.

    * For reference on how to structure your App Engine configuration YAML files, see the [Example](#example) section below or refer to sample YAML files in the `configuration/consumer/AppEngine/FlexibleAppEngine/config` folder. These examples provide templates that you can adapt to your specific needs.

3. **Initialize Terraform:**

    * Open your terminal and navigate to the directory containing the Terraform configuration for App Engine Flexible (e.g., `FlexibleAppEngine`).

    * Run the following command to initialize Terraform:

    ```bash
    terraform init
    ```

4. **Review the Execution Plan:**

    * Use the `terraform plan` command to generate an execution plan. This will show you the changes Terraform will make to your Google Cloud infrastructure:

    ```bash
    terraform plan -var-file=../../../../configuration/consumer/AppEngine/FlexibleAppEngine/flexibleappengine.tfvars
    ```

    Carefully review the plan to ensure it aligns with your intended configuration.

5. **Apply the Configuration:**

    * Once you're satisfied with the plan, execute the `terraform apply` command to provision your App Engine services and versions:

    ```bash
    terraform apply -var-file=../../../../configuration/consumer/AppEngine/FlexibleAppEngine/flexibleappengine.tfvars
    ```

    Terraform will read the YAML files from the `configuration/consumer/AppEngine/FlexibleAppEngine/config` folder and create the corresponding App Engine resources in your Google Cloud project.

6. **Monitor and Manage:**

    * After the deployments are complete, you can monitor the status, performance, and logs of your App Engine applications through the Google Cloud Console or using the Google Cloud CLI.

    * Use Terraform to manage updates and changes to your App Engine deployments as needed.

### Example

To help you get started, we've provided examples of YAML configuration files that you can use as templates for your Flexible App Engine.

* **Minimal YAML (Mandatory Fields Only):**
This minimal example includes only the essential fields required to create a Flexible App Engine.

```yaml
# FlexibleAppEngine/config/instance1.yaml
project_id: <project-id>
service: "my-flexible-app"
version_id: "v1"
runtime: "python39"

readiness_check:
  path: "/"

liveness_check:
  path: "/"

automatic_scaling:
  cpu_utilization:
    target_utilization: 0.6
```

* **Comprehensive YAML (All Available Fields):**
This comprehensive example showcases all available fields, allowing you to customize your MRS instance with advanced settings for performance, availability and network configuration.

```yaml
# FlexibleAppEngine/config/instance2.yaml
project_id: "your-project-id"
service: "instance1-service"
version_id: "v1"
runtime: "python39"

readiness_check:
  path: "/ready"
  host: "example.com"
  failure_threshold: 3
  success_threshold: 2
  check_interval: "10s"
  timeout: "5s"
  app_start_timeout: "600s"

liveness_check:
  path: "/health"
  host: "example.com"
  failure_threshold: 5
  success_threshold: 3
  check_interval: "15s"
  timeout: "6s"
  initial_delay: "400s"

inbound_services:
  - INBOUND_SERVICE_MAIL
  - INBOUND_SERVICE_XMPP_MESSAGE

instance_class: "F2"

network:
  forwarded_ports:
    - "8080:80"
  instance_tag: "my-instance-tag"
  name: "default"
  subnetwork: "my-subnet"
  session_affinity: true

resources:
  cpu: 2
  disk_gb: 50
  memory_gb: 4
  volumes:
    - name: "my-volume"
      volume_type: "tmpfs"
      size_gb: 10

runtime_channel: "stable"

flexible_runtime_settings:
  operating_system: "linux"
  runtime_version: "1.23"

beta_settings:
  my-beta-setting: "enabled"

serving_status: "SERVING"

runtime_api_version: "1"

handlers:
  - url_regex: "/api"
    security_level: SECURE_ALWAYS
    login: LOGIN_REQUIRED
    auth_fail_action: AUTH_FAIL_ACTION_UNAUTHORIZED
    redirect_http_response_code: REDIRECT_HTTP_RESPONSE_CODE_301
    script:
      script_path: "api.php"
    static_files:
      path: "public"
      upload_path_regex: ".*\\.html"
      http_headers:
        Cache-Control: "public, max-age=3600"
      mime_type: "text/html"
      expiration: "1d"
      require_matching_file: true
      application_readable: true

runtime_main_executable_path: "app.py"

service_account: "your-service-account-email"

api_config:
  auth_fail_action: AUTH_FAIL_ACTION_UNAUTHORIZED
  login: LOGIN_ADMIN
  script: "auth.php"
  security_level: SECURE_NEVER
  url: "/auth"

env_variables:
  MY_ENV_VAR: "my-value"

default_expiration: "2d"

nobuild_files_regex: ".*\\.go"

deployment:
  zip:
    source_url: "gs://your-bucket/app.zip"
    files_count: 1000
  files:
    - name: "index.php"
      sha1_sum: "your-sha1-checksum"
      source_url: "gs://your-bucket/index.php"
  container:
    image: "us-docker.pkg.dev/cloudrun/container/hello"
  cloud_build_options:
    app_yaml_path: "app.yaml"
    cloud_build_timeout: "600s"

endpoints_api_service:
  name: "my-endpoints-service"
  config_id: "2023-05-01r1"
  rollout_strategy: "MANAGED"
  disable_trace_sampling: true

entrypoint:
  shell: "node server.js"

vpc_access_connector:
  name: "projects/your-project-id/locations/your-region/connectors/your-connector-name"

automatic_scaling:
  cool_down_period: "180s"
  max_concurrent_requests: 50
  max_idle_instances: 5
  max_total_instances: 10
  max_pending_latency: "20s"
  min_idle_instances: 1
  min_total_instances: 1
  min_pending_latency: "10s"
  cpu_utilization:
    aggregation_window_length: "60s"
    target_utilization: 0.8
  request_utilization:
    target_request_count_per_second: 2
    target_concurrent_requests: 40
  disk_utilization:
    target_write_bytes_per_second: 1000000
    target_write_ops_per_second: 500
    target_read_bytes_per_second: 2000000
    target_read_ops_per_second: 2000
  network_utilization:
    target_sent_bytes_per_second: 500000
    target_sent_packets_per_second: 250
    target_received_bytes_per_second: 1000000
    target_received_packets_per_second: 500

manual_scaling:
  instances: 2

noop_on_destroy: true
delete_service_on_destroy: false
```

## Important Notes

* **This README is a starting point.** Customize it to include specific details about your App Engine Flexible applications and relevant configurations.

* **Refer to the official Google Cloud App Engine documentation** for the most up-to-date information and best practices: [https://cloud.google.com/appengine/docs/flexible](https://cloud.google.com/appengine/docs/flexible)

* **Order of Execution:** While not strictly required for basic deployments, it's generally a good practice to complete the 01-organization stage before deploying App Engine Flexible applications. This ensures that your project is properly set up, necessary APIs are enabled, and basic IAM permissions are in place.  For more complex deployments, especially those involving custom networking or VPC access, completing the 02-networking stage is also recommended.

* **Secret Manager:** It is *crucial* to use Google Cloud Secret Manager to store and manage sensitive information (API keys, database passwords, etc.).  Do *not* hardcode secrets in your YAML files or Terraform variables.  The example YAML files show placeholders; you'll need to configure Secret Manager and retrieve the secret values within your Terraform configuration.

* **Required Settings:** Pay close attention to *required* settings in `variables.tf`. These *must* be present in your YAML configuration files.  Missing required settings will lead to deployment errors.

* **Deployment Method:** The Flexible Environment typically uses `files` (for individual files like `Dockerfile` and `app.yaml`) or `zip` for deployment.  The examples show both.  Make sure your application code and configuration files are uploaded to the specified Cloud Storage bucket.

* **Scaling:** Configure the `automatic_scaling` settings to match your application's needs.  The Flexible Environment offers more scaling options compared to the Standard Environment, allowing you to fine-tune the scaling behavior of your application. you can have automatic scaling or manual scaling.

* **Networking:**  You can use the `network` settings to customize the network configuration, including using VPC networks and subnets. This gives you more control over the network environment of your application.

* **Health Checks:** Define `liveness_check` and `readiness_check` in your YAML files to ensure App Engine can monitor the health of your application.

* **Troubleshooting:** If you encounter errors during the deployment process, verify that all prerequisites are satisfied, the dependencies between stages are correctly configured, and your YAML configuration files are valid.  Check the Google Cloud Console or the deployment logs for more detailed error information.
<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

No providers.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_flexible_app_engine_instance"></a> [flexible\_app\_engine\_instance](#module\_flexible\_app\_engine\_instance) | ../../../../../modules/app_engine/flexible | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auth_domain"></a> [auth\_domain](#input\_auth\_domain) | Default authentication domain. | `string` | `null` | no |
| <a name="input_automatic_scaling"></a> [automatic\_scaling](#input\_automatic\_scaling) | Default automatic\_scaling block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_beta_settings"></a> [beta\_settings](#input\_beta\_settings) | Default beta\_settings map if not specified in service YAML. | `map(string)` | `{}` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing App Engine Flexible YAML configuration files | `string` | `"../../../../../configuration/consumer/Serverless/AppEngine/Flexible/config"` | no |
| <a name="input_create_application"></a> [create\_application](#input\_create\_application) | Default value for whether to create the App Engine application. | `bool` | `false` | no |
| <a name="input_create_dispatch_rules"></a> [create\_dispatch\_rules](#input\_create\_dispatch\_rules) | Default value for whether to create dispatch rules. | `bool` | `false` | no |
| <a name="input_create_domain_mappings"></a> [create\_domain\_mappings](#input\_create\_domain\_mappings) | Default value for whether to create domain mappings. | `bool` | `false` | no |
| <a name="input_create_firewall_rules"></a> [create\_firewall\_rules](#input\_create\_firewall\_rules) | Default value for whether to create firewall rules. | `bool` | `false` | no |
| <a name="input_create_network_settings"></a> [create\_network\_settings](#input\_create\_network\_settings) | Default value for whether to create network settings. | `bool` | `false` | no |
| <a name="input_create_split_traffic"></a> [create\_split\_traffic](#input\_create\_split\_traffic) | Default value for whether to create split traffic settings. | `bool` | `false` | no |
| <a name="input_database_type"></a> [database\_type](#input\_database\_type) | Default database type. | `string` | `null` | no |
| <a name="input_delete_service_on_destroy"></a> [delete\_service\_on\_destroy](#input\_delete\_service\_on\_destroy) | Default delete\_service\_on\_destroy flag if not specified in service YAML. | `bool` | `false` | no |
| <a name="input_deployment"></a> [deployment](#input\_deployment) | Default deployment block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_dispatch_rules"></a> [dispatch\_rules](#input\_dispatch\_rules) | Default dispatch rules (if none in YAML). | <pre>list(object({<br/>    domain  = string<br/>    path    = string<br/>    service = string<br/>  }))</pre> | `[]` | no |
| <a name="input_domain_mappings"></a> [domain\_mappings](#input\_domain\_mappings) | Domain mappings for the App Engine application. | <pre>list(object({<br/>    domain_name = string<br/>    ssl_settings = optional(object({<br/>      certificate_id      = optional(string)<br/>      ssl_management_type = string<br/>    }))<br/>  }))</pre> | `[]` | no |
| <a name="input_endpoints_api_service"></a> [endpoints\_api\_service](#input\_endpoints\_api\_service) | Default endpoints\_api\_service block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_entrypoint"></a> [entrypoint](#input\_entrypoint) | Default entrypoint block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_env_variables"></a> [env\_variables](#input\_env\_variables) | Default env\_variables map if not specified in service YAML. | `map(string)` | `{}` | no |
| <a name="input_feature_settings"></a> [feature\_settings](#input\_feature\_settings) | Default feature settings (if none in YAML). | <pre>object({<br/>    split_health_checks = optional(bool, true)<br/>  })</pre> | `null` | no |
| <a name="input_firewall_rules"></a> [firewall\_rules](#input\_firewall\_rules) | Default firewall rules (if none in YAML). | <pre>list(object({<br/>    source_range = string<br/>    action       = string<br/>    priority     = optional(number)<br/>    description  = optional(string)<br/>  }))</pre> | `[]` | no |
| <a name="input_flexible_runtime_settings"></a> [flexible\_runtime\_settings](#input\_flexible\_runtime\_settings) | Default flexible\_runtime\_settings block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_iap_settings"></a> [iap\_settings](#input\_iap\_settings) | Default IAP settings (if none in YAML). | <pre>object({<br/>    enabled              = bool<br/>    oauth2_client_id     = string<br/>    oauth2_client_secret = string<br/>  })</pre> | `null` | no |
| <a name="input_inbound_services"></a> [inbound\_services](#input\_inbound\_services) | Default inbound\_services list if not specified in service YAML. | `list(string)` | `null` | no |
| <a name="input_instance_class"></a> [instance\_class](#input\_instance\_class) | Default instance class if not specified in service YAML. | `string` | `null` | no |
| <a name="input_labels"></a> [labels](#input\_labels) | Default labels map if not specified in service YAML. | `map(string)` | `{}` | no |
| <a name="input_liveness_check"></a> [liveness\_check](#input\_liveness\_check) | Default liveness\_check block if not specified in service YAML. | `any` | `"/"` | no |
| <a name="input_location_id"></a> [location\_id](#input\_location\_id) | Default location for the App Engine application. | `string` | `""` | no |
| <a name="input_manual_scaling"></a> [manual\_scaling](#input\_manual\_scaling) | Default manual\_scaling block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_network"></a> [network](#input\_network) | Default network block (for service VM) if not specified in service YAML. | `any` | `null` | no |
| <a name="input_network_settings_block"></a> [network\_settings\_block](#input\_network\_settings\_block) | Default network\_settings block (for service network settings resource) if not specified in service YAML. | `any` | `null` | no |
| <a name="input_nobuild_files_regex"></a> [nobuild\_files\_regex](#input\_nobuild\_files\_regex) | Default nobuild\_files\_regex if not specified in service YAML. | `string` | `null` | no |
| <a name="input_noop_on_destroy"></a> [noop\_on\_destroy](#input\_noop\_on\_destroy) | Default noop\_on\_destroy flag if not specified in service YAML. | `bool` | `false` | no |
| <a name="input_readiness_check"></a> [readiness\_check](#input\_readiness\_check) | Default readiness\_check block if not specified in service YAML. | `any` | `"/"` | no |
| <a name="input_resources"></a> [resources](#input\_resources) | Default resources block if not specified in service YAML. | `any` | `null` | no |
| <a name="input_runtime_api_version"></a> [runtime\_api\_version](#input\_runtime\_api\_version) | Default runtime\_api\_version if not specified in service YAML. | `string` | `"1"` | no |
| <a name="input_runtime_channel"></a> [runtime\_channel](#input\_runtime\_channel) | Default runtime\_channel if not specified in service YAML. | `string` | `null` | no |
| <a name="input_runtime_main_executable_path"></a> [runtime\_main\_executable\_path](#input\_runtime\_main\_executable\_path) | Default runtime\_main\_executable\_path if not specified in service YAML. | `string` | `null` | no |
| <a name="input_service_account"></a> [service\_account](#input\_service\_account) | Default service\_account if not specified in service YAML. If null, App Engine default SA is often used. | `string` | `null` | no |
| <a name="input_service_serving_status"></a> [service\_serving\_status](#input\_service\_serving\_status) | Default serving\_status for a specific service if not specified in its YAML (distinct from app-level default). | `string` | `null` | no |
| <a name="input_serving_status"></a> [serving\_status](#input\_serving\_status) | Default serving status. | `string` | `null` | no |
| <a name="input_split_traffic_block"></a> [split\_traffic\_block](#input\_split\_traffic\_block) | Default split\_traffic block (for service split traffic resource) if not specified in service YAML. | `any` | `null` | no |
| <a name="input_version_id"></a> [version\_id](#input\_version\_id) | Default version id for the app engine application | `string` | `"v1"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_instance_application_urls"></a> [instance\_application\_urls](#output\_instance\_application\_urls) | Map of instance keys (e.g., 'project\_service') to the corresponding App Engine application URL output by that module instance. Should ideally be the same URL for instances targeting the same project. |
| <a name="output_instance_domain_mapping_resource_records"></a> [instance\_domain\_mapping\_resource\_records](#output\_instance\_domain\_mapping\_resource\_records) | Map of instance keys to the domain mapping resource records output by that specific module instance. |
| <a name="output_instance_service_urls"></a> [instance\_service\_urls](#output\_instance\_service\_urls) | Map of instance keys to the map of service URLs provided by that instance. Each inner map usually contains the URL for the single service defined in the corresponding YAML. |
<!-- END_TF_DOCS -->
