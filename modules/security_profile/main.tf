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

resource "google_network_security_security_profile" "security_profile" {
  count = var.create_security_profile ? 1 : 0

  name        = var.security_profile_name
  parent      = "organizations/${var.organization_id}"
  type        = var.security_profile_type
  location    = var.location
  description = var.security_profile_description
  labels      = var.security_profile_labels

  dynamic "threat_prevention_profile" {
    for_each = var.security_profile_type == "THREAT_PREVENTION" && var.threat_prevention_profile != null ? [var.threat_prevention_profile] : []

    content {
      dynamic "severity_overrides" {
        for_each = lookup(threat_prevention_profile.value, "severity_overrides", [])
        content {
          severity = severity_overrides.value.severity
          action   = severity_overrides.value.action
        }
      }
      dynamic "threat_overrides" {
        for_each = lookup(threat_prevention_profile.value, "threat_overrides", [])
        content {
          threat_id = threat_overrides.value.threat_id
          action    = threat_overrides.value.action
        }
      }
      dynamic "antivirus_overrides" {
        for_each = lookup(threat_prevention_profile.value, "antivirus_overrides", [])
        content {
          protocol = antivirus_overrides.value.protocol
          action   = antivirus_overrides.value.action
        }
      }
    }
  }

  dynamic "custom_mirroring_profile" {
    for_each = var.security_profile_type == "CUSTOM_MIRRORING" && var.custom_mirroring_profile != null ? [var.custom_mirroring_profile] : []
    content {
      mirroring_endpoint_group = custom_mirroring_profile.value.mirroring_endpoint_group
    }
  }

  dynamic "custom_intercept_profile" {
    for_each = var.security_profile_type == "CUSTOM_INTERCEPT" && var.custom_intercept_profile != null ? [var.custom_intercept_profile] : []
    content {
      intercept_endpoint_group = custom_intercept_profile.value.intercept_endpoint_group
    }
  }
}

resource "google_network_security_security_profile_group" "security_profile_group" {
  count = var.create_security_profile_group ? 1 : 0

  name        = var.security_profile_group_name
  parent      = "organizations/${var.organization_id}"
  location    = var.location
  description = var.security_profile_group_description
  labels      = var.security_profile_group_labels

  threat_prevention_profile = local.threat_prevention_profile_link
  custom_mirroring_profile  = local.custom_mirroring_profile_link
  custom_intercept_profile  = local.custom_intercept_profile_link
}