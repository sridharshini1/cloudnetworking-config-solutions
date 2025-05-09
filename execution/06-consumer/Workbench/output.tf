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

output "workbench_instance_ids" {
  value       = { for k, v in module.workbench_instance : k => v.id }
  description = "The IDs of the created Workbench instances."
}

output "workbench_instance_proxy_uris" {
  value = {
    for k, v in module.workbench_instance :
    k => (v.proxy_uri != null && v.proxy_uri != "" ? v.proxy_uri : null)
  }
  description = "The proxy URIs of the created Workbench instances. Proxy Access is null when disable proxy access is set to true or proxy_uri is not available."
}