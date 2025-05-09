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


Name	Description	Type	Default	Required
project_id	The ID of the Google Cloud project where App Engine will be deployed.	string	-	yes
location_id	The location to deploy the App Engine application.	string	-	yes
auth_domain	The domain to use for authentication. Defaults to the project's default domain.	string	null	no
database_type	The type of database to use. Defaults to CLOUD_DATASTORE_COMPATIBILITY.	string	null	no
serving_status	The serving status of the application (SERVING, STOPPED). Defaults to SERVING.	string	null	no
feature_settings	Feature settings for the App Engine application.	object({ split_health_checks = optional(bool, true) })	null	no
iap	Configuration for Identity-Aware Proxy (IAP).	object({ enabled = bool, oauth2_client_id = string, oauth2_client_secret = string })	null	no
dispatch_rules	URL dispatch rules for routing traffic to different services.	list(object({ domain = string, path = string, service = string }))	[]	no
domain_mappings	Domain mappings for the App Engine application.	list(object({ domain_name = string, override_strategy = optional(string, "STRICT"), ssl_settings = optional(object({ certificate_id = optional(string), ssl_management_type = string })) }))	[]	no
firewall_rules	Firewall rules for the App Engine application.	list(object({ source_range = string, action = string, priority = optional(number), description = optional(string) }))	[]	no
services	A map of service configurations. The key is the service name.	map(object({...})) (See below for detailed structure)	{}	no
app_engine_apis	Enable App Engine APIs	bool	true	no
runtime_api_version	App Engine runtime API version	string	"1"	no
service_account	Service account to be used by the App Engine version	string	null	no
threadsafe	Whether the application is threadsafe	bool	true	no
inbound_services	A list of inbound services	list(string)	[]	no
instance_class	Instance class	string	"F1"	no
automatic_scaling_max_concurrent_requests	max_concurrent_requests	number	50	no
automatic_scaling_max_idle_instances	max_idle_instances	number	1	no
automatic_scaling_max_pending_latency	max_pending_latency	string	"30s"	no
automatic_scaling_min_idle_instances	min_idle_instances	number	0	no
automatic_scaling_min_pending_latency	min_pending_latency	string	"30ms"	no
automatic_scaling_standard_scheduler_settings_target_cpu_utilization	target_cpu_utilization	number	0.6	no
automatic_scaling_standard_scheduler_settings_target_throughput_utilization	target_throughput_utilization	number	0.6	no
automatic_scaling_standard_scheduler_settings_min_instances	min_instances	number	0	no
automatic_scaling_standard_scheduler_settings_max_instances	max_instances	number	100	no
basic_scaling_max_instances	max_instances	number	null	no
basic_scaling_idle_timeout	idle_timeout	string	null	no
manual_scaling_instances	instances	number	null	no
delete_service_on_destroy	Whether to delete the service when destroying the resource	bool	true	no
deployment_zip_source_url	source_url	string	null	no
deployment_zip_files_count	files_count	number	null	no
deployment_files_name	name	string	null	no
deployment_files_sha1_sum	sha1_sum	string	null	no
deployment_files_source_url	source_url	string	null	no
env_variables	env_variables	map(string)	{}	no
entrypoint_shell	shell	string	null	no
handlers_auth_fail_action	auth_fail_action	string	null	no
handlers_login	login	string	null	no
handlers_redirect_http_response_code	redirect_http_response_code	string	null	no
handlers_script_script_path	script_path	string	null	no
handlers_security_level	security_level	string	null	no
handlers_url_regex	url_regex	string	null	no
handlers_static_files_path	path	string	null	no
handlers_static_files_upload_path_regex	upload_path_regex	string	null	no
handlers_static_files_http_headers	http_headers	map(string)	null	no
handlers_static_files_mime_type	mime_type	string	null	no
handlers_static_files_expiration	expiration	string	null	no
handlers_static_files_require_matching_file	require_matching_file	bool	null	no
handlers_static_files_application_readable	application_readable	bool	null	no
libraries_name	name	string	null	no
libraries_version	version	string	null	no
vpc_access_connector_name	name	string	null	no
vpc_access_connector_egress_setting	egress_setting	string	null	no
noop_on_destroy	Whether to prevent service destroy.	bool	false	no