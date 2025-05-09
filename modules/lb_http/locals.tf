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
  address                 = var.create_address ? join("", google_compute_global_address.default[*].address) : var.address
  ipv6_address            = var.create_ipv6_address ? join("", google_compute_global_address.default_ipv6[*].address) : var.ipv6_address
  url_map                 = var.create_url_map ? join("", google_compute_url_map.default[*].self_link) : var.url_map
  create_http_forward     = var.http_forward || var.https_redirect
  health_checked_backends = { for backend_index, backend_value in var.backends : backend_index => backend_value if backend_value["health_check"] != null }
  is_internal             = var.load_balancing_scheme == "INTERNAL_SELF_MANAGED"
  internal_network        = local.is_internal ? var.network : null
}