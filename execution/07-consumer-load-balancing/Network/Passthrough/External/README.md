# Consumer Networking - External Passthrough Network Load Balancing

## Overview

This Terraform solution is designed specifically for deploying and managing **External Passthrough Network Load Balancers** on Google Cloud. It provides a flexible way to define multiple NLBs using YAML configuration files, enabling a highly available and scalable architecture for distributing incoming TCP/UDP traffic across multiple instances.

External Passthrough Network Load Balancers are **regional, non-proxied load balancers** that forward packets directly to backend instances, preserving the client's source IP address. They operate at Layer 4 and are ideal for high-performance, lift-and-shift migrations, and applications built on common protocols like TCP, UDP, and SCTP.

* [Learn more about External Passthrough Network Load Balancing](https://cloud.google.com/load-balancing/docs/network)

## Configuration Method: YAML Files

This solution leverages YAML files to define the parameters for each individual Network Load Balancer instance you wish to create. This allows for separating infrastructure definition from specific load balancer configurations and managing multiple similar NLBs efficiently.

The Terraform code reads all YAML files (excluding those starting with `_`) from a specified directory (`var.config_folder_path`), parses them using the `yamldecode` function, applies default values from `variables.tf`, and transforms them into the data structure required by the `net-lb-ext` module.

## Pre-Requisites

Ensure the following are in place before deploying:

### Prior Step Completion:

Successful deployment typically assumes the completion of foundational infrastructure stages and the creation of backend resources:

* **01-organization:** Handles organization-level setup and **enables necessary Google Cloud APIs** for the project.
* **02-networking:** Establishes core network infrastructure like **VPC networks** and **subnetworks**. If using **IPv6**, the subnetwork must be specifically configured with an IPv6 access type (usually `EXTERNAL`).
    * [VPC Networks documentation](https://cloud.google.com/vpc/docs/vpc)
    * [Subnetworks documentation](https://cloud.google.com/vpc/docs/subnets)
    * [Configuring IPv6 for subnetworks](https://cloud.google.com/vpc/docs/using-ipv6)
* **Backend Instance Groups:** You **must** have backend instance groups (either [Managed Instance Groups (MIGs)](https://cloud.google.com/compute/docs/instance-groups/using-managed-groups) These groups are referenced by their name in the YAML configuration files.
* **Static External IP Addresses (Optional):** If you require stable IP addresses for your load balancers, you should **pre-allocate regional external IP addresses**. You will reference the **full self-links** of these address resources in your YAML configuration under the `address` field of forwarding rules. If no address self-link is specified, an ephemeral IP will be allocated.
    * [Reserving a static external IP address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address)

### Enabled APIs:

Ensure the following Google Cloud APIs are enabled in your project (usually handled by the organization stage):

* `compute.googleapis.com` (Compute Engine API)
    * [Enabling APIs](https://cloud.google.com/apis/docs/getting-started#enabling_apis)

### Permissions:

The Identity and Access Management (IAM) principal (user account or service account) executing Terraform must have sufficient permissions in the target project. The following pre-defined roles (or a custom role with equivalent permissions) are typically required:

* `roles/compute.loadBalancerAdmin`: For creating, updating, and deleting Load Balancing resources (Backend Services, Forwarding Rules, Health Checks)

Alternatively, you can create a custom role with the specific `get`, `list`, and `use` permissions for `compute.instanceGroups`, `compute.addresses`, and `compute.subnetworks`, combined with the full permissions from `roles/compute.loadBalancerAdmin`.

* [Understanding IAM roles and permissions](https://cloud.google.com/iam/docs/understanding-roles)
* [Compute Engine IAM roles](https://cloud.google.com/compute/docs/access/iam)

## Execution Steps

1.  **Configuration:**

    * Navigate to the root directory containing your Terraform configuration (`main.tf`, `variables.tf`, `locals.tf`, etc.).
    * Locate the directory specified by the `config_folder_path` variable (`default: "../../../../../configuration/consumer-load-balancing/Network/Passthrough/External/config/"`).
    * Create or modify YAML configuration files (e.g., `instance-lite.yaml`, `instance-expanded.yaml`) within this folder. Each YAML file should define the configuration for one load balancer instance based on your `locals.tf` flattening logic.
    * Define your desired Network Load Balancer configurations in these YAML files. Refer to the **Examples** section below and the `variables.tf` file for the expected structure and available options. **Ensure that required fields like `name`, `project_id`, `region`, and `backend.group_name` (or `backend.group` self-link) are correctly specified.**

2.  **Terraform Initialization:**

    * Open your terminal in the root directory of your Terraform configuration.
    * Initialize Terraform to download the required provider plugins and modules:

    ```bash
    terraform init
    ```
    * [Terraform `init` documentation](https://developer.hashicorp.com/terraform/cli/commands/init)

3.  **Review the Execution Plan:**

    * Generate and review an execution plan to understand what Terraform will create, modify, or destroy:

    ```bash
    terraform plan
    ```
    * Carefully review the planned actions, especially the creation of backend services and forwarding rules, verifying names, regions, and **backend group self-links** under the `group` attribute in the backend service's `backend` block.
    * [Terraform `plan` documentation](https://developer.hashicorp.com/terraform/cli/commands/plan)

4.  **Apply the Configuration:**

    * If the plan is satisfactory, apply the configuration to provision the resources:

    ```bash
    terraform apply
    ```
    * [Terraform `apply` documentation](https://developer.hashicorp.com/terraform/cli/commands/apply)

5.  **Monitor and Manage:**

    * After applying, monitor the status of your load balancers, including **backend health status**, through the [Google Cloud Console Load Balancing section](https://console.cloud.google.com/networking/loadbalancing/list).
    * Use `gcloud compute` commands to get details about your load balancers, backend services, and health checks.
        * [gcloud compute forwarding-rules reference](https://cloud.google.com/sdk/gcloud/reference/compute/forwarding-rules)
        * [gcloud compute backend-services reference](https://cloud.google.com/sdk/gcloud/reference/compute/backend-services)
        * [gcloud compute health-checks reference](https://cloud.google.com/sdk/gcloud/reference/compute/health-checks)
    * Update your load balancer configurations by modifying the corresponding YAML files in the `config_folder_path` and re-running `terraform apply`.

## Examples

The `locals.tf` file processes YAML files from the `config_folder_path`. Each YAML file is typically expected to define configuration for a single load balancer instance, including backend and frontend settings.

Below are sample YAML structures demonstrating common configurations. Refer to `variables.tf` for the full list of default values and configurable options.

* **Basic TCP NLB (Regional MIG using Default Region):**

    This example defines a simple TCP NLB targeting an existing Regional Managed Instance Group (R-MIG). Since `group_zone` is not specified and `group_region` is also omitted, the R-MIG is assumed to be in the same region as the load balancer (`us-central1`).

    ```yaml
    name: nlb-simple-web
    project_id: your-gcp-project-id
    region: us-central1
    backends:
    - group_name: your-regional-mig-in-lb-region # This MIG is expected to be a Regional MIG in us-central1
    ```

* **Basic TCP NLB with Multiple Regional MIGs (Default Region):**

    This example targets multiple Regional MIGs, all assumed to be in the same region as the load balancer.

    ```yaml
    name: multi-backend-lb
    project_id: your-gcp-project-id
    region: "us-central1"
    backends:
    - group_name: my-regional-mig-1
    - group_name: my-regional-mig-2
    ```

* **NLB with Mixed Regional and Zonal MIGs:**

    This example demonstrates configuring an NLB with both a Regional MIG (implicitly in the LB's region) and a Zonal MIG.

    ```yaml
    name: nlb-hybrid-backends
    project_id: your-gcp-project-id
    region: us-central1

    backends:
      - group_name: my-app-regional-mig
        group_region: us-central1 # Optional for Regional MIG, defaults to lb.region (us-central1)
        description: "This is a Regional MIG"

      - group_name: my-app-zonal-mig
        group_zone: us-central1-a
        description: This is a Zonal MIG in us-central1-a
        failover: true
    ```

* **TCP/UDP NLB with Static IPs and Advanced Options:**

    This example demonstrates defining multiple forwarding rules, using static addresses, and configuring more specific backend service and health check parameters. The backend shown is a Regional MIG with an explicitly defined region.

    ```yaml
    name: expanded-nlb
    project_id: <your-project-id>
    region: <region> # e.g., us-central1
    description: Comprehensive NLB configured via YAML

    backend_service:
      protocol: TCP
      port_name: my-app-port
      timeout_sec: 60
      connection_draining_timeout_sec: 300
      log_sample_rate: 0.75
      locality_lb_policy: MAGLEV
      session_affinity: CLIENT_IP_PORT_PROTO
      connection_tracking:
          persist_conn_on_unhealthy: ALWAYS_PERSIST
          track_per_session: false
      failover_config:
          disable_conn_drain: true
          drop_traffic_if_unhealthy: true
          ratio: 0.8

    backends:
    - group_name: <your-specific-regional-mig>
      group_region: <group_region>
      failover: true
      description: Main production backend group for this service

    health_check:
      check_interval_sec: 8
      timeout_sec: 8
      healthy_threshold: 4
      unhealthy_threshold: 4
      enable_logging: true
      description: Custom auto-created health check
      tcp:
          port: 80
          port_specification: USE_FIXED_PORT
          request: HEALTH_CHECK
          response: OK

    forwarding_rules:
      "fwd-rule-tcp":
          protocol: TCP
          ports: ["80", "443", "8080"]
          description: Primary web traffic listener (TCP)
          ipv6: false
    ```

<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_nlb_passthrough_ext"></a> [nlb\_passthrough\_ext](#module\_nlb\_passthrough\_ext) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/net-lb-ext | v39.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_backend_group_self_link_format"></a> [backend\_group\_self\_link\_format](#input\_backend\_group\_self\_link\_format) | Format string used to construct the backend instance group self\_link. | `string` | `"projects/%s/regions/%s/instanceGroups/%s"` | no |
| <a name="input_backend_item_description"></a> [backend\_item\_description](#input\_backend\_item\_description) | Default description for a backend item if not specified in YAML. | `string` | `"Terraform managed backend group association."` | no |
| <a name="input_backend_item_failover"></a> [backend\_item\_failover](#input\_backend\_item\_failover) | Default failover setting for a backend item if not specified in YAML. | `bool` | `false` | no |
| <a name="input_backend_log_sample_rate"></a> [backend\_log\_sample\_rate](#input\_backend\_log\_sample\_rate) | Default backend service log sample rate (0.0 to 1.0). | `number` | `1` | no |
| <a name="input_backend_port_name"></a> [backend\_port\_name](#input\_backend\_port\_name) | Default name of the port used by the backend service (for session affinity/connection tracking). | `string` | `"tcp"` | no |
| <a name="input_backend_protocol"></a> [backend\_protocol](#input\_backend\_protocol) | Default protocol used by the backend service. Common for NLB: TCP, UDP, SCTP. Default: TCP. | `string` | `"TCP"` | no |
| <a name="input_backend_service_connection_draining_timeout_sec"></a> [backend\_service\_connection\_draining\_timeout\_sec](#input\_backend\_service\_connection\_draining\_timeout\_sec) | Default time in seconds to wait for connections to terminate before removing a backend instance. Set to null to use the service default or not configure. | `number` | `null` | no |
| <a name="input_backend_service_connection_tracking"></a> [backend\_service\_connection\_tracking](#input\_backend\_service\_connection\_tracking) | Default connection tracking policy block. If set, this policy is applied if no 'connection\_tracking' block is present in the YAML. Set to null to not configure connection tracking by default. | <pre>object({<br>    idle_timeout_sec          = optional(number)<br>    persist_conn_on_unhealthy = optional(string)<br>    track_per_session         = optional(bool)<br>  })</pre> | `null` | no |
| <a name="input_backend_service_connection_tracking_idle_timeout_sec"></a> [backend\_service\_connection\_tracking\_idle\_timeout\_sec](#input\_backend\_service\_connection\_tracking\_idle\_timeout\_sec) | Default connection tracking idle timeout in seconds. | `number` | `null` | no |
| <a name="input_backend_service_connection_tracking_persist_conn_on_unhealthy"></a> [backend\_service\_connection\_tracking\_persist\_conn\_on\_unhealthy](#input\_backend\_service\_connection\_tracking\_persist\_conn\_on\_unhealthy) | Default behavior for persisting connections on unhealthy backends. Options: NEVER\_PERSIST, ALWAYS\_PERSIST, DEFAULT\_FOR\_PROTOCOL. | `string` | `null` | no |
| <a name="input_backend_service_connection_tracking_track_per_session"></a> [backend\_service\_connection\_tracking\_track\_per\_session](#input\_backend\_service\_connection\_tracking\_track\_per\_session) | Default flag to track connections per session. | `bool` | `null` | no |
| <a name="input_backend_service_failover_config"></a> [backend\_service\_failover\_config](#input\_backend\_service\_failover\_config) | Default failover policy block. If set, this policy is applied if no 'failover\_config' block is present in the YAML. Set to null to not configure failover by default. | <pre>object({<br>    disable_conn_drain        = optional(bool)<br>    drop_traffic_if_unhealthy = optional(bool)<br>    ratio                     = optional(number)<br>  })</pre> | `null` | no |
| <a name="input_backend_service_failover_disable_conn_drain"></a> [backend\_service\_failover\_disable\_conn\_drain](#input\_backend\_service\_failover\_disable\_conn\_drain) | Default flag to disable connection draining on failover. | `bool` | `null` | no |
| <a name="input_backend_service_failover_drop_traffic_if_unhealthy"></a> [backend\_service\_failover\_drop\_traffic\_if\_unhealthy](#input\_backend\_service\_failover\_drop\_traffic\_if\_unhealthy) | Default flag to drop traffic if all backends are unhealthy. | `bool` | `null` | no |
| <a name="input_backend_service_failover_ratio"></a> [backend\_service\_failover\_ratio](#input\_backend\_service\_failover\_ratio) | Default failover ratio for the backend service. | `number` | `null` | no |
| <a name="input_backend_service_locality_lb_policy"></a> [backend\_service\_locality\_lb\_policy](#input\_backend\_service\_locality\_lb\_policy) | Default locality load balancing policy for the backend service. Options: MAGLEV, WEIGHTED\_MAGLEV. | `string` | `"MAGLEV"` | no |
| <a name="input_backend_service_session_affinity"></a> [backend\_service\_session\_affinity](#input\_backend\_service\_session\_affinity) | Default session affinity for the backend service. | `string` | `"NONE"` | no |
| <a name="input_backend_timeout_sec"></a> [backend\_timeout\_sec](#input\_backend\_timeout\_sec) | Default timeout in seconds for backend connections. | `number` | `10` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding NLB configuration values. | `string` | `"../../../../../configuration/consumer-load-balancing/Network/Passthrough/External/config/"` | no |
| <a name="input_description"></a> [description](#input\_description) | Optional description used for resources. | `string` | `"Terraform managed External Passthrough Network Load Balancer."` | no |
| <a name="input_forwarding_rule_address"></a> [forwarding\_rule\_address](#input\_forwarding\_rule\_address) | Default IP address (name or self\_link) for the forwarding rule. Set to null for an ephemeral IP. | `string` | `null` | no |
| <a name="input_forwarding_rule_description"></a> [forwarding\_rule\_description](#input\_forwarding\_rule\_description) | Default description for forwarding rules if not specified in YAML. | `string` | `null` | no |
| <a name="input_forwarding_rule_ipv6"></a> [forwarding\_rule\_ipv6](#input\_forwarding\_rule\_ipv6) | Default setting for enabling IPv6 on forwarding rules if not specified in YAML. | `bool` | `false` | no |
| <a name="input_forwarding_rule_ipv6_fallback"></a> [forwarding\_rule\_ipv6\_fallback](#input\_forwarding\_rule\_ipv6\_fallback) | Fallback default value for forwarding rule ipv6 attribute if not specified in YAML or var.forwarding\_rule\_ipv6. | `bool` | `false` | no |
| <a name="input_forwarding_rule_name_override"></a> [forwarding\_rule\_name\_override](#input\_forwarding\_rule\_name\_override) | Default name override for forwarding rules if not specified in YAML. Use with caution, module usually handles naming. | `string` | `null` | no |
| <a name="input_forwarding_rule_ports"></a> [forwarding\_rule\_ports](#input\_forwarding\_rule\_ports) | Default list of ports for the forwarding rule. Set to null to listen on all ports (only for TCP/UDP). | `list(string)` | `null` | no |
| <a name="input_forwarding_rule_protocol"></a> [forwarding\_rule\_protocol](#input\_forwarding\_rule\_protocol) | Default protocol for the forwarding rule (listener). Common for NLB: TCP, UDP, SCTP. Default: TCP. | `string` | `"TCP"` | no |
| <a name="input_forwarding_rule_subnetwork"></a> [forwarding\_rule\_subnetwork](#input\_forwarding\_rule\_subnetwork) | Default subnetwork for forwarding rules if not specified in YAML. Required for IPv6. | `string` | `null` | no |
| <a name="input_forwarding_rules_map"></a> [forwarding\_rules\_map](#input\_forwarding\_rules\_map) | Default map of forwarding rules if none are specified in YAML. Defaults to a single rule with key ''. | <pre>map(object({<br>    address     = optional(string)<br>    description = optional(string)<br>    ipv6        = optional(bool)<br>    name        = optional(string)<br>    ports       = optional(list(string))<br>    protocol    = optional(string)<br>    subnetwork  = optional(string)<br>  }))</pre> | <pre>{<br>  "": {}<br>}</pre> | no |
| <a name="input_hc_grpc_config"></a> [hc\_grpc\_config](#input\_hc\_grpc\_config) | Default value for the health check's GRPC configuration if not specified. Typically null. | `any` | `null` | no |
| <a name="input_hc_http2_config"></a> [hc\_http2\_config](#input\_hc\_http2\_config) | Default value for the health check's HTTP2 configuration if not specified. Typically null. | `any` | `null` | no |
| <a name="input_hc_http_config"></a> [hc\_http\_config](#input\_hc\_http\_config) | Default value for the health check's HTTP configuration if not specified. Typically null. | `any` | `null` | no |
| <a name="input_hc_https_config"></a> [hc\_https\_config](#input\_hc\_https\_config) | Default value for the health check's HTTPS configuration if not specified. Typically null. | `any` | `null` | no |
| <a name="input_hc_ssl_config"></a> [hc\_ssl\_config](#input\_hc\_ssl\_config) | Default value for the health check's SSL configuration if not specified. Typically null. | `any` | `null` | no |
| <a name="input_hc_tcp_config"></a> [hc\_tcp\_config](#input\_hc\_tcp\_config) | Default value for the health check's TCP configuration if not specified (and not applying the global default TCP). Typically null. | `any` | `null` | no |
| <a name="input_health_check_check_interval_sec"></a> [health\_check\_check\_interval\_sec](#input\_health\_check\_check\_interval\_sec) | Default health check interval in seconds. | `number` | `5` | no |
| <a name="input_health_check_config"></a> [health\_check\_config](#input\_health\_check\_config) | Default value for the entire health\_check block if not provided in the YAML. Should be null to indicate absence unless a specific default structure is desired. | `any` | `null` | no |
| <a name="input_health_check_enable_logging"></a> [health\_check\_enable\_logging](#input\_health\_check\_enable\_logging) | Default flag to enable logging for the health check. | `bool` | `false` | no |
| <a name="input_health_check_healthy_threshold"></a> [health\_check\_healthy\_threshold](#input\_health\_check\_healthy\_threshold) | Default number of consecutive successful health checks required for a backend to be considered healthy. | `number` | `2` | no |
| <a name="input_health_check_name_override"></a> [health\_check\_name\_override](#input\_health\_check\_name\_override) | Default value for health\_check.name if an existing health check name is not provided in the YAML. Typically null to let the module auto-create or use other logic. | `string` | `null` | no |
| <a name="input_health_check_tcp_port"></a> [health\_check\_tcp\_port](#input\_health\_check\_tcp\_port) | Default port for auto-created TCP health checks. | `number` | `null` | no |
| <a name="input_health_check_tcp_port_spec_fallback"></a> [health\_check\_tcp\_port\_spec\_fallback](#input\_health\_check\_tcp\_port\_spec\_fallback) | Fallback default value for auto-created TCP health check port\_specification if not specified in YAML or var.health\_check\_tcp\_port\_specification. | `string` | `"USE_SERVING_PORT"` | no |
| <a name="input_health_check_tcp_port_specification"></a> [health\_check\_tcp\_port\_specification](#input\_health\_check\_tcp\_port\_specification) | Default port specification for auto-created TCP health checks (USE\_FIXED\_PORT, USE\_NAMED\_PORT, USE\_SERVING\_PORT). | `string` | `"USE_SERVING_PORT"` | no |
| <a name="input_health_check_tcp_proxy_header"></a> [health\_check\_tcp\_proxy\_header](#input\_health\_check\_tcp\_proxy\_header) | Default proxy header for auto-created TCP health checks (NONE, PROXY\_V1). | `string` | `null` | no |
| <a name="input_health_check_tcp_request"></a> [health\_check\_tcp\_request](#input\_health\_check\_tcp\_request) | Default request string for auto-created TCP health checks. | `string` | `null` | no |
| <a name="input_health_check_tcp_response"></a> [health\_check\_tcp\_response](#input\_health\_check\_tcp\_response) | Default expected response string for auto-created TCP health checks. | `string` | `null` | no |
| <a name="input_health_check_timeout_sec"></a> [health\_check\_timeout\_sec](#input\_health\_check\_timeout\_sec) | Default health check timeout in seconds. | `number` | `5` | no |
| <a name="input_health_check_unhealthy_threshold"></a> [health\_check\_unhealthy\_threshold](#input\_health\_check\_unhealthy\_threshold) | Default number of consecutive failed health checks required for a backend to be considered unhealthy. | `number` | `2` | no |
| <a name="input_labels"></a> [labels](#input\_labels) | Labels set on resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_nlb_backend_service_self_links"></a> [nlb\_backend\_service\_self\_links](#output\_nlb\_backend\_service\_self\_links) | Map of Network Load Balancer names to their backend service self links. |
| <a name="output_nlb_backend_services"></a> [nlb\_backend\_services](#output\_nlb\_backend\_services) | Detailed backend service resources for each NLB. |
| <a name="output_nlb_forwarding_rule_addresses"></a> [nlb\_forwarding\_rule\_addresses](#output\_nlb\_forwarding\_rule\_addresses) | Map of Network Load Balancer names to their forwarding rule names and IP addresses. |
| <a name="output_nlb_forwarding_rule_self_links"></a> [nlb\_forwarding\_rule\_self\_links](#output\_nlb\_forwarding\_rule\_self\_links) | Map of Network Load Balancer names to their forwarding rule names and self links. |
| <a name="output_nlb_forwarding_rules"></a> [nlb\_forwarding\_rules](#output\_nlb\_forwarding\_rules) | Detailed forwarding rule resources for each NLB. |
| <a name="output_nlb_health_check_self_links"></a> [nlb\_health\_check\_self\_links](#output\_nlb\_health\_check\_self\_links) | Map of Network Load Balancer names to their auto-created health check self links (if applicable). |
| <a name="output_nlb_health_checks"></a> [nlb\_health\_checks](#output\_nlb\_health\_checks) | Detailed auto-created health check resources for each NLB (if applicable). |

<!-- END_TF_DOCS -->