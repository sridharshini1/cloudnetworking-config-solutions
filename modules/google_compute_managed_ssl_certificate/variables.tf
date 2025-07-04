/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

variable "certificate_id" {
  description = "(Optional) The unique identifier for the SSL certificate resource, assigned by Google Cloud. This is an output attribute and typically not set by the user when creating a new certificate. If you are trying to import an existing certificate, other mechanisms are generally used."
  type        = number
  default     = null
}

variable "description" {
  description = "(Optional) A human-readable description for this SSL certificate resource. This can be useful for annotating the certificate's purpose or management details."
  type        = string
  default     = null
}

variable "name" {
  description = "(Optional, but Recommended for new resources) The user-defined name for the SSL certificate resource. This name must be 1-63 characters long and comply with RFC1035. Specifically, it must match the regular expression '[a-z]([-a-z0-9]*[a-z0-9])?', meaning the first character must be a lowercase letter, and all subsequent characters must be a dash, lowercase letter, or digit, except for the last character, which cannot be a dash. If not provided, a name may be generated. SSL certificate names are unique within a project and are in the same namespace as other Google Cloud SSL certificates."
  type        = string
  default     = null
}

variable "project" {
  description = "(Optional) The ID of the Google Cloud project in which the SSL certificate will be created. If not provided, the project will be inferred from the Google provider configuration."
  type        = string
  default     = null
}

variable "type" {
  description = "(Optional) The type of SSL certificate. For this module, it defaults to and primarily supports 'MANAGED', indicating that Google manages the certificate provisioning and renewal. Possible values defined by Google Cloud are [\"MANAGED\", \"SELF_MANAGED\"], though this module is tailored for 'MANAGED'."
  type        = string
  default     = "MANAGED" # Explicitly setting default as per typical usage.
}

variable "managed" {
  description = "(Required for MANAGED type certificates) Configuration block for a Google-managed SSL certificate. This block is necessary when `type` is set to \"MANAGED\". It primarily specifies the domain names that the certificate will secure. You can specify multiple domains for a single certificate."
  type = set(object(
    {
      domains = list(string) # List of fully qualified domain names (e.g., ['example.com', 'www.example.com']).
    }
  ))
  default = [] # Default is empty, but this block will be required by GCP if type is MANAGED.
}

variable "timeouts" {
  description = "(Optional) A block configuring timeouts for the create and delete operations of the SSL certificate resource. Allows customization of how long Terraform will wait for these actions to complete."
  type = set(object(
    {
      create = string # (Optional) How long to wait for the certificate to be created (e.g., "30m").
      delete = string # (Optional) How long to wait for the certificate to be deleted (e.g., "10m").
    }
  ))
  default = []
}