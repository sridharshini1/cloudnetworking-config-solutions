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

"""Class for loading default resource configurations from JSON files."""

import logging
import pathlib
from typing import Any, Dict, Sequence

from . import json_loader_utils

Path = pathlib.Path
logger = logging.getLogger(__name__)


class DefaultsJsonLoader:
    """Loads default configurations for various resource types."""

    def __init__(self, defaults_dir: str):
        """Initializes the DefaultsJsonLoader.

        Args:
          defaults_dir: The directory path where default JSON files are located.
        """
        self._defaults_dir = defaults_dir
        self._loaded_defaults: Dict[str, Any] = {}

    def load_defaults_for_resource_types(
        self, resource_types: Sequence[str]
    ) -> Dict[str, Any]:
        """Loads all default JSON files for the given list of resource types.

        Args:
          resource_types: A list of strings representing the resource types for
            which to load defaults.

        Returns:
          A dictionary mapping each resource type to its loaded default
          configuration (another dictionary). If a default file for a type
          is not found or fails to load, that type will map to an empty dict,
          and a warning will be logged.
        """
        logger.info(
            "Loading defaults for resource types: %s from directory: %s",
            resource_types,
            self._defaults_dir,
        )
        for resource_type in resource_types:
            defaults_filename = f"{resource_type}_defaults.json"
            file_path = Path(self._defaults_dir) / defaults_filename
            defaults_data = json_loader_utils.load_json_file(file_path)
            if defaults_data:
                self._loaded_defaults[resource_type] = defaults_data
            else:
                logger.warning(
                    "Defaults not found or failed to load for resource type: '%s' at"
                    " %s. Using empty defaults.",
                    resource_type,
                    file_path,
                )
                self._loaded_defaults[resource_type] = {}
        return self._loaded_defaults

    def get_defaults(self) -> Dict[str, Any]:
        """Returns the currently loaded default configurations.

        Returns:
          A dictionary of loaded defaults.
        """
        return self._loaded_defaults
