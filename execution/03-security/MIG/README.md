# Overview

This Terraform module simplifies the process of creating and managing firewall rules for Google Cloud Managed Instance Groups (MIGs). Properly configured firewall rules are essential to ensure secure communication between your MIG instances and external services, including load balancers and health checks.

## Key Features

- **Firewall Rule Automation:** Effortlessly define and deploy firewall rules specific to the needs of your Managed Instance Groups.
- **Health Check Support:** Automatically allows health check traffic from Google Cloud's designated IP ranges to ensure the availability of your instances.
- **Flexible Configuration:**
  - Control allowed source IP addresses or ranges (source_ranges).
  - Apply rules to specific instances using target tags (target_tags).
- **Integration with Existing Networks:** Works seamlessly with your existing GCP networks.
- **Google Provider Integration:** Leverages the official Terraform Google provider for reliability.

## Description

- **Firewall Rules:** Define how network traffic is filtered for your Managed Instance Groups. This module focuses on creating firewall rules that allow necessary traffic while maintaining security best practices.
- **Health Checks:** The module specifically allows traffic from Google Cloud's health check IP ranges, ensuring that health checks can successfully reach your instances.
- **Source Ranges:** Specify the IP addresses or ranges permitted to initiate connections to your instances.
- **Target Tags:** Apply the firewall rule to specific MIG instances by tagging them with a designated label (e.g., "allow-health-checks").

## Example Firewall Rules

### Ingress Rules Configuration

The following example illustrates how to configure ingress rules for a Managed Instance Group:

```hcl
# Project ID for the Google Cloud project
project_id = "<project-id>"

# Network name where the firewall rules will be applied
network    = "projects/<project-id>/global/networks/<vpc-name>"

# Ingress rules configuration
ingress_rules = {
  "fw-allow-health-checks" = {
    deny               = false
    description        = "Allow health checks"
    destination_ranges = []
    disabled           = false
    enable_logging     = {
      include_metadata = true
    }
    priority           = 1000
    source_ranges      = [
      "130.211.0.0/22",
      "35.191.0.0/16"
    ]
    targets            = ["allow-health-checks"]
    rules              = [
      {
        protocol = "tcp"
        ports    = ["80"]
      }
    ]
  }
}
```
<!-- BEGIN_TF_DOCS -->

### Outputs

The module provides an output `firewall_rules` listing the details of the created firewall rules for easier reference and management.

## Security Best Practices

- **Least Privilege:** Restrict `source_ranges` to only the IP addresses that genuinely require access to your MIG instances.
- **Regular Audits:** Periodically review your firewall rules to ensure they align with your security requirements and adjust as necessary.
- **Use Tags Wisely:** Apply target tags consistently across your instance templates and firewall rules to ensure proper application of security policies.
- **Monitor Logs:** Enable logging for firewall rules to track access patterns and identify potential security issues.

This document provides a clear framework for implementing and managing firewall rules for Google Cloud Managed Instance Groups, ensuring secure and efficient operation of your cloud infrastructure.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_ssh_firewall"></a> [ssh\_firewall](#module\_ssh\_firewall) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/net-vpc-firewall | v30.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_default_rules_config"></a> [default\_rules\_config](#input\_default\_rules\_config) | Optionally created convenience rules. Set the 'disabled' attribute to true, or individual rule attributes to empty lists to disable. | <pre>object({<br>    admin_ranges = optional(list(string))<br>    disabled     = optional(bool, true)<br>    http_ranges = optional(list(string), [<br>      "35.191.0.0/16", "130.211.0.0/22", "209.85.152.0/22", "209.85.204.0/22"]<br>    )<br>    http_tags = optional(list(string), ["http-server"])<br>    https_ranges = optional(list(string), [<br>      "35.191.0.0/16", "130.211.0.0/22", "209.85.152.0/22", "209.85.204.0/22"]<br>    )<br>    https_tags = optional(list(string), ["https-server"])<br>    ssh_ranges = optional(list(string), ["35.235.240.0/20"])<br>    ssh_tags   = optional(list(string), ["ssh"])<br>  })</pre> | <pre>{<br>  "disabled": true<br>}</pre> | no |
| <a name="input_egress_rules"></a> [egress\_rules](#input\_egress\_rules) | List of egress rule definitions, default to deny action. Null destination ranges will be replaced with 0/0. | <pre>map(object({<br>    deny               = optional(bool, true)<br>    description        = optional(string)<br>    destination_ranges = optional(list(string))<br>    disabled           = optional(bool, false)<br>    enable_logging = optional(object({<br>      include_metadata = optional(bool)<br>    }))<br>    priority             = optional(number, 1000)<br>    source_ranges        = optional(list(string))<br>    targets              = optional(list(string))<br>    use_service_accounts = optional(bool, false)<br>    rules = optional(list(object({<br>      protocol = string<br>      ports    = optional(list(string))<br>    })), [{ protocol = "all" }])<br>  }))</pre> | `{}` | no |
| <a name="input_ingress_rules"></a> [ingress\_rules](#input\_ingress\_rules) | List of ingress rule definitions, default to allow action. Null source ranges will be replaced with 0/0. | <pre>map(object({<br>    deny               = optional(bool, false)<br>    description        = optional(string)<br>    destination_ranges = optional(list(string), []) # empty list is needed as default to allow deletion after initial creation with a value. See https://github.com/hashicorp/terraform-provider-google/issues/14270<br>    disabled           = optional(bool, false)<br>    enable_logging = optional(object({<br>      include_metadata = optional(bool)<br>    }))<br>    priority             = optional(number, 1000)<br>    source_ranges        = optional(list(string))<br>    sources              = optional(list(string))<br>    targets              = optional(list(string))<br>    use_service_accounts = optional(bool, false)<br>    rules = optional(list(object({<br>      protocol = string<br>      ports    = optional(list(string))<br>    })), [{ protocol = "all" }])<br>  }))</pre> | `{}` | no |
| <a name="input_network"></a> [network](#input\_network) | The name (or self-link) of the network to create the firewall rule in. | `string` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The ID of the Google Cloud project. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_rules"></a> [rules](#output\_rules) | Map of firewall rules created. |
<!-- END_TF_DOCS -->