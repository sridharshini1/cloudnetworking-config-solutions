# Network Connectivity Center (NCC)

## Overview

This Terraform configuration provides a modular and YAML-driven approach for deploying and managing Google Cloud Network Connectivity Center (NCC) resources. It enables you to create and manage NCC hubs, VPC spokes, producer VPC spokes, hybrid spokes, and router appliance spokes, supporting both mesh and hub-and-spoke topologies.

Key features of this configuration include:

- **YAML-driven configuration:** Define NCC hubs and spokes in YAML files for easy management and reproducibility.
- **Multiple topologies:** Choose between mesh and hub-and-spoke using the `preset_topology` variable.
- **Flexible spoke types:** Configure VPC, producer, hybrid, and router appliance spokes.
- **Auto-accept projects:** Automatically accept connections from specified projects.
- **PSC transitivity:** Optionally enable Private Service Connect transitivity.

### Benefits

- **Modularity:** Easily add or remove NCC resources by editing YAML files.
- **Reusability:** Use the same configuration structure for different environments.
- **Automation:** Supports automated deployment of complex NCC topologies.

## Prerequisites

Before creating NCC resources, ensure you have completed the following prerequisites:

1. **Completed Prior Stages:**  
   - **01-organization:** This stage handles the activation of required Google Cloud APIs.

