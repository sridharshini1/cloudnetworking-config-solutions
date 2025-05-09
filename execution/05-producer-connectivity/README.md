# Manual Networking Setup for PSC

**NOTE** : Please skip this step if you are not provisioning a producer service using [Private Service Connect](https://cloud.google.com/vpc/docs/private-service-connect).

## Overview

This stage establishes a Private Service Connect (PSC) connection between your consumer and the producer service you created in the previous "04-producer" step. This is done by creating a forwarding rule that directs traffic from a reserved IP address to the PSC attachment on your producer service (e.g. Cloud SQL database or AlloyDB instance/cluster). PSC enables secure and private communication within Google Cloud Platform (GCP), shielding your services from the public internet. This stage configures a Private Service Connect (PSC) connection between your consumer project and the producer services you've set up. It does this by creating:

1. **Internal IP Addresses:** Reserved within specific subnetworks in your consumer project, acting as private endpoints for the PSC connection.
2. **Forwarding Rules:** Directing traffic destined for these internal IPs to your producer services through the PSC connection.

## How It Works

This configuration is designed for flexibility. You can define multiple producer services within your `producer-connectivity.tfvars` file in the `configuration/` folder.

**Key Points:**

* **`google_compute_address`:**  This resource will create a new internal IP address if `ip_address_literal` is specified for a `psc_endpoint` in your `producer-connectivity.tfvars` file. If no address is provided, it will automatically reserve an address.
* **`google_compute_forwarding_rule`:**  This resource sets up the forwarding rule, connecting the internal IP (or the automatically created one) to the `psc_service_attachment_link` of your producer service.

## Configuration

This stage uses a modularized approach. The `psc-forwarding-rules.tf` file in the root directory orchestrates the creation of multiple forwarding rules based on the configuration provided in the `producer-connectivity.tfvars` file in the `configuration/` folder.

While running this stage, please carefully note the following details:

- The variable `psc_endpoints` is a list of objects, where each object represents a connection to a producer service:

  - `endpoint_project_id` = "your-consumer-project-id"  # Project where the forwarding rule is created
  - `producer_instance_project_id` = "your-producer-project-id"  # Project hosting the service (e.g., Cloud SQL)
  - `subnetwork_name` = "your-subnetwork-name"  # Subnet for allocating the internal IP
  - `network_name` = "your-network-name"  # VPC network for the forwarding rule
  - `ip_address_literal` = ""  # (Optional) Specific internal IP, or leave empty for automatic allocation
  - `region` = "your-region"  # Region for the resources
  - `producer_cloudsql` = { "instance_name" = "your-sql-instance-name" }  # (Optional) Name of the producer Cloud SQL instance
  - `producer_alloydb` = { "instance_name" = "your-alloydb-instance-name", "cluster_id" = "your-cluster-id" }  # (Optional) AlloyDB instance and cluster ID
  - `target` = "your-service-attachment-link"  # (Optional) Service attachment link

- **Exactly one of `producer_cloudsql.instance_name`, `producer_alloydb.instance_name`, or `target` must be specified.**
- **Regions:** Ensure your producer service, subnetwork, and service attachment all reside within the same GCP region.
- **IP Addresses:** Verify the specified `ip_address_literal` values are available and not already in use.

## Example configuration/producer-connectivity.tfvars

Example 1: Connecting to a Cloud SQL instance

This example shows how to connect to a specific Cloud SQL instance using its name.

```
psc_endpoints = [
  {
    endpoint_project_id = "your-consumer-project-id"
    producer_instance_project_id = "your-producer-project-id"
    subnetwork_name = "your-subnetwork-name"
    network_name = "your-network-name"
    ip_address_literal = "10.128.0.20"  // Or leave empty for dynamic allocation
    region = "us-central1"
    producer_cloudsql = {
      instance_name = "your-cloud-sql-instance-name"
    }
  }
]
```

Example 2: Connecting to a service attachment

This example demonstrates connecting to a service attachment, which can represent a different type of producer service or a group of instances.

```
psc_endpoints = [
  {
    endpoint_project_id = "your-consumer-project-id"
    producer_instance_project_id = "your-producer-project-id"
    subnetwork_name = "your-subnetwork-name"
    network_name = "your-network-name"
    ip_address_literal = ""  // Or specify an IP address if needed
    region = "us-central1"
    target = "projects/your-project/regions/us-central1/serviceAttachments/your-service-attachment-name"
  }
]
```

Example 3: Connecting to an AlloyDB instance

This example shows how to connect to an AlloyDB instance.

```
psc_endpoints = [
  {
    endpoint_project_id = "your-consumer-project-id"
    producer_instance_project_id = "your-producer-project-id"
    subnetwork_name = "your-subnetwork-name"
    network_name = "your-network-name"
    ip_address_literal = ""  // Or specify an IP address if needed
    region = "us-central1"
    producer_alloydb = {
      instance_name = "your-alloydb-instance-id"
      cluster_id = "your-alloydb-cluster-id"
    }
  }
]
```

## Usage

**NOTE** : run the terraform commands with the -var-file referencing the `producer-connectivity.tfvars` present under the `/configuration` folder. 

Example :

1. **Initialize**: Run `terraform init`.
2. **Plan**: Run terraform plan -var-file=../../configuration/producer-connectivity.tfvars to review the planned changes.

```
terraform plan -var-file=../../configuration/producer-connectivity.tfvars
```
3. **Apply**: If the plan looks good, run terraform apply -var-file=../../configuration/producer-connectivity.tfvars to create or update the resources.

```
terraform apply -var-file=../../configuration/producer-connectivity.tfvars
```

<!-- BEGIN_TF_DOCS -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_psc_forwarding_rules"></a> [psc\_forwarding\_rules](#module\_psc\_forwarding\_rules) | ../../modules/psc_forwarding_rule | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_psc_endpoints"></a> [psc\_endpoints](#input\_psc\_endpoints) | List of PSC Endpoint configurations | <pre>list(object({<br>    # The Google Cloud project ID where the forwarding rule and address will be created.<br>    endpoint_project_id = string<br><br>    # The Google Cloud project ID where the Cloud SQL instance is located.<br>    producer_instance_project_id = string<br><br>    # The name of the subnet where the internal IP address will be allocated.<br>    subnetwork_name = string<br><br>    # The name of the network where the forwarding rule will be created.<br>    network_name = string<br><br>    # The region where the forwarding rule and address will be created.<br>    region = optional(string)<br><br>    # Optional: The static internal IP address to use. If not provided,<br>    # Google Cloud will automatically allocate an IP address.<br>    ip_address_literal = optional(string, "")<br>    # Allow access to the PSC endpoint from any region.<br>    allow_psc_global_access = optional(bool, false)<br>    # Resource labels to apply to the forwarding rule.<br>    labels = optional(map(string), {})<br><br>    # The name of the Cloud SQL instance.<br>    producer_cloudsql_instance_name = optional(string)<br><br>    # The name of the AlloyDB instance.<br>    producer_alloydb_instance_name = optional(string)<br><br>    # The ID of the AlloyDB cluster.<br>    alloydb_cluster_id = optional(string)<br><br>    # The target for the forwarding rule.<br>    target = optional(string)<br>  }))</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_forwarding_rule_self_link"></a> [forwarding\_rule\_self\_link](#output\_forwarding\_rule\_self\_link) | Self-links of the created forwarding rules |
| <a name="output_address_self_link"></a> [address\_self\_link](#output\_address\_self\_link) | Self-links of the created addresses |
| <a name="output_ip_address_literal"></a> [ip\_address\_literal](#output\_ip\_address\_literal) | IP addresses of the created addresses |
| <a name="output_forwarding_rule_target"></a> [forwarding\_rule\_target](#output\_forwarding\_rule\_target) | Targets for the PSC forwarding rules |

<!-- END_TF_DOCS -->