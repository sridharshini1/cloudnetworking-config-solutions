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

"""Executable entry point for the JSON generator."""

import logging
import sys

from absl import app
from absl import flags

from . import json_generator

FLAGS = flags.FLAGS

_BASIC_CONFIG_PATH = flags.DEFINE_string(
    "basic_config_path",
    None,
    "Path to the basic JSON configuration file.",
    required=True,
)
_OUTPUT_DIR = flags.DEFINE_string(
    "output_dir",
    None,
    "Directory to save the generated config file.",
    required=True,
)


def main(argv):
    """Parses flags and calls the core generator function."""
    del argv  # Unused

    logging.info("--- Starting Configuration Generation ---")
    output_path = json_generator.generate_config_from_path(
        _BASIC_CONFIG_PATH.value, _OUTPUT_DIR.value
    )

    if output_path:
        logging.info("--- Configuration Generation Finished Successfully ---")
        logging.info("Output file created at: %s", output_path)
    else:
        logging.error("--- Configuration Generation Failed ---")
        sys.exit(1)


if __name__ == "__main__":
    app.run(main)
