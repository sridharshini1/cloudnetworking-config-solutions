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

locals {
  should_link_new_profile = var.link_security_profile_to_group && var.create_security_profile
  new_profile_id          = var.create_security_profile ? one(google_network_security_security_profile.security_profile).id : null
  threat_prevention_profile_link = (
    local.should_link_new_profile && var.security_profile_type == "THREAT_PREVENTION"
    ? local.new_profile_id
    : var.existing_threat_prevention_profile_id
  )
  custom_mirroring_profile_link = (
    local.should_link_new_profile && var.security_profile_type == "CUSTOM_MIRRORING"
    ? local.new_profile_id
    : var.existing_custom_mirroring_profile_id
  )
  custom_intercept_profile_link = (
    local.should_link_new_profile && var.security_profile_type == "CUSTOM_INTERCEPT"
    ? local.new_profile_id
    : var.existing_custom_intercept_profile_id
  )
}