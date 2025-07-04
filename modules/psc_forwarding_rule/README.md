# Terraform PSC Forwarding Rule Module

This module creates Private Service Connect (PSC) forwarding rules and internal IP addresses for connecting to Cloud SQL or AlloyDB instances via PSC endpoints.
## Usage

Basic usage of this module is as follows:

```hcl
module "psc_forwarding_rule" {
     source = "../../modules/psc_forwarding_rule"

     psc_endpoints = [
          // Configuration for a PSC endpoint with a CloudSQL instance
          // This configuration includes a producer_instance_name field that specifies the CloudSQL instance name.
          // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
          {
               endpoint_project_id          = "your-endpoint-project-id"
               producer_instance_project_id = "your-producer-instance-project-id"
               subnetwork_name              = "subnetwork-1"
               network_name                 = "network-1"
               ip_address_literal           = "10.128.0.26"
               region                       = "" # Example: us-central1
               producer_cloudsql = {
                    instance_name = "psc-instance-name"
               }
          },
          // Configuration for a PSC endpoint with an AlloyDB instance
          // This configuration includes producer_alloydb_instance_name and cluster_id fields that specify the AlloyDB instance and cluster.
          // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
          {
               endpoint_project_id          = "your-endpoint-project-id"
               producer_instance_project_id = "your-producer-instance-project-id"
               subnetwork_name              = "subnetwork-2"
               network_name                 = "network-2"
               ip_address_literal           = "10.128.0.27"
               region                       = "" # Example: us-central2
               producer_alloydb = {
                    instance_name = "your-alloydb-instance-name"
                    cluster_id    = "your-cluster-id"
               }
          },
          // Configuration for a PSC endpoint with a target
          // This configuration includes a target field that specifies the service attachment URL.
          // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
          {
               endpoint_project_id          = "your-endpoint-project-id"
               producer_instance_project_id = "your-producer-instance-project-id"
               subnetwork_name              = "subnetwork-3"
               network_name                 = "network-3"
               ip_address_literal           = "10.0.0.10"
               region                       = "" # Example: us-central1
               target                       = "projects/your-project-id/regions/us-central1/serviceAttachments/your-service-attachment-id"
          }
     ]
}
```

<!-- BEGIN_TF_DOCS -->
Copyright 2024-2025 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.7.0 |
| <a name="requirement_google"></a> [google](#requirement\_google) | ~> 6.28 |
| <a name="requirement_google-beta"></a> [google-beta](#requirement\_google-beta) | ~> 6.28 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | ~> 6.28 |

## Resources

| Name | Type |
|------|------|
| [google_compute_address.psc_address](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_address) | resource |
| [google_compute_forwarding_rule.psc_forwarding_rule](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_forwarding_rule) | resource |
| [google_alloydb_instance.alloydb_instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/alloydb_instance) | data source |
| [google_sql_database_instance.cloudsql_instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/sql_database_instance) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_psc_endpoints"></a> [psc\_endpoints](#input\_psc\_endpoints) | List of PSC Endpoint configurations | <pre>list(object({<br/>    # The Google Cloud project ID where the forwarding rule and address will be created.<br/>    endpoint_project_id = string<br/><br/>    # The Google Cloud project ID where the Cloud SQL instance is located.<br/>    producer_instance_project_id = string<br/><br/>    # The name of the subnet where the internal IP address will be allocated.<br/>    subnetwork_name = string<br/><br/>    # The name of the network where the forwarding rule will be created.<br/>    network_name = string<br/><br/>    # The region where the forwarding rule and address will be created.<br/>    region = optional(string)<br/><br/>    # Optional: The static internal IP address to use. If not provided,<br/>    # Google Cloud will automatically allocate an IP address.<br/>    ip_address_literal = optional(string, "")<br/><br/>    # Allow access to the PSC endpoint from any region.<br/>    allow_psc_global_access = optional(bool, false)<br/><br/>    # Resource labels to apply to the forwarding rule.<br/>    labels = optional(map(string), {})<br/><br/>    # The Cloud SQL instance.<br/>    producer_cloudsql = optional(object({<br/>      # The name of the Cloud SQL instance.<br/>      instance_name = optional(string)<br/>    }), {})<br/><br/>    # The AlloyDB instance.<br/>    producer_alloydb = optional(object({<br/>      # The name of the AlloyDB instance.<br/>      instance_name = optional(string)<br/>      # The ID of the AlloyDB cluster.<br/>      cluster_id = optional(string)<br/>    }), {})<br/><br/>    # The target for the forwarding rule.<br/>    target = optional(string)<br/>  }))</pre> | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_address_self_link"></a> [address\_self\_link](#output\_address\_self\_link) | Map of created address self-links (for static IPs), keyed by endpoint index. |
| <a name="output_forwarding_rule_self_link"></a> [forwarding\_rule\_self\_link](#output\_forwarding\_rule\_self\_link) | Map of created forwarding rule self-links, keyed by endpoint index. |
| <a name="output_forwarding_rule_target"></a> [forwarding\_rule\_target](#output\_forwarding\_rule\_target) | Map of forwarding rule targets, keyed by endpoint index |
| <a name="output_ip_address_literal"></a> [ip\_address\_literal](#output\_ip\_address\_literal) | Map of created address IP literals (for static IPs), keyed by endpoint index. |
<!-- END_TF_DOCS -->