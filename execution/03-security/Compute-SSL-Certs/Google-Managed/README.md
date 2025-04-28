# Google Cloud Managed SSL Certificate

## Overview

This Terraform module facilitates the creation and management of Google Cloud Managed SSL Certificates. Google Cloud Managed SSL Certificates are SSL/TLS certificates that you can provision, deploy, and manage for your domains. These certificates are obtained and renewed automatically by Google, simplifying the lifecycle management of your SSL/TLS configurations.

Managed SSL certificates are essential for enabling HTTPS traffic for your applications served via Google Cloud External HTTP(S) Load Balancers. By using Google-managed certificates, you delegate the complexities of certificate issuance and renewal to Google, ensuring your applications remain secure and accessible over HTTPS.

### Key Features of Google Cloud Managed SSL Certificates:

* **Automated Provisioning and Renewal:** Google handles the entire lifecycle, from initial issuance to timely renewal, reducing administrative overhead.
* **Multiple Domain (SAN) Support:** Secure multiple domain names (Subject Alternative Names) with a single certificate.
* **Strong Encryption:** Leverages Google's robust infrastructure for secure key management and certificate operations.
* **Integration with Google Cloud Load Balancing:** Seamlessly integrates with External HTTP(S) Load Balancers and Target SSL Proxies.
* **No Additional Cost:** Google-managed SSL certificates are provided at no additional charge.

### Public Documentation References:

