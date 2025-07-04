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

output "security_profile_id" {
  description = "The full resource ID of the created security profile."
  value       = var.create_security_profile ? one(google_network_security_security_profile.security_profile).id : null
}

output "security_profile_name" {
  description = "The name of the created security profile."
  value       = var.create_security_profile ? one(google_network_security_security_profile.security_profile).name : null
}

output "security_profile_group_id" {
  description = "The full resource ID of the created security profile group."
  value       = var.create_security_profile_group ? one(google_network_security_security_profile_group.security_profile_group).id : null
}

output "security_profile_group_name" {
  description = "The name of the created security profile group."
  value       = var.create_security_profile_group ? one(google_network_security_security_profile_group.security_profile_group).name : null
}