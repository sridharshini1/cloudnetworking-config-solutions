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

"""Utility functions for loading JSON files."""

import json
import logging
import pathlib
from typing import Any


Path = pathlib.Path
logger = logging.getLogger(__name__)


def load_json_file(file_path: str | Path) -> dict[str, Any] | None:
    """Loads a JSON file with error handling and returns its content.

    Args:
      file_path: The path to the JSON file. Can be a string or a Path object.

    Returns:
      A dictionary representing the JSON content, or None if an error
      occurred during loading or parsing.
    """
    logger.debug("Attempting to load JSON file: %s", file_path)
    try:
        with open(Path(file_path), "r", encoding="utf-8") as f:
            return json.load(f)
    except FileNotFoundError:
        logger.error("File not found: %s", file_path)
        return None
    except json.JSONDecodeError as e:
        logger.exception("Error decoding JSON from %s: %s", file_path, e)
        return None
    except OSError as e:
        logger.exception("An unexpected error occurred loading %s: %s", file_path, e)
        return None