* **Google Cloud SSL Certificates Overview:** [https://cloud.google.com/load-balancing/docs/ssl-certificates/managed-ssl-certificates](https://cloud.google.com/load-balancing/docs/ssl-certificates/managed-ssl-certificates)
* **Terraform `google_compute_managed_ssl_certificate` Resource:** [https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_managed_ssl_certificate](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_managed_ssl_certificate)
* **Using Google-managed SSL certificates:** [https://cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs](https://cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs)

## Purpose

This implementaion simplifies the process of defining and deploying `google_compute_managed_ssl_certificate` resources using Terraform. It abstracts the resource configuration into reusable code, promoting consistency and best practices for managing SSL certificates within your Google Cloud environment.

## Prerequisites

Before using this stage, ensure you have the following:

1.  **Google Cloud Project:** A Google Cloud Project with billing enabled.
2.  **Terraform:** Terraform (version 1.0.0 or later recommended) installed on your local machine or CI/CD environment.
3.  **Google Cloud SDK (gcloud):** Authenticated with appropriate permissions to manage Compute Engine resources, specifically SSL certificates. Alternatively, a service account key with the necessary roles (e.g., "Compute Admin" - `roles/compute.admin`, or more granularly "Compute Security Admin" - `roles/compute.securityAdmin` and "Compute Network Admin" - `roles/compute.networkAdmin` if you are also managing related Load Balancing components).
    * Ensure the service account or user has permissions like `compute.managedSslCertificates.create`, `compute.managedSslCertificates.get`, `compute.managedSslCertificates.list`, and `compute.managedSslCertificates.delete`.
4.  **Domain Ownership:** You must own or control the domain names you intend to secure. Google will perform domain validation to confirm your control before the certificate can be successfully provisioned and activated. This typically involves ensuring your domain's DNS records point to the IP address of the Google Cloud Load Balancer that will use this certificate.
5.  **Configured Load Balancer (Recommended):** While this module provisions the certificate, the certificate itself becomes fully active and useful once associated with a Google Cloud External HTTP(S) Load Balancer's Target Proxy. Ensure your DNS A/AAAA records for the specified domains point to the load balancer's IP address.

## Usage and Execution Steps

Follow these steps to use this Terraform module to create a Google Cloud Managed SSL Certificate:

1. **Create your configuration .tfvars files:**

    * Create `ssl.tfvars` file defining the values for ingress rules and egress rules. Ensure these files are stored in the `configuration/security/compute_managed_ssl` folder.

    * For reference on how to structure your `ssl.tfvars` file , refer to sample `terraform.tfvars.example` file . Each field and its structure is described in the [input section](#inputs) below.


2. **Initialize Terraform:**

    * Run the following command to initialize Terraform:

    ```
    terraform init
    ```

3. **Review the Execution Plan:**

    * Use the terraform plan command to generate an execution plan. This will show you the changes Terraform will make to your Google Cloud infrastructure:

    ```
    terraform plan
    ```

Carefully review the plan to ensure it aligns with your intended configuration.

4. **Apply the Configuration:**

    Once you're satisfied with the plan, execute the terraform apply command to provision your Cloud SQL instances:

    ```
    terraform apply -var-file="../../../configuration/security/compute_ssl_certs/google_managed/google_managed_ssl.tfvars"
    ```

<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_ssl_certificate"></a> [ssl\_certificate](#module\_ssl\_certificate) | ../../../../modules/google_compute_managed_ssl_certificate | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The ID of the Google Cloud project where the SSL certificate will be created. | `string` | n/a | yes |
| <a name="input_ssl_certificate_description"></a> [ssl\_certificate\_description](#input\_ssl\_certificate\_description) | (Optional) An optional description of this SSL certificate resource. | `string` | `"Terraform managed SSL Certificate"` | no |
| <a name="input_ssl_certificate_id"></a> [ssl\_certificate\_id](#input\_ssl\_certificate\_id) | (Optional) The unique identifier for the SSL certificate resource. If not provided, a new one will be created. | `number` | `null` | no |
| <a name="input_ssl_certificate_name"></a> [ssl\_certificate\_name](#input\_ssl\_certificate\_name) | Name of the SSL certificate resource. Must be 1-63 characters long, and comply with RFC1035. | `string` | n/a | yes |
| <a name="input_ssl_certificate_type"></a> [ssl\_certificate\_type](#input\_ssl\_certificate\_type) | (Optional) Type of the certificate. Defaults to 'MANAGED'. | `string` | `"MANAGED"` | no |
| <a name="input_ssl_managed_domains"></a> [ssl\_managed\_domains](#input\_ssl\_managed\_domains) | Configuration for the managed SSL certificate, primarily the list of domains. | <pre>set(object(<br>    {<br>      domains = list(string)<br>    }<br>  ))</pre> | n/a | yes |
| <a name="input_ssl_timeouts"></a> [ssl\_timeouts](#input\_ssl\_timeouts) | (Optional) Timeouts for creating and deleting the SSL certificate resource. | <pre>set(object(<br>    {<br>      create = string<br>      delete = string<br>    }<br>  ))</pre> | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_managed_ssl_certificate_creation_timestamp"></a> [managed\_ssl\_certificate\_creation\_timestamp](#output\_managed\_ssl\_certificate\_creation\_timestamp) | Creation timestamp of the managed SSL certificate. |
| <a name="output_managed_ssl_certificate_details"></a> [managed\_ssl\_certificate\_details](#output\_managed\_ssl\_certificate\_details) | All attributes of the created google\_compute\_managed\_ssl\_certificate. |
| <a name="output_managed_ssl_certificate_expire_time"></a> [managed\_ssl\_certificate\_expire\_time](#output\_managed\_ssl\_certificate\_expire\_time) | Expire time of the managed SSL certificate. |
| <a name="output_managed_ssl_certificate_full_id"></a> [managed\_ssl\_certificate\_full\_id](#output\_managed\_ssl\_certificate\_full\_id) | The full ID of the managed SSL certificate. |
| <a name="output_managed_ssl_certificate_id"></a> [managed\_ssl\_certificate\_id](#output\_managed\_ssl\_certificate\_id) | The unique identifier for the managed SSL certificate resource. |
| <a name="output_managed_ssl_certificate_project"></a> [managed\_ssl\_certificate\_project](#output\_managed\_ssl\_certificate\_project) | The project in which the managed SSL certificate was created. |
| <a name="output_managed_ssl_certificate_self_link"></a> [managed\_ssl\_certificate\_self\_link](#output\_managed\_ssl\_certificate\_self\_link) | The self link of the managed SSL certificate. |
| <a name="output_managed_ssl_certificate_subject_alternative_names"></a> [managed\_ssl\_certificate\_subject\_alternative\_names](#output\_managed\_ssl\_certificate\_subject\_alternative\_names) | Subject Alternative Names (SANs) of the managed SSL certificate. |

<!-- END_TF_DOCS -->