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

"""Core library for generating a complete JSON configuration.

This module contains the primary function to orchestrate the generation process,
making the logic importable and reusable by other tools and tests.
"""

import json
import logging
import os

from . import json_loader_utils
from . import orchestrator

_JsonGenerator = orchestrator.JsonGenerator


def generate_config_from_path(
    basic_config_path: str,
    config_name: str,
    output_dir: str,
    supported_resources_path: str,
    schemas_dir: str,
    defaults_dir: str,
) -> str | None:
    """Generates a complete config from a path and saves it to the output dir.

    Args:
        basic_config_path: The full path to the input basic.json file.
        config_name: The base name for the output file (e.g., 'cloudsql_psc').
        output_dir: The directory where the complete.json will be saved.

    Returns:
        The full path to the generated file on success, otherwise None.
    """
    try:
        generator = _JsonGenerator(
            supported_resources_path=supported_resources_path,
            schemas_dir=schemas_dir,
            defaults_dir=defaults_dir,
        )
        logging.info("Loading basic configuration from: %s", basic_config_path)
        basic_config = json_loader_utils.load_json_file(basic_config_path)
        if not basic_config:
            raise ValueError("Basic configuration file could not be loaded.")

        complete_config = generator.generate(basic_config)

        # Use the passed-in config_name for the output file
        output_filename = f"{config_name}-complete.json"
        output_path = os.path.join(output_dir, output_filename)

        # The directory is already created by the calling script, but this is safe
        logging.info("Saving complete configuration to: %s", output_path)
        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(complete_config, f, indent=2, ensure_ascii=False)

        return output_path
    except (ValueError, FileNotFoundError, json.JSONDecodeError) as e:
        logging.error("An error occurred during JSON generation: %s", e, exc_info=True)
        return None
