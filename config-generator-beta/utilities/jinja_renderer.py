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

"""A shared utility for rendering Jinja2 templates.

This module provides a single, reusable function to handle the logic of
rendering a Jinja2 template with a given context, ensuring that templating
logic is centralized and consistent.
"""

import logging
from typing import Any, Mapping
import re

import jinja2

logger = logging.getLogger(__name__)


# 1. Define the custom filter function.
def regex_replace_filter(s: str, find: str, replace: str) -> str:
    """A Jinja2 filter to perform a regex search and replace."""
    if not isinstance(s, str):
        return s
    return re.sub(find, replace, s)


def render_template(
    template_dir: str, template_name: str, context: Mapping[str, Any]
) -> str | None:
    """Renders a Jinja2 template with the given context.

    Args:
      template_dir: The absolute path to the directory containing the templates.
      template_name: The filename of the template to render.
      context: A dictionary of variables to pass to the template.

    Returns:
      A string containing the rendered template content, or None if an error
      occurred.
    """
    logger.debug(
        "Rendering template '%s' from directory '%s'", template_name, template_dir
    )
    try:
        env = jinja2.Environment(
            loader=jinja2.FileSystemLoader(template_dir),
            autoescape=jinja2.select_autoescape(),
            trim_blocks=True,
            lstrip_blocks=True,
        )

        # 2. Register the custom filter with the environment.
        env.filters["regex_replace"] = regex_replace_filter

        template = env.get_template(template_name)
        return template.render(context)
    except (
        jinja2.TemplateError,
        Exception,
    ) as e:  # pylint: disable=broad-exception-caught
        logger.exception("Failed to render template '%s': %s", template_name, e)
        return None
