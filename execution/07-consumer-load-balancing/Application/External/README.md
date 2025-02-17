## Consumer Networking - External Application Load Balancing

### Overview

This Terraform solution is designed specifically for deploying and managing **External Application Load Balancers** on Google Cloud Platform (GCP) using YAML configuration files. The solution focuses on creating a highly available and scalable load balancing architecture that efficiently distributes incoming application traffic across multiple instances.

### Pre-Requisites

#### Prior Step Completion:

- **Completed Prior Stages:** Successful deployment of the application load balancer requires the completion of the following stages:

    - **01-organization:** This stage handles the activation of required Google Cloud APIs for Load Balancing.
    - **02-networking:** This stage sets up the necessary network infrastructure, such as VPCs and subnets.
    - **06-consumer:** This stage involves creating Managed Instance Groups (MIGs), which serve as the backend for the load balancer.

#### Enabled APIs:

Ensure the following Google Cloud APIs are enabled in your project:

- Compute Engine API
- Cloud Load Balancing API

#### Permissions:

The user or service account executing Terraform must have the following roles (or equivalent permissions):

- Load Balancer Admin (for managing load balancers)
- **Optional** : Compute Admin (for managing VMs and instance groups)

### Execution Steps

1. **Configuration:**

    - Create YAML configuration files (e.g., lb_instance1.yaml, lb_instance2.yaml) in the specified configuration directory.
    - Edit the YAML files to define the desired application load balancer configurations. (See **Examples** below)

2. **Terraform Initialization:**

    - Open your terminal and navigate to the directory containing the Terraform configuration.
    - Run the following command to initialize Terraform:

    ```bash
    terraform init
    ```

3. **Review the Execution Plan:**

    - Generate an execution plan with the following command to review changes Terraform will make to your Google Cloud infrastructure:

    ```bash
    terraform plan -var-file=../../../configuration/consumer-load-balancing/Application/External/lb.tfvars
    ```

4. **Apply the Configuration:**

    Once satisfied with the plan, execute the terraform apply command to provision your application load balancer:

    ```bash
    terraform apply -var-file=../../../configuration/consumer-load-balancing/Application/External/lb.tfvars
    ```

5. **Monitor and Manage:**

    * After creating the load balancer, monitor its status and performance through the Google Cloud Console or using Google Cloud CLI.
    
    * Use Terraform to manage updates and changes to your load balancer as needed.

### Examples

- **lb_instance1.yaml:** This sample YAML file defines a configuration for an External Application Load Balancer with custom health checks.

  ```yaml
    name: load-balancer-custom-hc
    project: <project-id>
    network: default
    backends:
    default:
        protocol: "HTTP"
        port: 80
        port_name: "http"
        timeout_sec: 30
        enable_cdn: false
        health_check:
        request_path: "/healthz"
        port: 80
        log_config:
        enable: true
        sample_rate: 0.5
        groups:
        - group: instance-group
            region : us-central1
  ```

- **lb_instance2.yaml:** Another sample YAML file using a default health check.

  ```yaml
    name: load-balancer-default-hc
    project: <project-id>
    network: default
    backends:
    default:
        groups:
        - group: instance-group
            region: us-central1
  ```

### Important Notes

- The solution assumes that Managed Instance Groups (MIGs) have been created in your project as part of previous steps.
- Customize YAML configuration files according to specific requirements (e.g., project ID, target tags, firewall networks).
- Ensure correct service account credentials are provided for Terraform to interact with your Google Cloud project.
- Refer to `variables.tf` for a complete list of available variables and their descriptions.
- The Terraform module used in this solution may have additional configuration options; refer to its documentation for further customization.
- Remember to replace placeholders (e.g., `<project-id>`, `<target-tag>`, `<firewall-network>`, `<instance-group-name>`) with actual values in YAML files and Terraform configurations.

<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_lb_http"></a> [lb_http](#module\_lb_http) | ../../../../modules/lb_http/ | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_backend_port"></a> [backend\_port](#input\_backend\_port) | Port used by the backend service. | `number` | `80` | no |
| <a name="input_backend_port_name"></a> [backend\_port\_name](#input\_backend\_port\_name) | Name of the port used by the backend service. | `string` | `"http"` | no |
| <a name="input_backend_protocol"></a> [backend\_protocol](#input\_backend\_protocol) | Protocol used by the backend service. | `string` | `"HTTP"` | no |
| <a name="input_backend_timeout_sec"></a> [backend\_timeout\_sec](#input\_backend\_timeout\_sec) | Timeout in seconds for backend requests. | `number` | `10` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding LB configuration values. | `string` | `"../../../../configuration/consumer-load-balancing/Application/External/config/"` | no |
| <a name="input_enable_cdn"></a> [enable\_cdn](#input\_enable\_cdn) | Enable CDN for the backend service. | `bool` | `false` | no |
| <a name="input_health_check"></a> [health\_check](#input\_health\_check) | Health check configuration for the load balancer. | <pre>object({<br>    request_path = string<br>    port         = number<br>  })</pre> | <pre>{<br>  "port": 80,<br>  "request_path": "/"<br>}</pre> | no |
| <a name="input_iap_config"></a> [iap\_config](#input\_iap\_config) | IAP (Identity-Aware Proxy) configuration for the load balancer. | <pre>object({<br>    enable = bool<br>  })</pre> | <pre>{<br>  "enable": false<br>}</pre> | no |
| <a name="input_instance_group"></a> [instance\_group](#input\_instance\_group) | Instance group for the Load Balancer | `string` | `null` | no |
| <a name="input_log_config"></a> [log\_config](#input\_log\_config) | Log configuration for the load balancer. | <pre>object({<br>    enable      = bool<br>    sample_rate = number<br>  })</pre> | <pre>{<br>  "enable": true,<br>  "sample_rate": 1<br>}</pre> | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_load_balancer_ips"></a> [load\_balancer\_ips](#output\_load\_balancer\_ips) | A map of load balancer names to their external IP addresses. |
| <a name="output_load_balancer_ipv6s"></a> [load\_balancer\_ipv6s](#output\_load\_balancer\_ipv6s) | A map of load balancer names to their IPv6 addresses, if enabled; else "undefined". |
| <a name="output_load_balancers"></a> [load\_balancers](#output\_load\_balancers) | Detailed information about each load balancer. |

<!-- END_TF_DOCS -->