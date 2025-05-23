## Firewall Policies

### Overview

Google Cloud Network Firewall Policies are a key component of Google Cloud's network security, offering a more centralized and scalable way to manage firewall rules compared to traditional Virtual Private Cloud (VPC) firewall rules. They are part of Google Cloud's Next Generation Firewall (NGFW) service, designed to provide advanced protection capabilities and granular control over network traffic.

This module allows creation and management of three different firewall policy types:

- A [hierarchical policy](https://cloud.google.com/firewall/docs/firewall-policies) in a folder or organization.
- A [global](https://cloud.google.com/vpc/docs/network-firewall-policies) network policy.
- A [regional](https://cloud.google.com/vpc/docs/regional-firewall-policies) network policy.

### Pre-Requisites

#### Prior Step Completion:

  - **Completed Prior Stages:** Successful deployment of the `security/FirewallPolicy` stage requires the completion of the following stages:

      - **01-organization:** This stage handles the activation of required Google Cloud APIs.
      - **02-networking:** This stage sets up the necessary network infrastructure, such as VPCs.

#### Permissions:

The user or service account executing Terraform must have the following roles (or equivalent permissions):

- Security Admin (roles/iam.securityAdmin)
- Firewall Policy Admin (roles/compute.orgFirewallPolicyAdmin)

### Execution Steps

1. **Configuration:**

    - clarity : Create YAML configuration files (e.g., hierarchical-instance-lite.yaml) within the directory specified by the config_folder_path variable.
    - Edit the YAML files to define the desired configuration for network firewall policy configurations. (See **Examples** below)

2. **Terraform Initialization:**

    - Open your terminal and navigate to the directory containing the Terraform configuration.
    - Run the following command to initialize Terraform:

      ```bash
      terraform init
      ```

3. **Review the Execution Plan:**

    - Generate an execution plan with the following command to review changes Terraform will make to your Google Cloud infrastructure:

      ```bash
      terraform plan -var-file=../../../configuration/security/Firewall/FirewallPolicy/firewallpolicy.tfvars
      ```

4. **Apply the Configuration:**

    - Once satisfied with the plan, execute the terraform apply command to provision your network firewall policies as needed:

      ```bash
      terraform apply -var-file=../../../configuration/security/Firewall/FirewallPolicy/firewallpolicy.tfvars
      ```

5. **Monitor and Manage:**

    * Use Terraform to manage updates and changes to your network firewall policies as needed.

### Examples

1. Hierarchical Firewall Policy
    ```
    name      : "instance-hierarchicalpolicy"
    parent_id : <Replace with Folder ID or Organisation ID> #e.g. folder/11111222 or organization/111111222
    attachments :
      test : <Replace with Folder ID or Organisation ID> #e.g. folder/11111222 or organization/111111222
    egress_rules :
      - smtp :
        priority : 900
        match :
          destination_ranges :
          - "10.1.1.0/24"
          layer4_configs :
          - protocol : tcp
            ports :
            - 25
    ```

2. Global Firewall Policy
    ```
    name      : "global-firewallpolicy"
    parent_id : <Replace with Project Id> # project_id incase of global/regional firewall policy.
    region : global
    attachments :
      vpc1 : <Replace with VPC Self Link> # e.g."projects/project-id/global/networks/vpc-name"
    egress_rules :
      - smtp :
        priority : 1002
        match :
          destination_ranges :
          - "10.1.1.0/24"
          layer4_configs :
          - protocol : tcp
            ports :
            - 25
    ```


3. Regional Firewall Policy
    ```
    name      : "regional-firewallpolicy"
    parent_id : <Replace with Project Id>  # project_id incase of global/regional firewall policy.
    region : <Replace with region>         # e.g. us-central1
    attachments :
      vpc : <Replace with VPC Self Link>   # e.g."projects/project-id/global/networks/vpc-name"
    egress_rules :
      - smtp :
        priority : 1000
        match :
          destination_ranges :
          - "10.1.1.0/24"
          layer4_configs :
          - protocol : tcp
            ports :
            - 25
    ```



<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_network_firewall_policy"></a> [network\_firewall\_policy](#module\_network\_firewall\_policy) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/net-firewall-policy | v40.1.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_address_groups"></a> [address\_groups](#input\_address\_groups) | Address groups which should be matched against the traffic. | `list(string)` | `null` | no |
| <a name="input_attachments"></a> [attachments](#input\_attachments) | Ids of the resources to which this policy will be attached, in descriptive name => self link format. Specify folders or organization for hierarchical policy, VPCs for network policy. | `map(string)` | `{}` | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding Firewall Policy configuration values. | `string` | `"../../../../configuration/security/Firewall/FirewallPolicy/config"` | no |
| <a name="input_description"></a> [description](#input\_description) | An optional description of this resource. Provide this property when you create the resource. | `string` | `null` | no |
| <a name="input_destination_ranges"></a> [destination\_ranges](#input\_destination\_ranges) | If destination ranges are specified, the firewall will apply only to traffic that has destination IP address in these ranges. | `list(string)` | `null` | no |
| <a name="input_disabled"></a> [disabled](#input\_disabled) | Denotes whether the firewall rule is disabled, i.e not applied to the network it is associated with. When set to true, the firewall rule is not enforced and the network behaves as if it did not exist. If this is unspecified, the firewall rule will be enabled. | `bool` | `false` | no |
| <a name="input_egress_action"></a> [egress\_action](#input\_egress\_action) | Action of the Egress Firewall Rule. | `string` | `"deny"` | no |
| <a name="input_egress_rules"></a> [egress\_rules](#input\_egress\_rules) | List of egress rule definitions, action can be 'allow', 'deny', 'goto\_next' or 'apply\_security\_profile\_group'. The match.layer4configs map is in protocol => optional [ports] format. | <pre>map(object({<br>    priority                = number<br>    action                  = optional(string, "deny")<br>    description             = optional(string)<br>    disabled                = optional(bool, false)<br>    enable_logging          = optional(bool)<br>    security_profile_group  = optional(string)<br>    target_resources        = optional(list(string))<br>    target_service_accounts = optional(list(string))<br>    target_tags             = optional(list(string))<br>    tls_inspect             = optional(bool, null)<br>    match = object({<br>      address_groups       = optional(list(string))<br>      fqdns                = optional(list(string))<br>      region_codes         = optional(list(string))<br>      threat_intelligences = optional(list(string))<br>      destination_ranges   = optional(list(string))<br>      source_ranges        = optional(list(string))<br>      source_tags          = optional(list(string))<br>      layer4_configs = optional(list(object({<br>        protocol = optional(string, "all")<br>        ports    = optional(list(string))<br>      })), [{}])<br>    })<br>  }))</pre> | `{}` | no |
| <a name="input_enable_logging"></a> [enable\_logging](#input\_enable\_logging) | This field denotes whether to enable logging for a particular firewall rule. | `bool` | `null` | no |
| <a name="input_fqdns"></a> [fqdns](#input\_fqdns) | Fully Qualified Domain Name (FQDN) which should be matched against traffic | `list(string)` | `null` | no |
| <a name="input_ingress_action"></a> [ingress\_action](#input\_ingress\_action) | Action of the Ingress Firewall Rule. | `string` | `"allow"` | no |
| <a name="input_ingress_rules"></a> [ingress\_rules](#input\_ingress\_rules) | List of ingress rule definitions, action can be 'allow', 'deny', 'goto\_next' or 'apply\_security\_profile\_group'. | <pre>map(object({<br>    priority                = number<br>    action                  = optional(string, "allow")<br>    description             = optional(string)<br>    disabled                = optional(bool, false)<br>    enable_logging          = optional(bool)<br>    security_profile_group  = optional(string)<br>    target_resources        = optional(list(string))<br>    target_service_accounts = optional(list(string))<br>    target_tags             = optional(list(string))<br>    tls_inspect             = optional(bool, null)<br>    match = object({<br>      address_groups       = optional(list(string))<br>      fqdns                = optional(list(string))<br>      region_codes         = optional(list(string))<br>      threat_intelligences = optional(list(string))<br>      destination_ranges   = optional(list(string))<br>      source_ranges        = optional(list(string))<br>      source_tags          = optional(list(string))<br>      layer4_configs = optional(list(object({<br>        protocol = optional(string, "all")<br>        ports    = optional(list(string))<br>      })), [{}])<br>    })<br>  }))</pre> | `{}` | no |
| <a name="input_layer4_configs"></a> [layer4\_configs](#input\_layer4\_configs) | Pairs of IP protocols and ports that the rule should match. | <pre>list(object({<br>    protocol = optional(string, "all")<br>    ports    = optional(list(string))<br>  }))</pre> | `null` | no |
| <a name="input_region"></a> [region](#input\_region) | Policy region. Leave null for hierarchical policy, set to 'global' for a global network policy. | `string` | `null` | no |
| <a name="input_region_codes"></a> [region\_codes](#input\_region\_codes) | Region codes whose IP addresses will be used to match for traffic. | `list(string)` | `null` | no |
| <a name="input_security_profile_group"></a> [security\_profile\_group](#input\_security\_profile\_group) | A fully-qualified URL of a SecurityProfile resource instance. Example: https://networksecurity.googleapis.com/v1/projects/{project}/locations/{location}/securityProfileGroups/my-security-profile-group | `string` | `null` | no |
| <a name="input_source_ranges"></a> [source\_ranges](#input\_source\_ranges) | If source ranges are specified, the firewall will apply only to traffic that has source IP address in these ranges. These ranges must be expressed in CIDR format. | `list(string)` | `null` | no |
| <a name="input_source_tags"></a> [source\_tags](#input\_source\_tags) | A list of source tags. | `list(string)` | `null` | no |
| <a name="input_target_resources"></a> [target\_resources](#input\_target\_resources) | A list of network resource URLs to which this rule applies. This field allows you to control which network's VMs get this rule. If this field is left blank, all VMs within the organization will receive the rule. | `list(string)` | `null` | no |
| <a name="input_target_service_accounts"></a> [target\_service\_accounts](#input\_target\_service\_accounts) | (Optional) A list of service accounts indicating the sets of instances that are applied with this rule. | `list(string)` | `null` | no |
| <a name="input_target_tags"></a> [target\_tags](#input\_target\_tags) | A list of instance tags indicating sets of instances located in the network that may make network connections as specified in allowed[]. If no targetTags are specified, the firewall rule applies to all instances on the specified network. | `list(string)` | `null` | no |
| <a name="input_threat_intelligences"></a> [threat\_intelligences](#input\_threat\_intelligences) | Names of Network Threat Intelligence lists. The IPs in these lists will be matched against traffic destination. | `list(string)` | `null` | no |
| <a name="input_tls_inspect"></a> [tls\_inspect](#input\_tls\_inspect) | (Optional) Boolean flag indicating if the traffic should be TLS decrypted. Can be set only if action = 'apply\_security\_profile\_group' and cannot be set for other actions. | `bool` | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_id"></a> [id](#output\_id) | Fully qualified network firewall policy ids. |
<!-- END_TF_DOCS -->
