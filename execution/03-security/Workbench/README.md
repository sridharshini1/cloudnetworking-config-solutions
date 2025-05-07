# Secure Firewall Rules for Google Cloud Vertex AI Workbench

## Overview

This Terraform module simplifies the process of creating and managing essential firewall rules for secure access to Google Cloud Vertex AI Workbench instances. Properly configured firewall rules are critical for protecting your instances while allowing necessary connectivity.

## Key Features

- **Firewall Rule Automation:** Effortlessly define and deploy firewall rules specific to the access needs of your Workbench instances.
- **Secure Access Configuration:** Define rules for essential access like SSH, including options for direct connections from trusted IPs.
- **Flexible Configuration:**
  - Control allowed source IP addresses or ranges (`source_ranges`).
  - Apply rules precisely to specific instances using target tags (`targets`).
- **Integration with Existing Networks:** Works seamlessly with your existing GCP VPC networks.
- **Google Provider Integration:** Leverages the official Terraform Google provider for reliability.

## Prerequisites

Before using this module, ensure the following prerequisites are met:

1. **Google Cloud Project:** You must have a Google Cloud project with billing enabled.
2. **Terraform Installed:** Install Terraform (v1.3.0 or later) on your local machine or CI/CD environment.
3. **Google Cloud SDK:** Install and authenticate the Google Cloud SDK (`gcloud`) to manage your project and resources.
4. **IAM Permissions:** Ensure you have the following IAM roles:
   - `roles/compute.securityAdmin` for managing firewall rules.
   - `roles/iam.serviceAccountUser` if using service accounts for Terraform.
5. **VPC Network:** A pre-existing VPC network with private google access enabled where the firewall rules will be applied.
6. **Terraform Google Provider:** Configure the Terraform Google provider with appropriate credentials and project settings.

By meeting these prerequisites, you can ensure a smooth setup and deployment process for securing your Vertex AI Workbench instances.

## Description

- **Firewall Rules:** Define how network traffic is filtered for your Workbench environments. This module focuses on creating firewall rules that allow necessary traffic (like SSH) while maintaining security best practices.
- **Secure Access Patterns:** Standard Vertex AI Workbench access involves SSH (port 22) for terminal access via private IPs. This module focuses solely on secure SSH access configurations for Workbench instances, ensuring private and restricted connectivity.
- **Source Ranges:** Specify the **trusted** IP addresses or CIDR ranges permitted to initiate SSH connections. For direct SSH access, this should be *your specific* workstation or corporate IP(s). For example, use `"203.0.113.5/32"` for a single trusted IP or `"192.168.1.0/24"` for a range. For IAP access, use Google's designated range (`35.235.240.0/20`). **Avoid using overly broad ranges like `0.0.0.0/0` to maintain security.** Always follow the principle of least privilege.
- **Target Tags:** Assign a specific network tag (e.g., `workbench-instance`) to your Workbench VMs during their creation or modification. This ensures the firewall rules are applied only to the intended instances. Reference this tag in the rule's `targets` attribute to enforce the rule on the correct VMs.

## Example Firewall Rules Configuration (`terraform.tfvars`)

The following example illustrates how to configure ingress rules in your `terraform.tfvars` file to allow secure SSH access to Workbench instances tagged with `workbench-instance`.

```hcl
# terraform.tfvars

# Project ID for the Google Cloud project
project_id = "<YOUR_PROJECT_ID>" # <-- Replace

# Network name (or self-link) where the firewall rules will be applied
network    = "<YOUR_VPC_NETWORK_SELF_LINK_OR_NAME>" # <-- Replace

# Egress rules (typically left empty to use default allow-all)
egress_rules = {}

# Ingress rules configuration for Vertex AI Workbench
ingress_rules = {

  # Rule 1: Allow SSH from specific trusted IP addresses/ranges.
  # IMPORTANT: Replace the placeholder source_ranges with your actual public IP(s). Avoid 0.0.0.0/0.
  "fw-allow-ssh-workbench-trusted" = {
    description      = "Allow SSH access to Workbench instances from trusted sources"
    priority         = 1000
    source_ranges    = [
      "YOUR_IP/32", # <--- *** REPLACE with your specific IP address/range (e.g., "203.0.113.5/32") ***
      # Add more trusted IPs/ranges if needed, e.g., "YOUR_VPN_RANGE/24"
    ]
    targets          = ["workbench-instance"] # Apply this tag to your Workbench VMs
    enable_logging = {
      include_metadata = true
    }
    rules = [
      {
        protocol = "tcp"
        ports    = ["22"] # Standard SSH port
      }
    ]
  }
}
```

<!-- BEGIN_TF_DOCS -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_workbench_firewall"></a> [workbench\_firewall](#module\_workbench\_firewall) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/net-vpc-firewall | v30.0.0 |

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
| <a name="output_workbench_firewall_rules"></a> [workbench\_firewall\_rules](#output\_workbench\_firewall\_rules) | Map of firewall rules created. |
<!-- END_TF_DOCS -->
