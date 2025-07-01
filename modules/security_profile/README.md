# Terraform Google Security Profile Module

This module creates and manages GCP Security Profiles and Security Profile Groups. It supports `THREAT_PREVENTION`, `CUSTOM_MIRRORING`, and `CUSTOM_INTERCEPT` profile types and can create either a profile, a group, or both and link them together.

## Usage

```terraform
module "security_profile_and_group" {
  source = "./modules/security-profiles"

  # General Configuration
  organization_id = "123456789012"

  # Security Profile Configuration
  create_security_profile   = true
  security_profile_name     = "my-app-threat-profile"
  security_profile_type     = "THREAT_PREVENTION"
  security_profile_description = "A profile to prevent common threats for my-app."
  threat_prevention_profile = {
    severity_overrides = [
      {
        severity = "CRITICAL"
        action   = "DENY"
      }
    ]
    threat_overrides    = []
    antivirus_overrides = []
  }

  # Security Profile Group Configuration
  create_security_profile_group = true
  security_profile_group_name   = "my-app-security-group"

  # Link the new profile to the new group
  link_security_profile_to_group = true
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
| [google_network_security_security_profile.security_profile](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_security_profile) | resource |
| [google_network_security_security_profile_group.security_profile_group](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/network_security_security_profile_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_create_security_profile"></a> [create\_security\_profile](#input\_create\_security\_profile) | Set to true to create a security profile. | `bool` | `false` | no |
| <a name="input_create_security_profile_group"></a> [create\_security\_profile\_group](#input\_create\_security\_profile\_group) | Set to true to create a security profile group. | `bool` | `false` | no |
| <a name="input_custom_intercept_profile"></a> [custom\_intercept\_profile](#input\_custom\_intercept\_profile) | Configuration for the custom intercept profile. Used when security\_profile\_type is CUSTOM\_INTERCEPT. | <pre>object({<br/>    intercept_endpoint_group = string<br/>  })</pre> | `null` | no |
| <a name="input_custom_mirroring_profile"></a> [custom\_mirroring\_profile](#input\_custom\_mirroring\_profile) | Configuration for the custom mirroring profile. Used when security\_profile\_type is CUSTOM\_MIRRORING. | <pre>object({<br/>    mirroring_endpoint_group = string<br/>  })</pre> | `null` | no |
| <a name="input_existing_custom_intercept_profile_id"></a> [existing\_custom\_intercept\_profile\_id](#input\_existing\_custom\_intercept\_profile\_id) | The resource ID of an existing CUSTOM\_INTERCEPT profile to link to the group. Used if create\_security\_profile is false. | `string` | `null` | no |
| <a name="input_existing_custom_mirroring_profile_id"></a> [existing\_custom\_mirroring\_profile\_id](#input\_existing\_custom\_mirroring\_profile\_id) | The resource ID of an existing CUSTOM\_MIRRORING profile to link to the group. Used if create\_security\_profile is false. | `string` | `null` | no |
| <a name="input_existing_threat_prevention_profile_id"></a> [existing\_threat\_prevention\_profile\_id](#input\_existing\_threat\_prevention\_profile\_id) | The resource ID of an existing THREAT\_PREVENTION profile to link to the group. Used if create\_security\_profile is false. | `string` | `null` | no |
| <a name="input_link_security_profile_to_group"></a> [link\_security\_profile\_to\_group](#input\_link\_security\_profile\_to\_group) | Set to true to link the newly created security profile to the security profile group. | `bool` | `false` | no |
| <a name="input_location"></a> [location](#input\_location) | The location for the resources. Defaults to 'global'. | `string` | `"global"` | no |
| <a name="input_organization_id"></a> [organization\_id](#input\_organization\_id) | The organization ID to which the resources will be associated. | `string` | n/a | yes |
| <a name="input_security_profile_description"></a> [security\_profile\_description](#input\_security\_profile\_description) | An optional description of the security profile. | `string` | `null` | no |
| <a name="input_security_profile_group_description"></a> [security\_profile\_group\_description](#input\_security\_profile\_group\_description) | An optional description for the security profile group. | `string` | `null` | no |
| <a name="input_security_profile_group_labels"></a> [security\_profile\_group\_labels](#input\_security\_profile\_group\_labels) | A map of labels to add to the security profile group. | `map(string)` | `{}` | no |
| <a name="input_security_profile_group_name"></a> [security\_profile\_group\_name](#input\_security\_profile\_group\_name) | The name of the security profile group. | `string` | `null` | no |
| <a name="input_security_profile_labels"></a> [security\_profile\_labels](#input\_security\_profile\_labels) | A map of labels to add to the security profile. | `map(string)` | `{}` | no |
| <a name="input_security_profile_name"></a> [security\_profile\_name](#input\_security\_profile\_name) | The name of the security profile. | `string` | `null` | no |
| <a name="input_security_profile_type"></a> [security\_profile\_type](#input\_security\_profile\_type) | The type of security profile. Must be one of THREAT\_PREVENTION, CUSTOM\_MIRRORING, or CUSTOM\_INTERCEPT. | `string` | `"THREAT_PREVENTION"` | no |
| <a name="input_threat_prevention_profile"></a> [threat\_prevention\_profile](#input\_threat\_prevention\_profile) | Configuration for the threat prevention profile. Used when security\_profile\_type is THREAT\_PREVENTION. | <pre>object({<br/>    severity_overrides = optional(list(object({<br/>      severity = string<br/>      action   = string<br/>    })), [])<br/>    threat_overrides = optional(list(object({<br/>      threat_id = string<br/>      action    = string<br/>    })), [])<br/>    antivirus_overrides = optional(list(object({<br/>      protocol = string<br/>      action   = string<br/>    })), [])<br/>  })</pre> | `null` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_security_profile_group_id"></a> [security\_profile\_group\_id](#output\_security\_profile\_group\_id) | The full resource ID of the created security profile group. |
| <a name="output_security_profile_group_name"></a> [security\_profile\_group\_name](#output\_security\_profile\_group\_name) | The name of the created security profile group. |
| <a name="output_security_profile_id"></a> [security\_profile\_id](#output\_security\_profile\_id) | The full resource ID of the created security profile. |
| <a name="output_security_profile_name"></a> [security\_profile\_name](#output\_security\_profile\_name) | The name of the created security profile. |
<!-- END_TF_DOCS -->