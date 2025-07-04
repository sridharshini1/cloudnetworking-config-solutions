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

resource "google_compute_managed_ssl_certificate" "ssl_cert" {
  description = var.description # description - (optional) is a type of string
  name        = var.name        # name - (optional) is a type of string
  project     = var.project     # project - (optional) is a type of string
  type        = var.type        # type - (optional) is a type of string

  dynamic "managed" {
    for_each = var.managed
    content {
      domains = managed.value["domains"] # domains - (required) is a type of list of string
    }
  }

  dynamic "timeouts" {
    for_each = var.timeouts
    content {
      create = timeouts.value["create"] # create - (optional) is a type of string
      delete = timeouts.value["delete"] # delete - (optional) is a type of string
    }
  }

}