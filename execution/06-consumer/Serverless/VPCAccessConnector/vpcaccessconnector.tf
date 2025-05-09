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

module "serverless-connector" {
  source     = "terraform-google-modules/network/google//modules/vpc-serverless-connector-beta"
  for_each   = { for vpcsc in local.instance_list : vpcsc.name => vpcsc }
  project_id = each.value.project_id
  vpc_connectors = [{
    name            = each.value.name
    region          = each.value.region
    network         = each.value.network
    ip_cidr_range   = each.value.ip_cidr_range
    subnet_name     = each.value.subnet_name
    host_project_id = each.value.host_project_id
    min_instances   = each.value.min_instances
    max_instances   = each.value.max_instances
    max_throughput  = each.value.max_throughput
    machine_type    = each.value.machine_type
  }]
}
