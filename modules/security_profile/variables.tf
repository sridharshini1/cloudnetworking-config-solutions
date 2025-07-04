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

variable "organization_id" {
  description = "The organization ID to which the resources will be associated."
  type        = string
}

variable "location" {
  description = "The location for the resources. Defaults to 'global'."
  type        = string
  default     = "global"
}

variable "create_security_profile" {
  description = "Set to true to create a security profile."
  type        = bool
  default     = false
}

variable "security_profile_name" {
  description = "The name of the security profile."
  type        = string
  default     = null
}

variable "security_profile_type" {
  description = "The type of security profile. Must be one of THREAT_PREVENTION, CUSTOM_MIRRORING, or CUSTOM_INTERCEPT."
  type        = string
  default     = "THREAT_PREVENTION"
  validation {
    condition     = contains(["THREAT_PREVENTION", "CUSTOM_MIRRORING", "CUSTOM_INTERCEPT"], var.security_profile_type)
    error_message = "Valid values for security_profile_type are THREAT_PREVENTION, CUSTOM_MIRRORING, or CUSTOM_INTERCEPT."
  }
}

variable "security_profile_description" {
  description = "An optional description of the security profile."
  type        = string
  default     = null
}

variable "security_profile_labels" {
  description = "A map of labels to add to the security profile."
  type        = map(string)
  default     = {}
}

variable "threat_prevention_profile" {
  description = "Configuration for the threat prevention profile. Used when security_profile_type is THREAT_PREVENTION."
  type = object({
    severity_overrides = optional(list(object({
      severity = string
      action   = string
    })), [])
    threat_overrides = optional(list(object({
      threat_id = string
      action    = string
    })), [])
    antivirus_overrides = optional(list(object({
      protocol = string
      action   = string
    })), [])
  })
  default = null
}

variable "custom_mirroring_profile" {
  description = "Configuration for the custom mirroring profile. Used when security_profile_type is CUSTOM_MIRRORING."
  type = object({
    mirroring_endpoint_group = string
  })
  default = null
}

variable "custom_intercept_profile" {
  description = "Configuration for the custom intercept profile. Used when security_profile_type is CUSTOM_INTERCEPT."
  type = object({
    intercept_endpoint_group = string
  })
  default = null
}

variable "create_security_profile_group" {
  description = "Set to true to create a security profile group."
  type        = bool
  default     = false
}

variable "security_profile_group_name" {
  description = "The name of the security profile group."
  type        = string
  default     = null
}

variable "security_profile_group_description" {
  description = "An optional description for the security profile group."
  type        = string
  default     = null
}

variable "security_profile_group_labels" {
  description = "A map of labels to add to the security profile group."
  type        = map(string)
  default     = {}
}

variable "link_security_profile_to_group" {
  description = "Set to true to link the newly created security profile to the security profile group."
  type        = bool
  default     = false
}

variable "existing_threat_prevention_profile_id" {
  description = "The resource ID of an existing THREAT_PREVENTION profile to link to the group. Used if create_security_profile is false."
  type        = string
  default     = null
}

variable "existing_custom_mirroring_profile_id" {
  description = "The resource ID of an existing CUSTOM_MIRRORING profile to link to the group. Used if create_security_profile is false."
  type        = string
  default     = null
}

variable "existing_custom_intercept_profile_id" {
  description = "The resource ID of an existing CUSTOM_INTERCEPT profile to link to the group. Used if create_security_profile is false."
  type        = string
  default     = null
}