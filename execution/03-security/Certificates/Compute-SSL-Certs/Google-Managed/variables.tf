# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable "project_id" {
  description = "The ID of the Google Cloud project where the SSL certificate will be created."
  type        = string
}

variable "ssl_certificate_name" {
  description = "Name of the SSL certificate resource. Must be 1-63 characters long, and comply with RFC1035."
  type        = string
}

variable "ssl_certificate_description" {
  description = "(Optional) An optional description of this SSL certificate resource."
  type        = string
  default     = "Terraform managed SSL Certificate"
}

variable "ssl_certificate_id" {
  description = "(Optional) The unique identifier for the SSL certificate resource. If not provided, a new one will be created."
  type        = number
  default     = null
}

variable "ssl_certificate_type" {
  description = "(Optional) Type of the certificate. Defaults to 'MANAGED'."
  type        = string
  default     = "MANAGED"
  validation {
    condition     = var.ssl_certificate_type == "MANAGED"
    error_message = "The type must be 'MANAGED'."
  }
}

variable "ssl_managed_domains" {
  description = "Configuration for the managed SSL certificate, primarily the list of domains."
  type = set(object(
    {
      domains = list(string)
    }
  ))
}

variable "ssl_timeouts" {
  description = "(Optional) Timeouts for creating and deleting the SSL certificate resource."
  type = set(object(
    {
      create = string
      delete = string
    }
  ))
  default = []
}