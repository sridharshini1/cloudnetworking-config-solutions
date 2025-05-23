
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
variable "attachments" {
  description = "Ids of the resources to which this policy will be attached, in descriptive name => self link format. Specify folders or organization for hierarchical policy, VPCs for network policy."
  type        = map(string)
  default     = {}
  nullable    = false
}

variable "description" {
  description = "An optional description of this resource. Provide this property when you create the resource."
  type        = string
  default     = null
}

variable "egress_rules" {
  description = "List of egress rule definitions, action can be 'allow', 'deny', 'goto_next' or 'apply_security_profile_group'. The match.layer4configs map is in protocol => optional [ports] format."
  type = map(object({
    priority                = number
    action                  = optional(string, "deny")
    description             = optional(string)
    disabled                = optional(bool, false)
    enable_logging          = optional(bool)
    security_profile_group  = optional(string)
    target_resources        = optional(list(string))
    target_service_accounts = optional(list(string))
    target_tags             = optional(list(string))
    tls_inspect             = optional(bool, null)
    match = object({
      address_groups       = optional(list(string))
      fqdns                = optional(list(string))
      region_codes         = optional(list(string))
      threat_intelligences = optional(list(string))
      destination_ranges   = optional(list(string))
      source_ranges        = optional(list(string))
      source_tags          = optional(list(string))
      layer4_configs = optional(list(object({
        protocol = optional(string, "all")
        ports    = optional(list(string))
      })), [{}])
    })
  }))
  default  = {}
  nullable = false
  validation {
    condition = alltrue([
      for k, v in var.egress_rules :
      contains(["allow", "deny", "goto_next", "apply_security_profile_group"], v.action)
    ])
    error_message = "Action can only be one of 'allow', 'deny', 'goto_next' or 'apply_security_profile_group'."
  }
}

variable "ingress_rules" {
  description = "List of ingress rule definitions, action can be 'allow', 'deny', 'goto_next' or 'apply_security_profile_group'."
  type = map(object({
    priority                = number
    action                  = optional(string, "allow")
    description             = optional(string)
    disabled                = optional(bool, false)
    enable_logging          = optional(bool)
    security_profile_group  = optional(string)
    target_resources        = optional(list(string))
    target_service_accounts = optional(list(string))
    target_tags             = optional(list(string))
    tls_inspect             = optional(bool, null)
    match = object({
      address_groups       = optional(list(string))
      fqdns                = optional(list(string))
      region_codes         = optional(list(string))
      threat_intelligences = optional(list(string))
      destination_ranges   = optional(list(string))
      source_ranges        = optional(list(string))
      source_tags          = optional(list(string))
      layer4_configs = optional(list(object({
        protocol = optional(string, "all")
        ports    = optional(list(string))
      })), [{}])
    })
  }))
  default  = {}
  nullable = false
  validation {
    condition = alltrue([
      for k, v in var.ingress_rules :
      contains(["allow", "deny", "goto_next", "apply_security_profile_group"], v.action)
    ])
    error_message = "Action can only be one of 'allow', 'deny', 'goto_next' or 'apply_security_profile_group'."
  }
}

variable "region" {
  description = "Policy region. Leave null for hierarchical policy, set to 'global' for a global network policy."
  type        = string
  default     = null
}

variable "config_folder_path" {
  description = "Location of YAML files holding Firewall Policy configuration values."
  type        = string
  default     = "../../../../configuration/security/Firewall/FirewallPolicy/config"
}

variable "ingress_action" {
  description = "Action of the Ingress Firewall Rule."
  type        = string
  default     = "allow"
}

variable "egress_action" {
  description = "Action of the Egress Firewall Rule."
  type        = string
  default     = "deny"
}

variable "enable_logging" {
  description = "This field denotes whether to enable logging for a particular firewall rule."
  type        = bool
  default     = null
}

variable "address_groups" {
  description = "Address groups which should be matched against the traffic."
  type        = list(string)

  default = null
}

variable "security_profile_group" {
  description = "A fully-qualified URL of a SecurityProfile resource instance. Example: https://networksecurity.googleapis.com/v1/projects/{project}/locations/{location}/securityProfileGroups/my-security-profile-group"
  type        = string
  default     = null
}

variable "disabled" {
  description = "Denotes whether the firewall rule is disabled, i.e not applied to the network it is associated with. When set to true, the firewall rule is not enforced and the network behaves as if it did not exist. If this is unspecified, the firewall rule will be enabled."
  type        = bool
  default     = false
}

variable "target_resources" {
  description = "A list of network resource URLs to which this rule applies. This field allows you to control which network's VMs get this rule. If this field is left blank, all VMs within the organization will receive the rule."
  type        = list(string)
  default     = null
}

variable "target_service_accounts" {
  description = "(Optional) A list of service accounts indicating the sets of instances that are applied with this rule."
  type        = list(string)
  default     = null
}

variable "target_tags" {
  description = "A list of instance tags indicating sets of instances located in the network that may make network connections as specified in allowed[]. If no targetTags are specified, the firewall rule applies to all instances on the specified network."
  type        = list(string)
  default     = null
}

variable "tls_inspect" {
  description = "(Optional) Boolean flag indicating if the traffic should be TLS decrypted. Can be set only if action = 'apply_security_profile_group' and cannot be set for other actions."
  type        = bool
  default     = null
}

variable "fqdns" {
  description = "Fully Qualified Domain Name (FQDN) which should be matched against traffic"
  type        = list(string)
  default     = null
}

variable "region_codes" {
  description = "Region codes whose IP addresses will be used to match for traffic."
  type        = list(string)
  default     = null
}

variable "threat_intelligences" {
  description = "Names of Network Threat Intelligence lists. The IPs in these lists will be matched against traffic destination."
  type        = list(string)
  default     = null
}

variable "destination_ranges" {
  description = "If destination ranges are specified, the firewall will apply only to traffic that has destination IP address in these ranges."
  type        = list(string)
  default     = null
}

variable "source_ranges" {
  description = "If source ranges are specified, the firewall will apply only to traffic that has source IP address in these ranges. These ranges must be expressed in CIDR format."
  type        = list(string)
  default     = null
}

variable "source_tags" {
  description = "A list of source tags."
  type        = list(string)
  default     = null
}

variable "layer4_configs" {
  description = "Pairs of IP protocols and ports that the rule should match. "
  type = list(object({
    protocol = optional(string, "all")
    ports    = optional(list(string))
  }))
  default = null
}
