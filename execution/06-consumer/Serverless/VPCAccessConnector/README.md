# Terraform VPC Serverless Connector Beta

This submodule is part of the the `terraform-google-network` module. It creates the vpc serverless connector using the beta components available.

It supports creating:

- serverless connector
- serverless vpc access connector

## Example Usage

1. Basic usage of this module which uses `subnet`.

    ```
    name: <connector-name>
    project_id: <connector-project>
    region: <connector-region>
    subnet_name: <subnet-to-attach>
    ```

2. Basic usage of this module which uses `network` and `IP CIDR` range.

    ```
    name: <connector-name>
    project_id: <connector-project>
    region: <connector-region>
    network: <network-to-attach>
    ip_cidr_range: <ip-cidr-range-to-assign>
    ```

<!-- BEGIN_TF_DOCS -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_serverless-connector"></a> [serverless-connector](#module\_serverless-connector) | terraform-google-modules/network/google//modules/vpc-serverless-connector-beta | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding VPC Access Connector configuration values. | `string` | `"../../../../configuration/consumer/Serverless/VPCAccessConnector/config"` | no |
| <a name="input_host_project_id"></a> [host\_project\_id](#input\_host\_project\_id) | Default host project ID for the subnet if not specified in YAML (for Shared VPC). | `string` | `null` | no |
| <a name="input_ip_cidr_range"></a> [ip\_cidr\_range](#input\_ip\_cidr\_range) | Default IP CIDR range if not specified in YAML. | `string` | `null` | no |
| <a name="input_machine_type"></a> [machine\_type](#input\_machine\_type) | Default machine type for the connector instances. | `string` | `"e2-standard-4"` | no |
| <a name="input_max_instances"></a> [max\_instances](#input\_max\_instances) | Default maximum number of instances (Note: not used by vpc-serverless-connector-beta module). | `number` | `null` | no |
| <a name="input_max_throughput"></a> [max\_throughput](#input\_max\_throughput) | Default maximum throughput in Mbps. | `number` | `null` | no |
| <a name="input_min_instances"></a> [min\_instances](#input\_min\_instances) | Default minimum number of instances (Note: not used by vpc-serverless-connector-beta module). | `number` | `null` | no |
| <a name="input_min_throughput"></a> [min\_throughput](#input\_min\_throughput) | Default minimum throughput in Mbps. | `number` | `null` | no |
| <a name="input_network"></a> [network](#input\_network) | Default VPC network name if not specified in YAML. | `string` | `null` | no |
| <a name="input_subnet_name"></a> [subnet\_name](#input\_subnet\_name) | Default subnet name if not specified in YAML. | `string` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_connector_ids"></a> [connector\_ids](#output\_connector\_ids) | The ID of the VPC Serverless Access connectors created. |
<!-- END_TF_DOCS -->
