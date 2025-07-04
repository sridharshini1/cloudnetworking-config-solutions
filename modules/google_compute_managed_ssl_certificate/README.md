Copyright 2025 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | n/a |

## Resources

| Name | Type |
|------|------|
| [google_compute_managed_ssl_certificate.ssl_cert](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_managed_ssl_certificate) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_certificate_id"></a> [certificate\_id](#input\_certificate\_id) | (Optional) The unique identifier for the SSL certificate resource, assigned by Google Cloud. This is an output attribute and typically not set by the user when creating a new certificate. If you are trying to import an existing certificate, other mechanisms are generally used. | `number` | `null` | no |
| <a name="input_description"></a> [description](#input\_description) | (Optional) A human-readable description for this SSL certificate resource. This can be useful for annotating the certificate's purpose or management details. | `string` | `null` | no |
| <a name="input_managed"></a> [managed](#input\_managed) | (Required for MANAGED type certificates) Configuration block for a Google-managed SSL certificate. This block is necessary when `type` is set to "MANAGED". It primarily specifies the domain names that the certificate will secure. You can specify multiple domains for a single certificate. | <pre>set(object(<br>    {<br>      domains = list(string) # List of fully qualified domain names (e.g., ['example.com', 'www.example.com']).<br>    }<br>  ))</pre> | `[]` | no |
| <a name="input_name"></a> [name](#input\_name) | (Optional, but Recommended for new resources) The user-defined name for the SSL certificate resource. This name must be 1-63 characters long and comply with RFC1035. Specifically, it must match the regular expression '[a-z]([-a-z0-9]*[a-z0-9])?', meaning the first character must be a lowercase letter, and all subsequent characters must be a dash, lowercase letter, or digit, except for the last character, which cannot be a dash. If not provided, a name may be generated. SSL certificate names are unique within a project and are in the same namespace as other Google Cloud SSL certificates. | `string` | `null` | no |
| <a name="input_project"></a> [project](#input\_project) | (Optional) The ID of the Google Cloud project in which the SSL certificate will be created. If not provided, the project will be inferred from the Google provider configuration. | `string` | `null` | no |
| <a name="input_timeouts"></a> [timeouts](#input\_timeouts) | (Optional) A block configuring timeouts for the create and delete operations of the SSL certificate resource. Allows customization of how long Terraform will wait for these actions to complete. | <pre>set(object(<br>    {<br>      create = string # (Optional) How long to wait for the certificate to be created (e.g., "30m").<br>      delete = string # (Optional) How long to wait for the certificate to be deleted (e.g., "10m").<br>    }<br>  ))</pre> | `[]` | no |
| <a name="input_type"></a> [type](#input\_type) | (Optional) The type of SSL certificate. For this module, it defaults to and primarily supports 'MANAGED', indicating that Google manages the certificate provisioning and renewal. Possible values defined by Google Cloud are ["MANAGED", "SELF\_MANAGED"], though this module is tailored for 'MANAGED'. | `string` | `"MANAGED"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_certificate_id"></a> [certificate\_id](#output\_certificate\_id) | The unique numerical identifier for the managed SSL certificate resource, assigned by Google Cloud. |
| <a name="output_creation_timestamp"></a> [creation\_timestamp](#output\_creation\_timestamp) | The timestamp in RFC3339 format indicating when this SSL certificate resource was created. |
| <a name="output_expire_time"></a> [expire\_time](#output\_expire\_time) | The timestamp in RFC3339 format indicating when this SSL certificate will expire. For 'MANAGED' type certificates, Google automatically renews the certificate before this time. |
| <a name="output_id"></a> [id](#output\_id) | The fully qualified identifier (ID) of the managed SSL certificate resource, typically in the format 'projects/PROJECT\_ID/global/sslCertificates/CERTIFICATE\_NAME'. |
| <a name="output_project"></a> [project](#output\_project) | The ID of the Google Cloud project in which the managed SSL certificate was created. |
| <a name="output_self_link"></a> [self\_link](#output\_self\_link) | The self-referential URI of the created managed SSL certificate resource. |
| <a name="output_ssl_cert"></a> [ssl\_cert](#output\_ssl\_cert) | All attributes of the created `google_compute_managed_ssl_certificate` resource. This output provides the full object, allowing access to any of its properties. |
| <a name="output_subject_alternative_names"></a> [subject\_alternative\_names](#output\_subject\_alternative\_names) | A list of Subject Alternative Names (SANs) that are secured by this SSL certificate. These usually correspond to the domains provided in the 'managed' block. |