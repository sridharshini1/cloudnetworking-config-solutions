# Terraform Network Connectivity Center Module

 It creates a Network Connectivity Center Hub and attaches spokes.

## Usage

Basic usage of this submodule is as follows:

```hcl
module "ncc" {
    source  = "../../modules/network-connectivity-center"
    version = "~> 9.0.0"

    project_id   = "<PROJECT ID>"
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
| [google_network_connectivity_group.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_group) | resource |
| [google_network_connectivity_hub.hub](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_hub) | resource |
| [google_network_connectivity_spoke.hybrid_spoke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_spoke) | resource |
| [google_network_connectivity_spoke.producer_vpc_spoke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_spoke) | resource |
| [google_network_connectivity_spoke.router_appliance_spoke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_spoke) | resource |
| [google_network_connectivity_spoke.vpc_spoke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_connectivity_spoke) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auto_accept_projects"></a> [auto\_accept\_projects](#input\_auto\_accept\_projects) | List of projects to auto-accept. | `list(string)` | `[]` | no |
| <a name="input_create_new_hub"></a> [create\_new\_hub](#input\_create\_new\_hub) | Indicates if a new hub should be created. | `bool` | `false` | no |
| <a name="input_existing_hub_uri"></a> [existing\_hub\_uri](#input\_existing\_hub\_uri) | URI of an existing NCC hub to use, if null a new one is created. | `string` | `null` | no |
| <a name="input_export_psc"></a> [export\_psc](#input\_export\_psc) | Whether Private Service Connect transitivity is enabled for the hub. | `bool` | `false` | no |
| <a name="input_group_decription"></a> [group\_decription](#input\_group\_decription) | Description for the network connectivity group. | `string` | `"Used for auto-accepting projects"` | no |
| <a name="input_group_name"></a> [group\_name](#input\_group\_name) | Name of the network connectivity group. | `string` | `"default"` | no |
| <a name="input_hybrid_spokes"></a> [hybrid\_spokes](#input\_hybrid\_spokes) | A map of Hybrid spokes (VPN/Interconnect) to be created. | <pre>map(object({<br/>    project_id                 = string<br/>    location                   = optional(string, "global")<br/>    description                = optional(string)<br/>    spoke_type                 = string # "vpn" or "interconnect"<br/>    uris                       = list(string)<br/>    site_to_site_data_transfer = optional(bool, false)<br/>    labels                     = optional(map(string))<br/>  }))</pre> | `{}` | no |
| <a name="input_ncc_hub_description"></a> [ncc\_hub\_description](#input\_ncc\_hub\_description) | This can be used to provide additional context or details about the purpose or usage of the hub. | `string` | `"Network Connectivity Center hub for managing and connecting multiple network resources."` | no |
| <a name="input_ncc_hub_labels"></a> [ncc\_hub\_labels](#input\_ncc\_hub\_labels) | Labels to be attached to network connectivity center hub resource. | `map(string)` | <pre>{<br/>  "environment": "prod",<br/>  "owner": "network-team"<br/>}</pre> | no |
| <a name="input_ncc_hub_name"></a> [ncc\_hub\_name](#input\_ncc\_hub\_name) | The Name of the NCC Hub. | `string` | n/a | yes |
| <a name="input_policy_mode"></a> [policy\_mode](#input\_policy\_mode) | Policy mode for the NCC hub. | `string` | `"PRESET"` | no |
| <a name="input_preset_topology"></a> [preset\_topology](#input\_preset\_topology) | Preset topology for the NCC hub. | `string` | `"MESH"` | no |
| <a name="input_producer_vpc_spokes"></a> [producer\_vpc\_spokes](#input\_producer\_vpc\_spokes) | A map of Producer VPC spokes to be created. | <pre>map(object({<br/>    project_id            = string<br/>    location              = optional(string, "global")<br/>    description           = optional(string)<br/>    uri                   = string # In this context, it's the network URI<br/>    peering               = string<br/>    labels                = optional(map(string))<br/>    exclude_export_ranges = optional(list(string))<br/>    include_export_ranges = optional(list(string))<br/>  }))</pre> | `{}` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Project ID for NCC Hub resources. | `string` | n/a | yes |
| <a name="input_router_appliance_spokes"></a> [router\_appliance\_spokes](#input\_router\_appliance\_spokes) | A map of Router Appliance spokes to be created. | <pre>map(object({<br/>    location    = string<br/>    description = optional(string)<br/>    instances = list(object({<br/>      virtual_machine = string<br/>      ip_address      = string<br/>    }))<br/>    site_to_site_data_transfer = bool<br/>    labels                     = optional(map(string))<br/>  }))</pre> | `{}` | no |
| <a name="input_spoke_labels"></a> [spoke\_labels](#input\_spoke\_labels) | Default labels to be merged with spoke-specific labels. | `map(string)` | <pre>{<br/>  "environment": "prod",<br/>  "owner": "network-team"<br/>}</pre> | no |
| <a name="input_vpc_spokes"></a> [vpc\_spokes](#input\_vpc\_spokes) | A map of VPC spokes to be created. The key should be the spoke name. | <pre>map(object({<br/>    project_id            = string<br/>    uri                   = string<br/>    description           = optional(string)<br/>    labels                = optional(map(string))<br/>    exclude_export_ranges = optional(list(string))<br/>    include_export_ranges = optional(list(string))<br/>  }))</pre> | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_hybrid_spokes"></a> [hybrid\_spokes](#output\_hybrid\_spokes) | All created hybrid spoke resource objects |
| <a name="output_ncc_hub"></a> [ncc\_hub](#output\_ncc\_hub) | The NCC Hub object |
| <a name="output_producer_vpc_spokes"></a> [producer\_vpc\_spokes](#output\_producer\_vpc\_spokes) | All created producer VPC spoke resource objects |
| <a name="output_router_appliance_spokes"></a> [router\_appliance\_spokes](#output\_router\_appliance\_spokes) | All created router appliance spoke resource objects |
| <a name="output_vpc_spokes"></a> [vpc\_spokes](#output\_vpc\_spokes) | All created vpc spoke resource objects |
<!-- END_TF_DOCS -->