2. **Enable the following APIs:**
    - [Compute Engine API](https://cloud.google.com/compute/docs/reference/rest/v1): Used for VPC networks, subnets, and related resources.
    - [Service Networking API](https://cloud.google.com/service-infrastructure/docs/service-networking/getting-started): Required for Private Service Access (PSA) configurations.
    - [Network Connectivity API](https://cloud.google.com/network-connectivity/docs/reference/networkconnectivity/rest): Enables NCC resources.
    - [Service Consumer Management API](https://cloud.google.com/service-infrastructure/docs/service-consumer-management/reference/rest): Required for Private Service Connect endpoints.

3. **Permissions required for this stage:**
    - [Network Connectivity Admin](https://cloud.google.com/iam/docs/understanding-roles#networkconnectivity.admin): `roles/networkconnectivity.admin` – Full control over NCC resources.
    - [Compute Network Admin](https://cloud.google.com/iam/docs/understanding-roles#compute.networkAdmin): `roles/compute.networkAdmin` – Manage VPC networks and related resources.

## Components

- `locals.tf`: Loads and processes YAML configuration files for NCC.
- `ncc.tf`: Instantiates the NCC module for each hub defined in the configuration.
- `variables.tf`: Input variables for customizing the deployment.
- `output.tf`: Exposes module outputs.

## Configuration

To configure NCC for your environment, create YAML files in the `../../../configuration/ncc/config/` directory. Example:

```yaml
hubs:
  - name: <hub_name>
    project_id: <hub_project_id>
    description: "Example NCC Hub"
    labels:
      env: prod
    export_psc: true
    policy_mode: PRESET
    preset_topology: MESH
    auto_accept_projects:
      - <hub_project_id>
      - <secondary_project_id>
    create_new_hub: false
    existing_hub_uri: "projects/<hub_project_id>/locations/global/hubs/<hub_name>"
    group_name: default
    group_decription: "Auto-accept group"
    spoke_labels:
      team: network

  - type: "producer_vpc_spoke"
    name: "producer-spoke-1"
    project_id: "producerspoke1-project-id"
    location: "global"
    uri: "projects/producerspoke1-project-id/global/networks/producer-spoke-1-vpc"
    description: "Producer VPC spoke for shared services"
    peering: "servicenetworking-googleapis-com"
    exclude_export_ranges: []
    labels:
      env: "prod"
```

## Usage

**NOTE:** Run Terraform commands with the `-var-file` referencing your NCC tfvars file if you override defaults.

```sh
terraform init
terraform plan
terraform apply
```

The module will read all YAML files in the config folder and create the corresponding NCC resources.

## Example Scenarios

### 1. Create a new NCC hub, spokes, and producer VPC spoke

```yaml
hubs:
  - name: <hub_name>
    project_id: <hub_project_id>
    description: "Example NCC Hub"
    labels:
      env: prod
    export_psc: true
    policy_mode: PRESET
    preset_topology: MESH
    auto_accept_projects:
      - <hub_project_id>
      - <secondary_project_id>
    create_new_hub: true
    existing_hub_uri: ""
    group_name: default
    group_decription: "Auto-accept group"
    spoke_labels:
      team: network
spokes:
  - type: "vpc_spoke"
    name: "spoke-1"
    project_id: "<spoke1_project_id>"
    uri: "projects/<spoke1_project_id>/global/networks/spoke-1-vpc"
    description: "Primary VPC spoke for production"
    labels:
      env: "prod"
  - type: "producer_vpc_spoke"
    name: "producer-spoke-1"
    project_id: "producerspoke1-project-id"
    location: "global"
    uri: "projects/producerspoke1-project-id/global/networks/producer-spoke-1-vpc"
    description: "Producer VPC spoke for shared services"
    peering: "servicenetworking-googleapis-com"
    exclude_export_ranges: []
    labels:
      env: "prod"
```

### 2. Use an existing NCC hub to create new spokes and a producer VPC spoke

```yaml
hubs:
  - name: <hub_name>
    project_id: <hub_project_id>
    description: "Example NCC Hub"
    labels:
      env: prod
    export_psc: true
    policy_mode: PRESET
    preset_topology: MESH
    auto_accept_projects:
      - <hub_project_id>
      - <secondary_project_id>
    create_new_hub: false
    existing_hub_uri: "projects/<hub_project_id>/locations/global/hubs/<hub_name>"
    group_name: default
    group_decription: "Auto-accept group"
    spoke_labels:
      team: network
spokes:
  - type: "vpc_spoke"
    name: "spoke-1"
    project_id: "<spoke1_project_id>"
    uri: "projects/<spoke1_project_id>/global/networks/spoke-1-vpc"
    description: "Primary VPC spoke for production"
    labels:
      env: "prod"
  - type: "producer_vpc_spoke"
    name: "producer-spoke-1"
    project_id: "producerspoke1-project-id"
    location: "global"
    uri: "projects/producerspoke1-project-id/global/networks/producer-spoke-1-vpc"
    description: "Producer VPC spoke for shared services"
    peering: "servicenetworking-googleapis-com"
    exclude_export_ranges: []
    labels:
      env: "prod"
```

### 3. Use an existing NCC hub and spoke to create a producer VPC spoke

```yaml
hubs:
  - name: <hub_name>
    project_id: <hub_project_id>
    description: "Example NCC Hub"
    labels:
      env: prod
    export_psc: true
    policy_mode: PRESET
    preset_topology: MESH
    auto_accept_projects:
      - <hub_project_id>
      - <secondary_project_id>
    create_new_hub: false
    existing_hub_uri: "projects/<hub_project_id>/locations/global/hubs/<hub_name>"
    group_name: default
    group_decription: "Auto-accept group"
    spoke_labels:
      team: network
spokes: 
  - type: "producer_vpc_spoke"
    name: "producer-spoke-1"
    project_id: "producerspoke1-project-id"
    location: "global"
    uri: "projects/producerspoke1-project-id/global/networks/producer-spoke-1-vpc"
    description: "Producer VPC spoke for shared services"
    peering: "servicenetworking-googleapis-com"
    exclude_export_ranges: []
    labels:
      env: "prod"
```

### 4. Create all resources together if a user does not have them

```yaml
hubs:
  - name: <hub_name>
    project_id: <hub_project_id>
    description: "Example NCC Hub"
    labels:
      env: prod
    export_psc: true
    policy_mode: PRESET
    preset_topology: MESH
    auto_accept_projects:
      - <hub_project_id>
      - <secondary_project_id>
    create_new_hub: true
    existing_hub_uri: ""
    group_name: default
    group_decription: "Auto-accept group"
    spoke_labels:
      team: network
spokes:
  - type: "vpc_spoke"
    name: "spoke-1"
    project_id: "<spoke1_project_id>"
    uri: "projects/<spoke1_project_id>/global/networks/spoke-1-vpc"
    description: "Primary VPC spoke for production"
    labels:
      env: "prod"
  - type: "producer_vpc_spoke"
    name: "producer-spoke-1"
    project_id: "producerspoke1-project-id"
    location: "global"
    uri: "projects/producerspoke1-project-id/global/networks/producer-spoke-1-vpc"
    description: "Producer VPC spoke for shared services"
    peering: "servicenetworking-googleapis-com"
    exclude_export_ranges: []
    labels:
      env: "prod"
```
## Usage

**NOTE** : run the terraform commands with the `-var-file` referencing the networking stage present under the /configuration folder.

## Outputs

- `ncc_module`: Outputs from the NCC module, including hub and spoke details.

## Notes

- Ensure all required APIs are enabled and permissions are granted.
- Adjust YAML fields as per your environment and naming conventions.
- For advanced topologies, refer to the [Google Cloud NCC documentation](https://cloud.google.com/network-connectivity/docs/network-connectivity-center).

<!-- BEGIN_TF_DOCS -->
## Requirements

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_network_connectivity_center"></a> [network\_connectivity\_center](#module\_network\_connectivity\_center) | ../../../modules/network-connectivity-center | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auto_accept_projects"></a> [auto\_accept\_projects](#input\_auto\_accept\_projects) | List of projects to auto-accept. | `list(string)` | `[]` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding NCC configuration values. | `string` | `"../../../configuration/networking/ncc/config"` | no |
| <a name="input_create_new_hub"></a> [create\_new\_hub](#input\_create\_new\_hub) | Indicates if a new hub should be created. | `bool` | `false` | no |
| <a name="input_existing_hub_uri"></a> [existing\_hub\_uri](#input\_existing\_hub\_uri) | URI of an existing NCC hub to use, if null a new one is created. | `string` | `null` | no |
| <a name="input_export_psc"></a> [export\_psc](#input\_export\_psc) | Whether Private Service Connect transitivity is enabled for the hub. | `bool` | `false` | no |
| <a name="input_group_decription"></a> [group\_decription](#input\_group\_decription) | Description for the network connectivity group. | `string` | `"Used for auto-accepting projects"` | no |
| <a name="input_group_name"></a> [group\_name](#input\_group\_name) | Name of the network connectivity group. | `string` | `"default"` | no |
| <a name="input_ncc_hub_description"></a> [ncc\_hub\_description](#input\_ncc\_hub\_description) | This can be used to provide additional context or details about the purpose or usage of the hub. | `string` | `"Network Connectivity Center hub for managing and connecting multiple network resources."` | no |
| <a name="input_ncc_hub_labels"></a> [ncc\_hub\_labels](#input\_ncc\_hub\_labels) | Labels to be attached to network connectivity center hub resource. | `map(string)` | `null` | no |
| <a name="input_policy_mode"></a> [policy\_mode](#input\_policy\_mode) | Policy mode for the NCC hub. | `string` | `"PRESET"` | no |
| <a name="input_preset_topology"></a> [preset\_topology](#input\_preset\_topology) | Preset topology for the NCC hub. | `string` | `"MESH"` | no |
| <a name="input_spoke_labels"></a> [spoke\_labels](#input\_spoke\_labels) | Labels to be attached to network connectivity center spoke resource. | `map(string)` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_ncc_module"></a> [ncc\_module](#output\_ncc\_module) | The NCC Module outputs |
<!-- END_TF_DOCS -->