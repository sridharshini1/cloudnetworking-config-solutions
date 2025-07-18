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

"""Class for loading resource schema definitions from JSON files."""

import logging
import pathlib
from typing import Any, Dict, Mapping, Sequence

from . import json_loader_utils


Path = pathlib.Path
logger = logging.getLogger(__name__)


class SchemasLoader:
    """Loads schema definitions for various resource types."""

    def __init__(self, schemas_dir: str):
        """Initializes the SchemasLoader.

        Args:
          schemas_dir: The directory path where schema JSON files are located.
        """
        self._schemas_dir = schemas_dir
        self._loaded_schemas: Dict[str, Any] = {}

    def load_schemas_for_resource_types(
        self,
        resource_types: Sequence[str],
        supported_resources: Mapping[str, Any],
    ) -> Dict[str, Any]:
        """Loads schema JSON files for resource types listed as supported.

        Args:
          resource_types: A list of strings representing resource types.
          supported_resources: A dictionary defining which resources are supported
            and their details (e.g., loaded from supported_resources.json).

        Returns:
          A dictionary mapping each resource type to its loaded schema
          (another dictionary). If a type is not in supported_resources,
          or if its schema file is missing/corrupt, it maps to an empty dict,
          and a warning is logged.
        """
        logger.info(
            "Loading schemas for resource types: %s from directory: %s",
            resource_types,
            self._schemas_dir,
        )
        for resource_type in resource_types:
            if resource_type not in supported_resources:
                logger.warning(
                    "Resource type '%s' is not in supported_resources. Skipping"
                    " schema load; will use empty schema.",
                    resource_type,
                )
                self._loaded_schemas[resource_type] = {}
                continue

            schema_filename = f"{resource_type}_schema.json"
            file_path = Path(self._schemas_dir) / schema_filename
            schema_data = json_loader_utils.load_json_file(file_path)

            if schema_data:
                self._loaded_schemas[resource_type] = schema_data
            else:
                logger.warning(
                    "Schema not found or failed to load for supported resource type:"
                    " '%s' at %s. Using empty schema.",
                    resource_type,
                    file_path,
                )
                self._loaded_schemas[resource_type] = {}
        return self._loaded_schemas

    def get_schemas(self) -> Dict[str, Any]:
        """Returns the currently loaded schema definitions.

        Returns:
          A dictionary of loaded schemas.
        """
        return self._loaded_schemas
