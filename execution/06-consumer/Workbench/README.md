# Vertex AI Workbench

## Overview

This Terraform solution offers a streamlined and modular approach to deploying and managing Google Cloud Vertex AI Workbench instances. Vertex AI Workbench provides scalable, fully managed Jupyter notebooks on Google Cloud, designed for high availability and ease of use. The solution leverages Terraform modules, with the `workbench` module utilizing the `vertex-ai` module from the Google Cloud Platform. This design simplifies the creation and configuration of Workbench instances by using YAML configuration files. These YAML files allow users to define customizable parameters for each Workbench instance in a structured and readable format.

## Pre-Requisites

### Prior Step Completion:

- **Completed Prior Stages:** Successful deployment of Workbench resources depends on the completion of the following stages:

    - **01-organization:** Activation of required Google Cloud APIs for Workbench.
    - **02-networking:** Setup of necessary network infrastructure, including VPCs and subnets, to support Workbench connectivity. For Workbench, ensure that you use a custom mode subnet VPC to enable proper configuration and connectivity. Additionally, set `create_nat` and `enable private access` to `true` to ensure proper network address translation and private Google access for the Workbench instances.
    - **03-security:** Configuration of firewall rules to allow access to Workbench instances on appropriate ports and IP ranges.

### Enabled APIs:

Ensure the following Google Cloud APIs are enabled in your project:

- Vertex AI API
- Cloud DNS API

### Permissions:

The user or service account executing Terraform must have the following roles (or equivalent permissions):

-   **Vertex AI Admin (`roles/aiplatform.admin`)**: This role is essential for managing Vertex AI Workbench instances, allowing the creation, modification, and deletion of these resources.
-   **Service Account User (`roles/iam.serviceAccountUser`)**: If you are using service accounts for your Workbench instances, this role is required to allow the user or service account executing Terraform to act as the specified service account.
-   **Service Account Token Creator (`roles/iam.serviceAccountTokenCreator`)**: This role is necessary when using service accounts, as it grants permission to generate access tokens for the service account, enabling it to interact with Google Cloud services.

## Execution Steps

1. **Configuration:**

    - Define your YAML configurations for each Workbench instance.
    - You can place them in the `configuration/consumer/Workbench/config` folder.

2. **Terraform Initialization:**

    - Open your terminal and navigate to the directory containing your Terraform configuration.
    - Run the following command to initialize Terraform:

    ```bash
    terraform init
    ```

3. **Review the Execution Plan:**

    - Use the following command to generate an execution plan. This will show you the changes Terraform will make to your Google Cloud infrastructure:

    ```bash
    terraform plan
    ```

   Carefully review the plan to ensure it aligns with your intended configuration.

4. **Apply the Configuration:**

    Once you're satisfied with the plan, execute the terraform apply command to provision your Workbench instances:

    ```bash
    terraform apply
    ```

   Terraform will create the corresponding Workbench instances in your Google Cloud project based on your configurations.

5. **Execute with `run.sh`:**

   - You can also execute this stage using the `run.sh` script.
   - To do this, you need to set the `stage` variable to `06-consumer/Workbench`.
   - Additionally, you can specify the `config_folder_path` variable to indicate the location of your YAML files.
   - Here's an example of how to run it:

     ```bash
     ./run.sh --stage=06-consumer/Workbench --vars='config_folder_path=../../../configuration/consumer/Workbench/workbench.tfvars'
     ```

## Important Notes

- The solution assumes that all required network and subnetwork resources are pre-configured in your project as outlined in the previous steps.
- Provide accurate service account credentials (if applicable) to enable Terraform to interact seamlessly with your Google Cloud project.
- Review the `variables.tf` file for a comprehensive list of available variables and their detailed descriptions.
- The Terraform modules used in this solution (`GoogleCloudPlatform/vertex-ai/google//modules/workbench`) include additional configuration options. Refer to their official documentation to explore and implement further customizations as needed.

