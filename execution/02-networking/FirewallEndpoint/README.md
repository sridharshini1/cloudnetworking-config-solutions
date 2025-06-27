## Overview

This Terraform configuration simplifies the process of creating and managing Google Cloud Firewall Endpoints and their associations with VPC networks. Using a modular, data-driven approach, it reads declarative YAML files to deploy regional security resources for traffic inspection across your organization.

## Key Features

- **YAML-Driven Automation:** Effortlessly define and deploy firewall endpoints and their associations using simple, readable YAML files.
- **Flexible Resource Creation:** Create an organization-level firewall endpoint, a project-level association, or both simultaneously to link them to a VPC network.
- **Clear Separation of Concerns:** The model respects GCP's structure by defining the endpoint at the organization level and the association at the project level, all within a single, cohesive configuration file.
- **Centralized Management:** Manage firewall endpoint configurations for your entire GCP organization from a single source-controlled repository.
- **Integration with Custom Module:** Leverages the custom Terraform module we built for reliable and consistent resource deployment.

## Prerequisites

Before using this configuration, ensure the following prerequisites are met:

1.  **Google Cloud Organization and Project:** You must have a GCP Organization (for the endpoint) and a GCP Project containing a VPC network (for the association).
2.  **Terraform Installed:** Install Terraform (v1.3.0 or later) on your local machine or CI/CD environment.
3.  **Google Cloud SDK:** Install and authenticate the Google Cloud SDK (`gcloud`) to manage your project and resources.
4.  **IAM Permissions:** Ensure the principal (user, service account) running Terraform has the **Compute Network Admin** (`roles/compute.networkAdmin`) role at the **organization level**.
5.  **VPC Network:** A pre-existing VPC network where the endpoint association will be made.
6.  **Terraform Firewall Endpoint Module:** The custom module for firewall endpoints must be available at the path specified in `firewallendpoint.tf` (e.g., `../../../modules/firewall_endpoint/`).
7.  **Terraform Google Provider:** Configure the Terraform Google provider with appropriate credentials and project settings (typically in a `providers.tf` file) to handle authentication and billing.

## Description

- **Firewall Endpoints:** An organization-level, zonal resource that represents a point where traffic can be intercepted for inspection by a firewall. It must be created within a specific zone.
- **Firewall Endpoint Associations:** A project-level resource that links a Firewall Endpoint to a specific VPC network within a project. This activates the endpoint for that network.
- **YAML Configuration:** This setup works by reading all `.yaml` files from a specified directory (e.g., `config/`). Each YAML file declaratively defines the endpoint and/or association you want to create. This allows for a clear, auditable trail of your network security infrastructure.
- **Key YAML Blocks:**
    - `location`: A mandatory top-level key specifying the zone (e.g., `us-central1-a`) where the resources will be deployed.
    - `firewall_endpoint`: This block contains the data for the `google_network_security_firewall_endpoint` resource, including its `name`, `organization_id`, and the `billing_project_id`.
    - `firewall_endpoint_association`: This block contains the data for the `google_network_security_firewall_endpoint_association` resource, including the `association_project_id` and the `vpc_id`.

## Example YAML Configuration (`prod_fw_endpoint.yaml`)

The following example illustrates how to define a Firewall Endpoint and an Association in the `us-central1-a` zone and link them together. Place this file inside your configuration directory.

```yaml
# config/prod_fw_endpoint.yaml

# The location (zone) is mandatory for these resources.
location: "us-central1-a"

# --- Defines the organization-level Firewall Endpoint ---
firewall_endpoint:
  create: true
  name: "prod-fw-endpoint-us-central1-a"
  organization_id: "YOUR_ORGANIZATION_ID"         # <-- Replace
  billing_project_id: "my-central-billing-project"  # <-- Replace with project to bill
  labels:
    env: "prod"
    region: "us-central1"

# --- Defines the project-level Association ---
firewall_endpoint_association:
  create: true
  name: "assoc-to-prod-vpc-us-central1"
  association_project_id: "my-production-project"   # <-- Replace with project containing VPC
  vpc_id: "projects/my-production-project/global/networks/prod-vpc-us-east"
  labels:
    env: "prod"
```
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_firewall_endpoints"></a> [firewall\_endpoints](#module\_firewall\_endpoints) | ../../../modules/firewall_endpoint | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_association_disabled"></a> [association\_disabled](#input\_association\_disabled) | Disabled State for an association. | `bool` | `false` | no |
| <a name="input_association_labels"></a> [association\_labels](#input\_association\_labels) | Labels for an association. | `map(string)` | `{}` | no |
| <a name="input_association_name"></a> [association\_name](#input\_association\_name) | Name for an association for firewall endpoint. | `string` | `null` | no |
| <a name="input_association_project_id"></a> [association\_project\_id](#input\_association\_project\_id) | Project ID for the firewall endpoint association. | `string` | `null` | no |
| <a name="input_billing_project_id"></a> [billing\_project\_id](#input\_billing\_project\_id) | Project id to be billed for the resources not deployed in any specific project | `string` | `null` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Path to the folder containing the YAML configuration files for firewall endpoints. | `string` | `"../../../configuration/networking/FirewallEndpoint/config/"` | no |
| <a name="input_create_firewall_endpoint"></a> [create\_firewall\_endpoint](#input\_create\_firewall\_endpoint) | Control condition to create firewall endpoint. | `bool` | `false` | no |
| <a name="input_create_firewall_endpoint_association"></a> [create\_firewall\_endpoint\_association](#input\_create\_firewall\_endpoint\_association) | Control variable for creating an association for firewall endpoint. | `bool` | `false` | no |
| <a name="input_existing_firewall_endpoint_id"></a> [existing\_firewall\_endpoint\_id](#input\_existing\_firewall\_endpoint\_id) | Existing firewall endpoint ID to be linked to the association. | `string` | `null` | no |
| <a name="input_firewall_endpoint_labels"></a> [firewall\_endpoint\_labels](#input\_firewall\_endpoint\_labels) | Labels for a firewall endpoint. | `map(string)` | `{}` | no |
| <a name="input_firewall_endpoint_name"></a> [firewall\_endpoint\_name](#input\_firewall\_endpoint\_name) | Firewall endpoint name | `string` | `null` | no |
| <a name="input_location"></a> [location](#input\_location) | Location (zone) for the endpoint to be deployed. | `string` | `null` | no |
| <a name="input_organization_id"></a> [organization\_id](#input\_organization\_id) | Organization id to be used to deploy the resources. | `string` | `null` | no |
| <a name="input_tls_inspection_policy_id"></a> [tls\_inspection\_policy\_id](#input\_tls\_inspection\_policy\_id) | TLS Inspection Policy name. | `string` | `null` | no |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | VPC network name for the firewall endpoint association. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_firewall_endpoint_associations"></a> [firewall\_endpoint\_associations](#output\_firewall\_endpoint\_associations) | A map of all created firewall endpoint associations with their details. |
| <a name="output_firewall_endpoints"></a> [firewall\_endpoints](#output\_firewall\_endpoints) | A map of all created firewall endpoints with their details. |
<!-- END_TF_DOCS -->