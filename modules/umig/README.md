# Terraform Unmanaged Instance Group (UMIG) Module

This module creates a Google Cloud Unmanaged Instance Group (UMIG) and allows you to specify existing VM instances and named ports.

## Usage

This module creates an unmanaged instance group.

Here is a sample usage of the module:

```terraform
module "umig" {
  source     = "./modules/umig"
  name       = "cncs-umig"
  project_id = "your-project-id"
  zone       = "us-central1-a"
  instances  = ["my-instance-1", "my-instance-2"]
  named_ports = {
    "http" = 8080
  }
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
| [google_compute_instance_group.unmanaged](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_description"></a> [description](#input\_description) | Description for the instance group. | `string` | `"Instance group managed by the UMIG Terraform module."` | no |
| <a name="input_instances"></a> [instances](#input\_instances) | List of instance names. | `list(string)` | n/a | yes |
| <a name="input_name"></a> [name](#input\_name) | The name of the instance group. | `string` | n/a | yes |
| <a name="input_named_ports"></a> [named\_ports](#input\_named\_ports) | Map of named ports. | `map(number)` | `{}` | no |
| <a name="input_network"></a> [network](#input\_network) | The name or self-link of the network to filter instances. | `string` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The GCP project ID. | `string` | n/a | yes |
| <a name="input_zone"></a> [zone](#input\_zone) | The zone for the instance group. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_instances"></a> [instances](#output\_instances) | List of instance self-links in the Unmanaged Instance Group. |
| <a name="output_self_link"></a> [self\_link](#output\_self\_link) | The self-link of the Unmanaged Instance Group. |
<!-- END_TF_DOCS -->