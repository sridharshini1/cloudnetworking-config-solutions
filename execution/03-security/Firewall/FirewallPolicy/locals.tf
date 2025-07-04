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

locals {
  config_folder_path = var.config_folder_path
  instance_map       = { for instance in local.instance_list : instance.name => instance }
  instances          = [for file in fileset(local.config_folder_path, "[^_]*.yaml") : yamldecode(file("${local.config_folder_path}/${file}"))]
  instance_list = flatten([
    for instance in try(local.instances, []) : {
      name        = instance.name
      parent_id   = instance.parent_id
      attachments = try(instance.attachments, var.attachments)
      description = try(instance.description, var.description)
      region      = try(instance.region, var.region)
      ingress_rules = {
        for k, v in try(instance.ingress_rules, var.ingress_rules) : "ingress/${k}" => {
          direction               = "INGRESS" // This can only be INGRESS
          name                    = k
          priority                = v.priority
          action                  = lookup(v, "action", var.ingress_action)
          description             = lookup(v, "description", var.description)
          disabled                = lookup(v, "disabled", var.disabled)
          enable_logging          = lookup(v, "enable_logging", var.enable_logging)
          security_profile_group  = lookup(v, "security_profile_group", var.security_profile_group)
          target_resources        = lookup(v, "target_resources", var.target_resources)
          target_service_accounts = lookup(v, "target_service_accounts", var.target_service_accounts)
          target_tags             = lookup(v, "target_tags", var.target_tags)
          tls_inspect             = lookup(v, "tls_inspect", var.tls_inspect)
          match = {
            address_groups       = lookup(v.match, "address_groups", var.address_groups)
            fqdns                = lookup(v.match, "fqdns", var.fqdns)
            region_codes         = lookup(v.match, "region_codes", var.region_codes)
            threat_intelligences = lookup(v.match, "threat_intelligences", var.threat_intelligences)
            destination_ranges   = lookup(v.match, "destination_ranges", var.destination_ranges)
            source_ranges        = lookup(v.match, "source_ranges", var.source_ranges)
            source_tags          = lookup(v.match, "source_tags", var.source_tags)
            layer4_configs = (
              lookup(v.match, "layer4_configs", var.layer4_configs) == null
              ? [{ protocol = "all", ports = null }]
              : [
                for c in v.match.layer4_configs :
                merge({ protocol = "all", ports = [] }, c)
              ]
            )

          }
        }

      }

      egress_rules = {
        for k, v in try(instance.egress_rules, var.egress_rules) : "egress/${k}" => {
          direction               = "EGRESS" // This can only be EGRESS
          name                    = k
          priority                = v.priority
          action                  = lookup(v, "action", var.egress_action)
          description             = lookup(v, "description", var.description)
          disabled                = lookup(v, "disabled", var.disabled)
          enable_logging          = lookup(v, "enable_logging", var.enable_logging)
          security_profile_group  = lookup(v, "security_profile_group", var.security_profile_group)
          target_resources        = lookup(v, "target_resources", var.target_resources)
          target_service_accounts = lookup(v, "target_service_accounts", var.target_service_accounts)
          target_tags             = lookup(v, "target_tags", var.target_tags)
          tls_inspect             = lookup(v, "tls_inspect", var.tls_inspect)
          match = {
            address_groups       = lookup(v.match, "address_groups", var.address_groups)
            fqdns                = lookup(v.match, "fqdns", var.fqdns)
            region_codes         = lookup(v.match, "region_codes", var.region_codes)
            threat_intelligences = lookup(v.match, "threat_intelligences", var.threat_intelligences)
            destination_ranges   = lookup(v.match, "destination_ranges", var.destination_ranges)
            source_ranges        = lookup(v.match, "source_ranges", var.source_ranges)
            source_tags          = lookup(v.match, "source_tags", var.source_tags)
            layer4_configs = (
              lookup(v.match, "layer4_configs", var.layer4_configs) == null
              ? [{ protocol = "all", ports = null }]
              : [
                for c in v.match.layer4_configs :
                merge({ protocol = "all", ports = [] }, c)
              ]
            )
          }
        }
      }
    }
  ])
}
