## Introduction

[App Engine Standard Environment](https://cloud.google.com/appengine/docs/standard) is a fully managed serverless platform that lets you build and deploy applications that scale automatically.  This Terraform module automates the deployment of applications to the App Engine Standard Environment.

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

    More granular permissions can be used if desired.

4. **Gcloud CLI:** The Google Cloud SDK (`gcloud`) command-line tool installed and configured.

5. **Application Code:**  Have your application code ready for deployment.  This code should be structured according to the requirements of the App Engine Standard Environment and the runtime you are using (e.g., Python, Node.js, Java, Go).

## Let's Get Started! 🚀

With the prerequisites in place and your App Engine configuration files ready, you can now use Terraform to automate the creation of your App Engine Standard Environment deployments.

### Execution Steps

1.  **Clone the repository:**

    * Clone the repository from the offcial cloud-networking config solutions git repository
    ```bash
    git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
    cd cloudnetworking-config-solution/configuration/producer/MRS/
    ```

2. **Configure YAML files:**

    * Create YAML files defining the properties of each App Engine service and version you want to deploy. These files should be stored in the `configuration/consumer/AppEngine/StandardAppEngine/config` folder within this repository.

    * Each YAML file can define multiple services and versions.  The structure of the YAML files is crucial and is described in the [Inputs](#inputs) section below.  Ensure all *required* settings are present in your YAML configuration.

    * For reference on how to structure your App Engine configuration YAML files, see the [Example](#example) section below or refer to sample YAML files in the `configuration/consumer/AppEngine/StandardAppEngine/config` folder. These examples provide templates that you can adapt to your specific needs.

3. **Initialize Terraform:**

    * Open your terminal and navigate to the directory containing the Terraform configuration for App Engine Standard (e.g., `StandardAppEngine`).

    * Run the following command to initialize Terraform:

    ```bash
    terraform init
    ```

4. **Review the Execution Plan:**

    * Use the `terraform plan` command to generate an execution plan. This will show you the changes Terraform will make to your Google Cloud infrastructure:

    ```bash
    terraform plan -var-file=../../../../configuration/consumer/AppEngine/StandardAppEngine/standardappengine.tfvars
    ```

    Carefully review the plan to ensure it aligns with your intended configuration.

5. **Apply the Configuration:**

    * Once you're satisfied with the plan, execute the `terraform apply` command to provision your App Engine services and versions:

    ```bash
    terraform apply -var-file=../../../../configuration/consumer/AppEngine/StandardAppEngine/standardappengine.tfvars
    ```

    Terraform will read the YAML files from the `configuration/consumer/AppEngine/StandardAppEngine/config` folder and create the corresponding App Engine resources in your Google Cloud project.

6. **Monitor and Manage:**

    * After the deployments are complete, you can monitor the status, performance, and logs of your App Engine applications through the Google Cloud Console or using the Google Cloud CLI.

    * Use Terraform to manage updates and changes to your App Engine deployments as needed.

### Example

To help you get started, we've provided examples of YAML configuration files that you can use as templates for your Standard App Engine.

* **Minimal YAML (Mandatory Fields Only):**
This minimal example includes only the essential fields required to create a Standard App Engine.

```yaml
# StandardAppEngine/config/instance1.yaml
project_id: <project-id>
service: "my-standard-app"
version_id: "v1"
runtime: "python312"

deployment:
  files:
    - name: "app.yaml"
      source_url: "https://storage.googleapis.com/<bucket-name>/<object>"  # Replace with your actual bucket and file in the format "https://storage.googleapis.com/test.appspot.com/hello_world/app.yaml"

entrypoint:
  shell: "python main.py"  # Or your actual startup command

delete_service_on_destroy: true

vpc_access_connector:
  - name: "serverless-vpc"

automatic_scaling:
  - max_concurrent_requests: 4
    standard_scheduler_settings:
      target_cpu_utilization: 90
```

* **Comprehensive YAML (All Available Fields):**
This comprehensive example showcases all available fields, allowing you to customize your MRS instance with advanced settings for performance, availability and network configuration.
```yaml
# StandardAppEngine/config/instance2.yaml

project_id: <project-id>
service: "instance1-service"
version_id: "v1"
runtime: "python37"

# Optional: App Engine APIs
app_engine_apis: true

# Optional: Runtime API Version
runtime_api_version: "1"

# Optional: Service Account
service_account: "abc@bcd.com"

# Optional: Threadsafe
threadsafe: true

# Optional: Inbound Services
inbound_services:
  - INBOUND_SERVICE_MAIL

# Optional: Instance Class
instance_class: "F2"

# Optional: Automatic Scaling Configuration
automatic_scaling:
  max_concurrent_requests: 10
  max_idle_instances: 5
  max_pending_latency: "10s"
  min_idle_instances: 2
  min_pending_latency: "5s"
  standard_scheduler_settings:
    target_cpu_utilization: 0.8
    target_throughput_utilization: 0.9
    min_instances: 1
    max_instances: 10

# Optional: Delete service and all enclosed versions on destroy
delete_service_on_destroy: true

# Required: Deployment configuration
deployment:
  # Optional: Zip file to deploy
  zip:
    source_url: "gs://your-bucket/your-app.zip"
    files_count: 1000

  # Optional: Files to deploy
  files:
    - name: "app.yaml"
      sha1_sum: "your-sha1-checksum"
      source_url: "gs://your-bucket/app.yaml"

# Optional: Environment variables
env_variables:
  MY_VAR: "my-value"

# Required: Configures the entrypoint to the application
entrypoint:
  shell: "python main.py"

# Optional: Handlers for requests to a specific host
handlers:
  - auth_fail_action: AUTH_FAIL_ACTION_REDIRECT
    login: LOGIN_REQUIRED
    redirect_http_response_code: REDIRECT_HTTP_RESPONSE_CODE_302
    script:
      script_path: "auto"
    security_level: SECURE_ALWAYS
    url_regex: "/secure"
    static_files:
      path: "static"
      upload_path_regex: ".*\\.txt"
      http_headers:
        X-Foo: "bar"
      mime_type: "text/plain"
      expiration: "7d"
      require_matching_file: true
      application_readable: true

# Optional: Configuration for third-party Python runtime libraries
libraries:
  - name: "django"
    version: "1.11"

# Optional: VPC access connector specification
#vpc_access_connector:
#  name: "projects/your-project-id/locations/your-region/connectors/your-connector-name"
#  egress_setting: "ALL_TRAFFIC"

# Optional: If set to true, the service will be deleted if it is the last version
noop_on_destroy: true
```

## Important Notes

* **This README is a starting point.** Customize it to include specific details about your App Engine Standard applications and relevant configurations.

* **Refer to the official Google Cloud App Engine documentation** for the most up-to-date information and best practices: [https://cloud.google.com/appengine/docs/standard](https://cloud.google.com/appengine/docs/standard)

* **Order of Execution:** While not strictly required, it's generally a good practice to complete the 01-organization stage before deploying App Engine applications. This ensures that your project is properly set up, necessary APIs are enabled, and basic IAM permissions are in place.

* **Secret Manager:** It is *crucial* to use Google Cloud Secret Manager to store and manage sensitive information (API keys, database passwords, etc.).  Do *not* hardcode secrets in your YAML files or Terraform variables.  The example YAML files show placeholders; you'll need to configure Secret Manager and retrieve the secret values within your Terraform configuration.

* **Required Settings:** Pay close attention to *required* settings in `variables.tf`. These *must* be present in your YAML configuration files.  Missing required settings will lead to deployment errors.

* **Deployment Method:** Choose either `zip` or `bucket` for your deployments.  The examples show both.  Make sure your application code is uploaded to the specified Cloud Storage bucket if you are using the `bucket` deployment method.

* **Scaling:** Properly configure the `automatic_scaling` settings to match your application's needs.  Incorrect scaling settings can lead to performance issues or increased costs. You can configure only one of automatic scaling , basic scaling or manual scaling.

* **Health Checks:** Define `liveness_check` and `readiness_check` in your YAML files to ensure App Engine can monitor the health of your application.

* **Troubleshooting:** If you encounter errors during the deployment process, verify that all prerequisites are satisfied, the dependencies between stages are correctly configured, and your YAML configuration files are valid.  Check the Google Cloud Console or the deployment logs for more detailed error information.
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_appengine_standard_instance"></a> [appengine\_standard\_instance](#module\_appengine\_standard\_instance) | ../../../../../modules/app_engine/standard | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_app_engine_apis"></a> [app\_engine\_apis](#input\_app\_engine\_apis) | Enable AppEngine APIs for new services | `bool` | `null` | no |
| <a name="input_app_engine_application"></a> [app\_engine\_application](#input\_app\_engine\_application) | Default value for 'app\_engine\_application' block if missing in YAML. | `map(any)` | `{}` | no |
| <a name="input_auth_domain"></a> [auth\_domain](#input\_auth\_domain) | Default App Engine auth\_domain if app\_engine\_application.auth\_domain is missing in YAML. | `string` | `null` | no |
| <a name="input_automatic_scaling"></a> [automatic\_scaling](#input\_automatic\_scaling) | Default value for 'automatic\_scaling' block if missing in YAML. | `any` | `null` | no |
| <a name="input_basic_scaling"></a> [basic\_scaling](#input\_basic\_scaling) | Default value for 'basic\_scaling' block if missing in YAML. | `any` | `null` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing AppEngine YAML configuration files | `string` | `"../../../../../configuration/consumer/Serverless/AppEngine/Standard/config"` | no |
| <a name="input_create_app_engine_application"></a> [create\_app\_engine\_application](#input\_create\_app\_engine\_application) | Whether to create the google\_app\_engine\_application resource. | `bool` | `false` | no |
| <a name="input_create_app_version"></a> [create\_app\_version](#input\_create\_app\_version) | Whether to create appengine standard app version | `bool` | `true` | no |
| <a name="input_create_dispatch_rules"></a> [create\_dispatch\_rules](#input\_create\_dispatch\_rules) | Whether to create AppEngine dispatch rules. | `bool` | `false` | no |
| <a name="input_create_domain_mappings"></a> [create\_domain\_mappings](#input\_create\_domain\_mappings) | Whether to create AppEngine domain mappings. | `bool` | `false` | no |
| <a name="input_create_firewall_rules"></a> [create\_firewall\_rules](#input\_create\_firewall\_rules) | Whether to create AppEngine firewall rules. | `bool` | `false` | no |
| <a name="input_create_network_settings"></a> [create\_network\_settings](#input\_create\_network\_settings) | Whether to create AppEngine service network settings. | `bool` | `false` | no |
| <a name="input_create_split_traffic"></a> [create\_split\_traffic](#input\_create\_split\_traffic) | Whether to create AppEngine service split traffic settings. | `bool` | `false` | no |
| <a name="input_create_vpc_connector"></a> [create\_vpc\_connector](#input\_create\_vpc\_connector) | Whether to create a vpc access connector. | `bool` | `false` | no |
| <a name="input_database_type"></a> [database\_type](#input\_database\_type) | Default App Engine database\_type if app\_engine\_application.database\_type is missing in YAML. | `string` | `null` | no |
| <a name="input_delete_service_on_destroy"></a> [delete\_service\_on\_destroy](#input\_delete\_service\_on\_destroy) | Whether to delete the service when destroying the resource(default:true). | `bool` | `true` | no |
| <a name="input_deployment"></a> [deployment](#input\_deployment) | Default value for 'deployment' block if missing in YAML. | `any` | `null` | no |
| <a name="input_dispatch_rules"></a> [dispatch\_rules](#input\_dispatch\_rules) | Default value for 'dispatch\_rules' list if missing in YAML. | `list(any)` | `[]` | no |
| <a name="input_domain_mappings"></a> [domain\_mappings](#input\_domain\_mappings) | Default value for 'domain\_mappings' list if missing in YAML. | `list(any)` | `[]` | no |
| <a name="input_entrypoint"></a> [entrypoint](#input\_entrypoint) | Default value for 'entrypoint' block if missing in YAML. | `any` | `null` | no |
| <a name="input_entrypoint_shell"></a> [entrypoint\_shell](#input\_entrypoint\_shell) | entrypointshell | `string` | `null` | no |
| <a name="input_env_variables"></a> [env\_variables](#input\_env\_variables) | Environment variables for the AppEngine version(default:{}). | `map(string)` | `null` | no |
| <a name="input_feature_settings"></a> [feature\_settings](#input\_feature\_settings) | Default App Engine feature\_settings block if app\_engine\_application.feature\_settings is missing in YAML. | `any` | `null` | no |
| <a name="input_firewall_rules"></a> [firewall\_rules](#input\_firewall\_rules) | Default value for 'firewall\_rules' list if missing in YAML. | `list(any)` | `[]` | no |
| <a name="input_handlers"></a> [handlers](#input\_handlers) | Default value for 'handlers' if missing in YAML. | `list(any)` | `[]` | no |
| <a name="input_iap"></a> [iap](#input\_iap) | Default App Engine iap block if app\_engine\_application.iap is missing in YAML. | `any` | `null` | no |
| <a name="input_inbound_services"></a> [inbound\_services](#input\_inbound\_services) | (Optional)A list of inbound services that this service will receive traffic from. An empty list means it will not receive traffic from other services | `list(string)` | `null` | no |
| <a name="input_instance_class"></a> [instance\_class](#input\_instance\_class) | Instance class for the AppEngine version(default:F1). | `string` | `"F2"` | no |
| <a name="input_labels"></a> [labels](#input\_labels) | Labels to apply to the AppEngine version | `map(string)` | `{}` | no |
| <a name="input_libraries"></a> [libraries](#input\_libraries) | Default value for 'libraries' if missing in YAML. | `list(any)` | `[]` | no |
| <a name="input_libraries_name"></a> [libraries\_name](#input\_libraries\_name) | libraries name | `string` | `null` | no |
| <a name="input_libraries_version"></a> [libraries\_version](#input\_libraries\_version) | libraries version | `string` | `null` | no |
| <a name="input_manual_scaling"></a> [manual\_scaling](#input\_manual\_scaling) | Default value for 'manual\_scaling' block if missing in YAML. | `any` | `null` | no |
| <a name="input_network_settings"></a> [network\_settings](#input\_network\_settings) | Default value for 'network\_settings' block if missing in YAML. | `any` | `null` | no |
| <a name="input_noop_on_destroy"></a> [noop\_on\_destroy](#input\_noop\_on\_destroy) | If set to true,then the application version will not be deleted | `bool` | `null` | no |
| <a name="input_runtime_api_version"></a> [runtime\_api\_version](#input\_runtime\_api\_version) | AppEngine runtime API version | `string` | `null` | no |
| <a name="input_service_account"></a> [service\_account](#input\_service\_account) | Service account to be used by the AppEngine version(defaultstotheAppEnginedefaultserviceaccount). | `string` | `null` | no |
| <a name="input_serving_status"></a> [serving\_status](#input\_serving\_status) | Default App Engine serving\_status if app\_engine\_application.serving\_status is missing in YAML. | `string` | `null` | no |
| <a name="input_split_traffic"></a> [split\_traffic](#input\_split\_traffic) | Default value for 'split\_traffic' block if missing in YAML. | `any` | `null` | no |
| <a name="input_threadsafe"></a> [threadsafe](#input\_threadsafe) | Whether the application is threadsafe(default:true).Deprecatedinnewerruntimes. | `bool` | `null` | no |
| <a name="input_version_id"></a> [version\_id](#input\_version\_id) | Default value for the version\_id of missing in the YAML | `string` | `"v1"` | no |
| <a name="input_vpc_access_connector"></a> [vpc\_access\_connector](#input\_vpc\_access\_connector) | Default value for 'vpc\_access\_connector' block if missing in YAML. | `any` | `null` | no |
| <a name="input_vpc_connector_details"></a> [vpc\_connector\_details](#input\_vpc\_connector\_details) | Default value for 'vpc\_connector\_details' block if missing in YAML. | `any` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_all_instance_details"></a> [all\_instance\_details](#output\_all\_instance\_details) | Map of instance keys to an object containing key outputs for that instance. |
| <a name="output_instance_application_urls"></a> [instance\_application\_urls](#output\_instance\_application\_urls) | Map of instance keys to the default URL of the App Engine application associated with that instance. |
| <a name="output_instance_domain_mapping_details"></a> [instance\_domain\_mapping\_details](#output\_instance\_domain\_mapping\_details) | Map of instance keys to the domain mapping resource records created by that instance. |
| <a name="output_instance_service_urls"></a> [instance\_service\_urls](#output\_instance\_service\_urls) | Map of instance keys to the map of service names to their URLs for that instance. |
| <a name="output_instance_vpc_connector_names"></a> [instance\_vpc\_connector\_names](#output\_instance\_vpc\_connector\_names) | Map of instance keys to the name of the VPC Access Connector created by that instance (if any). |
| <a name="output_instance_vpc_connector_self_links"></a> [instance\_vpc\_connector\_self\_links](#output\_instance\_vpc\_connector\_self\_links) | Map of instance keys to the self-link of the VPC Access Connector created by that instance (if any). |
<!-- END_TF_DOCS -->