## Examples

### Sample YAML

Here is a sample YAML configuration for a Workbench instance:

```yaml
name: <your-workbench-instance-name>
project_id: <your-project-id>
location: us-central1-a
gce_setup:
  disable_public_ip: true
  network_interfaces:
  - network: projects/<your-project-id>/global/networks/<your-network-name>
    subnet: projects/<your-project-id>/regions/us-central1/subnetworks/<your-subnetwork-name>
```

<!-- BEGIN_TF_DOCS -->
## Modules

| Name                | Source                                             | Version |
|---------------------|----------------------------------------------------|---------|
| workbench_instance  | GoogleCloudPlatform/vertex-ai/google//modules/workbench | ~> 1.0 |

## Inputs

| Name                          | Description                                                   | Type         | Default                                      | Required |
|-------------------------------|---------------------------------------------------------------|--------------|----------------------------------------------|----------|
| region                        | The region where resources will be created.                   | `string`     | `"us-central1"`                              | no       |
| location                      | The zone where resources will be created.                     | `string`     | `"us-central1-a"`                            | no       |
| config_folder_path            | Location of YAML files holding Workbench configuration values.| `string`     | `"../../../configuration/consumer/Workbench/config"` | no       |
| machine_type                  | Machine type for the Workbench instance.                      | `string`     | `"e2-standard-2"`                            | no       |
| network_tags                  | Network tags for the Workbench instance.                      | `list(string)` | `[]`                                       | no       |
| data_disk_size                | Disk size in GB for data disk.                                | `number`     | `100`                                       | no       |
| data_disk_type                | Data disk type.                                               | `string`     | `"PD_SSD"`                                  | no       |
| boot_disk_size_gb             | Default boot disk size in GB.                                 | `number`     | `200`                                       | no       |
| boot_disk_type                | Default boot disk type.                                       | `string`     | `"PD_STANDARD"`                              | no       |
| nic_type                      | NIC type for the Workbench instance.                          | `string`     | `"GVNIC"`                                   | no       |
| vm_image_project              | Image project for the VM image.                               | `string`     | `"cloud-notebooks-managed"`                 | no       |
| vm_image_family               | Image family for the VM image.                                | `string`     | `"workbench-instances"`                     | no       |
| vm_image_name                 | Default VM image name for Workbench instances.                | `string`     | `null`                                      | no       |
| labels                        | Default labels for Workbench instance.                        | `map(string)` | `{ "owner": "your-desired-owner" }`         | no       |
| disable_public_ip_default     | Default setting for disabling public IPs.                     | `bool`       | `true`                                      | no       |
| disable_proxy_access_default  | Default setting for disabling proxy access.                   | `bool`       | `true`                                      | no       |
| disk_encryption_default       | Default disk encryption key for the Workbench instance.       | `string`     | `null`                                      | no       |
| enable_secure_boot_default    | Default setting for enabling secure boot.                     | `bool`       | `false`                                     | no       |
| enable_vtpm_default           | Default setting for enabling vTPM.                            | `bool`       | `false`                                     | no       |
| enable_integrity_monitoring_default | Default setting for enabling integrity monitoring.       | `bool` | `false` | no       |
| internal_ip_only              | Specifies whether the Workbench instance should use only internal IP addresses. | `bool` | `true` | no       |
| metadata_configs              | Predefined metadata to apply to this instance.                | `object`     | `{}`                                        | no       |
| metadata                      | Custom metadata to apply to this instance.                    | `map(string)` | `{}`                                       | no       |
| instance_owners               | List of email addresses of users who will have owner permissions on the Workbench instance. | `list(string)` | `[]` | no       |

## Outputs

| Name                          | Description                                                   |
|-------------------------------|---------------------------------------------------------------|
| workbench_instance_ids        | The IDs of the created Workbench instances.                   |
| workbench_instance_proxy_uris | The proxy URIs of the created Workbench instances.            |
<!-- END_TF_DOCS -->