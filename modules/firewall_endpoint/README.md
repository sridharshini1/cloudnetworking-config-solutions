# Terraform Google Firewall Endpoint Module

This module creates a GCP Firewall Endpoint at the organization level and a Firewall Endpoint Association at the project level. It allows for creating either resource individually or creating both to link the endpoint to a VPC network.

## Usage

```terraform
module "firewall_endpoint" {
  source = "./modules/firewall-endpoint"

  # Firewall Endpoint settings (Organization Level)
  create_firewall_endpoint = true
  organization_id          = "123456789012"
  firewall_endpoint_name   = "my-org-fw-endpoint-us-central1-a"
  billing_project_id       = "my-billing-project-id"
  location                 = "us-central1-a" # Must be a zone

  # Firewall Endpoint Association settings (Project Level)
  create_firewall_endpoint_association = true
  association_name                     = "assoc-prod-vpc-us-central1"
  association_project_id               = "my-production-project-id"
  vpc_id                               = "projects/my-production-project/global/networks/prod-vpc-us-east"
}
```

<!-- BEGIN_TF_DOCS -->

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | n/a |

## Resources

| Name | Type |
|------|------|
| [google_network_security_firewall_endpoint.firewall_endpoint](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_firewall_endpoint) | resource |
| [google_network_security_firewall_endpoint_association.firewall_endpoint_association](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_firewall_endpoint_association) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_association_disabled"></a> [association\_disabled](#input\_association\_disabled) | If true, the association is created but will not intercept traffic. | `bool` | `false` | no |
| <a name="input_association_labels"></a> [association\_labels](#input\_association\_labels) | A map of labels to add to the Firewall Endpoint Association. | `map(string)` | `{}` | no |
| <a name="input_association_name"></a> [association\_name](#input\_association\_name) | The name of the Firewall Endpoint Association. | `string` | `null` | no |
| <a name="input_association_project_id"></a> [association\_project\_id](#input\_association\_project\_id) | The Project ID where the association will be created and where the VPC network resides. | `string` | `null` | no |
| <a name="input_billing_project_id"></a> [billing\_project\_id](#input\_billing\_project\_id) | The Project ID to be billed for the Firewall Endpoint's usage. | `string` | `null` | no |
| <a name="input_create_firewall_endpoint"></a> [create\_firewall\_endpoint](#input\_create\_firewall\_endpoint) | Set to true to create a Firewall Endpoint. | `bool` | `false` | no |
| <a name="input_create_firewall_endpoint_association"></a> [create\_firewall\_endpoint\_association](#input\_create\_firewall\_endpoint\_association) | Set to true to create a Firewall Endpoint Association. | `bool` | `false` | no |
| <a name="input_existing_firewall_endpoint_id"></a> [existing\_firewall\_endpoint\_id](#input\_existing\_firewall\_endpoint\_id) | The full resource ID of an existing Firewall Endpoint to use for an association if `create_firewall_endpoint` is false. | `string` | `null` | no |
| <a name="input_firewall_endpoint_labels"></a> [firewall\_endpoint\_labels](#input\_firewall\_endpoint\_labels) | A map of labels to add to the Firewall Endpoint. | `map(string)` | `{}` | no |
| <a name="input_firewall_endpoint_name"></a> [firewall\_endpoint\_name](#input\_firewall\_endpoint\_name) | The name of the Firewall Endpoint. | `string` | `null` | no |
| <a name="input_location"></a> [location](#input\_location) | The location (zone) for the Firewall Endpoint and its Association. E.g., 'us-central1-a'. | `string` | n/a | yes |
| <a name="input_organization_id"></a> [organization\_id](#input\_organization\_id) | The GCP Organization ID where the Firewall Endpoint will be created. | `string` | `null` | no |
| <a name="input_tls_inspection_policy_id"></a> [tls\_inspection\_policy\_id](#input\_tls\_inspection\_policy\_id) | The name (not the full path) of an optional TLS Inspection Policy to attach to the association. | `string` | `null` | no |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | The name of the VPC network to associate with the Firewall Endpoint. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_association_id"></a> [association\_id](#output\_association\_id) | The full resource ID of the created Firewall Endpoint Association. |
| <a name="output_association_name"></a> [association\_name](#output\_association\_name) | The name of the created Firewall Endpoint Association. |
| <a name="output_firewall_endpoint_id"></a> [firewall\_endpoint\_id](#output\_firewall\_endpoint\_id) | The full resource ID of the created Firewall Endpoint. |
| <a name="output_firewall_endpoint_name"></a> [firewall\_endpoint\_name](#output\_firewall\_endpoint\_name) | The name of the created Firewall Endpoint. |
| <a name="output_firewall_endpoint_self_link"></a> [firewall\_endpoint\_self\_link](#output\_firewall\_endpoint\_self\_link) | The self-link of the created Firewall Endpoint. |
<!-- END_TF_DOCS -->