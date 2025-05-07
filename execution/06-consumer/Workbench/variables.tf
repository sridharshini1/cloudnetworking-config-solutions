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

variable "region" {
  type        = string
  description = "The region where resources will be created."
  default     = "us-central1"
}

variable "location" {
  type        = string
  description = "The zone where resources will be created."
  default     = "us-central1-a"
}

variable "machine_type" {
  type        = string
  description = "Default machine type for the Workbench instance."
  default     = "e2-standard-2"
}

variable "service_accounts" {
  type = list(object({
    email  = string
    scopes = list(string)
  }))
  description = "List of service accounts with their scopes."
  default     = []
}

variable "network_tags" {
  type        = list(string)
  description = "Default network tags for the Workbench instance."
  default     = []
}

variable "data_disk_size" {
  type        = number
  description = "Default disk size in GB for data disk."
  default     = 100
}

variable "data_disk_type" {
  type        = string
  description = "Default data disk type (e.g., PD_STANDARD, PD_SSD, PD_BALANCED)."
  default     = "PD_SSD"
}

variable "boot_disk_size_gb" {
  type        = number
  description = "Default boot disk size in GB."
  default     = 200
}

variable "boot_disk_type" {
  type        = string
  description = "Default boot disk type (e.g., PD_STANDARD, PD_SSD, PD_BALANCED)."
  default     = "PD_STANDARD"
}

variable "nic_type" {
  type        = string
  description = "Default NIC type for the Workbench instance (e.g., GVNIC, VIRTIO_NET)."
  default     = "GVNIC"
}

variable "vm_image_project" {
  type        = string
  description = "Default source project of the VM image for Workbench instances."
  default     = "cloud-notebooks-managed" # Replace with your desired default project
}

variable "vm_image_family" {
  type        = string
  description = "Default VM image OS family for Workbench instances."
  default     = "workbench-instances" # Replace with your desired default family
}

variable "vm_image_name" {
  type        = string
  description = "Default VM image name for Workbench instances. If null, the latest image in the family will be used."
  default     = null
}

variable "labels" {
  type        = map(string)
  description = "Default labels for Workbench instance."
  default = {
    owner = "your-desired-owner"
  }
}

variable "disable_public_ip_default" {
  type        = bool
  description = "Default setting for disabling public IPs. True means no public IP by default."
  default     = true
}

variable "disable_proxy_access_default" {
  type        = bool
  description = "Default setting for disabling proxy access. True means proxy access is disabled by default."
  default     = true
}

variable "disk_encryption_default" {
  type        = string
  description = "Default disk encryption key for the Workbench instance."
  default     = null
}

variable "enable_secure_boot_default" {
  type        = bool
  description = "Default setting for enabling secure boot."
  default     = false
}

variable "enable_vtpm_default" {
  type        = bool
  description = "Default setting for enabling vTPM."
  default     = false
}

variable "enable_integrity_monitoring_default" {
  type        = bool
  description = "Default setting for enabling integrity monitoring."
  default     = false
}

variable "config_folder_path" {
  description = "Location of YAML files holding Workbench configuration values."
  type        = string
  default     = "../../../configuration/consumer/Workbench/config"
}

variable "internal_ip_only" {
  description = "Specifies whether the Workbench instance should use only internal IP addresses."
  type        = bool
  default     = true
}

variable "metadata" {
  description = "Custom metadata to apply to this instance"
  type        = map(string)
  default     = {}
}

## variables for metadata setting https://cloud.google.com/vertex-ai/docs/workbench/instances/manage-metadata
variable "metadata_configs" {
  description = "predefined metadata to apply to this instance"
  type = object({
    idle-timeout-seconds            = optional(number)
    notebook-upgrade-schedule       = optional(string)
    notebook-disable-downloads      = optional(bool)
    notebook-disable-root           = optional(bool)
    post-startup-script             = optional(string)
    post-startup-script-behavior    = optional(string)
    nbconvert                       = optional(bool)
    notebook-enable-delete-to-trash = optional(bool)
    disable-mixer                   = optional(bool)
    jupyter-user                    = optional(string)
    report-event-health             = optional(bool)
  })
  default = {}
}

variable "instance_owners" {
  type        = list(string)
  description = "List of email addresses of users who will have owner permissions on the Workbench instance."
  default     = []
}