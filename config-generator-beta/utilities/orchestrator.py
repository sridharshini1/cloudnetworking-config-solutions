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

"""The core orchestration engine for generating a complete JSON configuration."""

import copy
import logging
import re
from typing import Any, List, Mapping, MutableMapping, Tuple

from . import defaults_json_loader
from . import json_loader_utils
from . import schemas_loader

SchemasLoader = schemas_loader.SchemasLoader
DefaultsJsonLoader = defaults_json_loader.DefaultsJsonLoader
logger = logging.getLogger(__name__)

_MAX_GCP_NAME_LENGTH = 63

_SANITIZE_PATTERN = re.compile(r"[^a-z0-9-]")
_DASH_PATTERN = re.compile(r"-+")


def _deep_update(
    source: MutableMapping[str, Any], overrides: Mapping[str, Any]
) -> MutableMapping[str, Any]:
    """Recursively updates a dictionary with values from another.

    Args:
      source: The dictionary to be updated.
      overrides: The dictionary providing the new values.

    Returns:
      The updated source dictionary.
    """
    for key, value in overrides.items():
        if isinstance(value, dict) and key in source and isinstance(source[key], dict):
            _deep_update(source[key], value)
        else:
            source[key] = value
    return source


def _sanitize_resource_name(name_idea: Any, length: int = _MAX_GCP_NAME_LENGTH) -> str:
    """Sanitizes a string to be a valid GCP resource name.

    Args:
      name_idea: The input string to be sanitized.
      length: The maximum allowed length for the final resource name.

    Returns:
      A sanitized string that conforms to GCP resource naming conventions.
    """
    name = str(name_idea).lower()
    name = _SANITIZE_PATTERN.sub("-", name)
    name = _DASH_PATTERN.sub("-", name)
    name = name.strip("-")
    return name[:length]


class JsonGenerator:
    """Orchestrates the generation of a complete configuration."""

    def __init__(
        self, supported_resources_path: str, schemas_dir: str, defaults_dir: str
    ):
        """Initializes the JsonGenerator engine.

        Args:
            supported_resources_path: Path to the supported_resources.json file.
            schemas_dir: Directory containing resource schema files.
            defaults_dir: Directory containing resource default value files.
        Raises:
            ValueError: If the critical supported_resources.json file fails to load.
        """
        self._supported_resources = json_loader_utils.load_json_file(
            supported_resources_path
        )
        if not self._supported_resources:
            raise ValueError("Failed to load supported_resources.json")

        self._defaults_loader = defaults_json_loader.DefaultsJsonLoader(defaults_dir)
        self._schemas_loader = schemas_loader.SchemasLoader(schemas_dir)
        self._schemas: MutableMapping[str, Any] = {}
        self._defaults: MutableMapping[str, Any] = {}

    def _merge_duplicate_projects(
        self, projects: List[Mapping[str, Any]]
    ) -> List[Mapping[str, Any]]:
        """
        Identifies projects with the same projectId and merges them into one.

        This prevents duplication errors if the input config defines the same
        project multiple times (e.g., once for VPC and once for producers).

        Args:
            projects: The original list of project dictionaries from the config.

        Returns:
            A new list of project dictionaries with duplicates merged.
        """
        merged_projects: dict[str, Any] = {}
        for project in projects:
            project_id = project.get("projectId")
            if not project_id:
                continue

            if project_id not in merged_projects:
                merged_projects[project_id] = copy.deepcopy(project)
            else:
                # If project ID already exists, recursively update it with the new data.
                # This combines keys from both duplicate entries.
                _deep_update(merged_projects[project_id], project)

        return list(merged_projects.values())

    def generate(self, basic_config: Mapping[str, Any]) -> dict[str, Any]:
        """Runs the full generation process.

        Args:
            basic_config: The user-provided basic configuration dictionary.

        Returns:
            A dictionary representing the complete and processed configuration.
        """
        logger.info("Starting JSON generation process.")

        # Create a mutable copy to work with
        working_config = copy.deepcopy(dict(basic_config))
        # De-duplicate and merge projects before any other processing
        if "projects" in working_config:
            logger.info("Checking for and merging duplicate project definitions.")
            merged_projects = self._merge_duplicate_projects(working_config["projects"])
            working_config["projects"] = merged_projects

        all_resource_types = self._get_all_resource_types(working_config)
        self._schemas = self._schemas_loader.load_schemas_for_resource_types(
            all_resource_types, self._supported_resources
        )
        self._defaults = self._defaults_loader.load_defaults_for_resource_types(
            all_resource_types
        )

        # Phase 1: Process explicit resources
        processed_config = self._process_explicit_resources(working_config)

        # Intermediate Phase: Extract nested resources FIRST
        extracted_config = self._extract_nested_resources(processed_config)

        # Phase 2a: Populate PSC allow-lists on the flattened structure
        self._populate_psc_allowed_consumer_projects(extracted_config)

        # Phase 2b: Derive implicit resources from the flattened structure
        derived_config = self._derive_all_implicit_resources(extracted_config)

        # Phase 3: Resolve all resource references
        resolved_config = self._resolve_all_resource_references(derived_config)

        logger.info("JSON generation process complete.")
        return resolved_config

    def _get_all_resource_types(self, config: Mapping[str, Any]) -> List[str]:
        """Scans the basic config to find all unique resource types mentioned.

        Args:
          config: The basic configuration dictionary to scan.

        Returns:
          A list of unique resource type strings found in the configuration.
        """
        resource_types = set()
        for project_data in config.get("projects", []):
            for _, resource_list in project_data.items():
                if not isinstance(resource_list, list):
                    continue
                for item in resource_list:
                    if isinstance(item, dict) and "type" in item:
                        resource_types.add(item["type"])
        return list(resource_types)

    def _get_instance_name(
        self, prefix: str, base_name: str, suffix: str, i: int, num_instances: int
    ) -> str:
        """Creates a sanitized resource name from its component parts.

        Args:
          prefix: The global name prefix.
          base_name: The base name of the resource from the config.
          suffix: The global name suffix.
          i: The current instance number in a count loop.
          num_instances: The total number of instances being created.

        Returns:
          The fully formed and sanitized resource name.
        """
        name_parts = [
            p
            for p in [
                prefix,
                base_name,
                str(i) if num_instances > 1 else "",
                suffix,
            ]
            if p
        ]
        return _sanitize_resource_name("-".join(name_parts))

    def _process_explicit_resources(
        self, basic_config: Mapping[str, Any]
    ) -> dict[str, Any]:
        """Processes defined resources, handling 'count' expansion and naming.

        Args:
          basic_config: The user-provided basic configuration.

        Returns:
          A new configuration dictionary with all explicit resources fully
          processed and expanded.
        """
        logger.info("Phase 1: Processing explicit resources.")
        processed_config = copy.deepcopy(dict(basic_config))
        prefix = processed_config.get("namePrefix", "")
        suffix = processed_config.get("nameSuffix", "")

        metadata_keys = {
            "projectId",
            "pscSettings",
            "description",
            "name",
            "namePrefix",
            "nameSuffix",
            "defaultRegion",
        }

        for project_data in processed_config.get("projects", []):
            for category, resource_list in project_data.items():
                if category in metadata_keys or not isinstance(resource_list, list):
                    continue
                processed_resources = []
                for item_data in resource_list:
                    if not isinstance(item_data, dict):
                        continue
                    base_name = item_data.get("name")
                    if not base_name:
                        continue
                    count = item_data.get("count", 1)
                    num_instances = count if isinstance(count, int) and count > 0 else 1
                    for i in range(1, num_instances + 1):
                        instance_data = copy.deepcopy(item_data)
                        if "count" in instance_data:
                            del instance_data["count"]
                        instance_data["name"] = self._get_instance_name(
                            prefix, base_name, suffix, i, num_instances
                        )
                        processed_item = self._process_resource_instance(instance_data)
                        if processed_item:
                            processed_resources.append(processed_item)
                project_data[category] = processed_resources
        return processed_config

    def _process_resource_instance(
        self, resource_data: Mapping[str, Any]
    ) -> Mapping[str, Any] | None:
        """Builds a complete resource instance by layering schema, defaults, and user data.

        Args:
          resource_data: The dictionary for a single resource instance.

        Returns:
          A fully processed dictionary for the resource, or None.
        """
        resource_type = resource_data.get("type")
        if not resource_type:
            logger.warning(
                "Resource data is missing 'type': %s.",
                resource_data.get("name", "Unnamed"),
            )
            return copy.deepcopy(dict(resource_data))

        schema = self._schemas.get(resource_type)
        defaults_for_type = (
            self._defaults.get(resource_type, {}) if self._defaults else {}
        )

        complete_config = self._create_instance_from_schema(schema or {})
        if defaults_for_type:
            _deep_update(complete_config, defaults_for_type)
        _deep_update(complete_config, resource_data)
        return complete_config

    def _create_instance_from_schema(
        self, schema_definition: Mapping[str, Any]
    ) -> MutableMapping[str, Any]:
        """Creates a dict from a JSON schema, initializing properties to None.

        Args:
          schema_definition: The JSON schema for a specific resource type.

        Returns:
          A dictionary with keys from the schema properties, initialized to None.
        """
        if (
            not isinstance(schema_definition, dict)
            or schema_definition.get("type") != "object"
        ):
            return {}
        instance = {}
        for prop_name, prop_schema in schema_definition.get("properties", {}).items():
            if isinstance(prop_schema, dict) and prop_schema.get("readOnly"):
                continue
            instance[prop_name] = None
        return instance

    def _get_resolved_connectivity_mode(
        self, producer_info: Mapping[str, Any]
    ) -> str | None:
        """Gets explicit connectivityType or determines it via fallback logic.

        Args:
          producer_info: The dictionary for a single producer resource.

        Returns:
          The connectivity mode string (e.g., 'psc') or None.
        """
        if resolved_mode := producer_info.get("connectivityType"):
            return resolved_mode
        producer_type = producer_info.get("type")
        if not producer_type:
            return None
        resource_details = (self._supported_resources or {}).get(producer_type, {})

        supported_options = resource_details.get("connectivityOptions", [])
        if "psc" in supported_options:
            return "psc"
        if "psa" in supported_options:
            return "psa"
        return None

    def _find_network_details(
        self, network_name: str, all_projects: List[Mapping[str, Any]]
    ) -> Tuple[str | None, str | None]:
        """Finds a network's owner project ID and full path.

        Args:
          network_name: The short name of the network to find.
          all_projects: A list of all project configurations.

        Returns:
          A tuple containing the project ID and the full network URI, or (None,
          None).
        """
        for proj in all_projects:
            for vpc in proj.get("vpc", []):
                if vpc.get("name") == network_name:
                    project_id = proj.get("projectId")
                    return (
                        project_id,
                        vpc.get("selfLink")
                        or f"projects/{project_id}/global/networks/{network_name}",
                    )
        return None, None

    def _find_subnet_details(
        self,
        subnet_name: str,
        network_name: str,
        all_projects: List[Mapping[str, Any]],
    ) -> Tuple[str | None, str | None, str | None]:
        """Finds a subnet's owner project ID, region, and full path.

        Args:
          subnet_name: The short name of the subnetwork to find.
          network_name: The name of the VPC that the subnet belongs to.
          all_projects: A list of all project configurations.

        Returns:
          A tuple containing project ID, region, and full subnet URI, or (None,
          None, None).
        """
        for proj in all_projects:
            project_id = proj.get("projectId")

            # 1. Check inside the VPC object
            for vpc in proj.get("vpc", []):
                if vpc and vpc.get("name") == network_name:
                    for subnet_list_key in ["subnets", "subnetworks"]:
                        subnet_list = vpc.get(subnet_list_key)
                        if subnet_list:
                            for subnet in subnet_list:
                                if subnet.get("name") == subnet_name:
                                    region = subnet.get("region")
                                    path = f"projects/{project_id}/regions/{region}/subnetworks/{subnet_name}"
                                    return project_id, region, path

            # 2. If not found, check the top-level 'subnets' list
            for subnet in proj.get("subnets", []):
                if subnet:
                    subnet_network_uri = subnet.get("network", "")
                    if (
                        subnet.get("name") == subnet_name
                        and network_name in subnet_network_uri
                    ):
                        region = subnet.get("region")
                        path = f"projects/{project_id}/regions/{region}/subnetworks/{subnet_name}"
                        return project_id, region, path

        return None, None, None

    def _derive_all_implicit_resources(self, config: dict[str, Any]) -> dict[str, Any]:
        """Scans the config and derives all implicit resources like NAT and PSC.

        Args:
          config: The configuration dictionary after explicit resources are
            processed.

        Returns:
          A new configuration dictionary with the derived resources merged in.
        """
        logger.info("Phase 2: Deriving Implicit Resources.")
        # This is a mutable copy, so we can modify it directly.
        config_copy = copy.deepcopy(config)
        derived_map: dict[str, dict[str, List[dict[str, Any]]]] = {}
        all_projects = config_copy.get("projects", [])
        if not all_projects:
            return config_copy

        # --- Main loop to derive NAT, Firewall, and PSC resources ---
        for project_data in all_projects:
            project_id = project_data.get("projectId")
            if not project_id:
                continue
            for vpc in project_data.get("vpc", []):
                if vpc.get("createNat"):
                    router_def = {
                        "type": "router",
                        "name": _sanitize_resource_name(f"router-{vpc['name']}-nat"),
                        "network": f"projects/{project_id}/global/networks/{vpc['name']}",
                        "region": config.get("defaultRegion"),
                    }
                    derived_map.setdefault(project_id, {}).setdefault(
                        "routers", []
                    ).append(router_def)

            for producer in project_data.get("producers", []):
                producer_name = producer.get("name", "")

                if producer.get("createRequiredFwRules"):
                    # First, try to get the network explicitly from the producer for overrides.
                    network_for_fw = producer.get("networkForFirewall") or producer.get(
                        "network"
                    )

                    # If not found, and it's a PSC producer, infer it from pscSettings.
                    if (
                        not network_for_fw
                        and self._get_resolved_connectivity_mode(producer) == "psc"
                    ):
                        # Find the project that contains the global PSC settings.
                        pscSettings_project = next(
                            (p for p in all_projects if "pscSettings" in p), None
                        )
                        if pscSettings_project:
                            network_for_fw = pscSettings_project.get(
                                "pscSettings", {}
                            ).get("networkForPsc")

                    # Now, if we successfully found a network name, create the firewall.
                    if network_for_fw:
                        # Find the project that owns the network where the firewall needs to be created.
                        fw_project_id, network_path = self._find_network_details(
                            network_for_fw, all_projects
                        )
                        if fw_project_id and network_path:
                            logger.info(
                                "Deriving Firewall Rule for producer '%s' in network '%s'",
                                producer_name,
                                network_path,
                            )
                            fw_def = {
                                "type": "firewall_rule",
                                "name": _sanitize_resource_name(
                                    f"fw-allow-{producer_name}"
                                ),
                                "network": network_path,
                                "targetTags": producer.get("allowedConsumersTags", []),
                            }
                            # Add the rule to be created in the project that owns the network.
                            derived_map.setdefault(fw_project_id, {}).setdefault(
                                "firewalls", []
                            ).append(fw_def)

                # --- PSC DERIVATION ---
                resolved_mode = self._get_resolved_connectivity_mode(producer)
                if resolved_mode == "psc":
                    # 1. Prioritize getting network/subnet from the producer itself.
                    net_name = producer.get("network")
                    sub_name = producer.get("subnet")
                    # 2. If not found on the producer, fall back to the global pscSettings.
                    if not net_name or not sub_name:
                        pscSettings_project = next(
                            (p for p in all_projects if "pscSettings" in p), None
                        )
                        if pscSettings_project:
                            pscSettings = pscSettings_project.get("pscSettings", {})
                            # Only use the global setting if the producer didn't have its own.
                            net_name = net_name or pscSettings.get("networkForPsc")
                            sub_name = sub_name or pscSettings.get("subnetForPsc")

                    # 3. If a network and subnet were found (either way), proceed.
                    if net_name and sub_name:
                        ep_proj, ep_region, sub_path = self._find_subnet_details(
                            sub_name, net_name, all_projects
                        )
                        _, net_path = self._find_network_details(net_name, all_projects)
                        if ep_proj and ep_region and sub_path and net_path:
                            addr_name = _sanitize_resource_name(
                                f"addr-{producer_name}-psc"
                            )
                            addr_def = {
                                "type": "address",
                                "name": addr_name,
                                "region": ep_region,
                                "subnetwork": sub_path,
                            }
                            fr_def = {
                                "type": "forwardingrule",
                                "name": _sanitize_resource_name(
                                    f"fwrule-{producer_name}-psc"
                                ),
                                "region": ep_region,
                                "network": net_path,
                                "subnetwork": sub_path,
                                "IPAddress": addr_name,
                                "targetProducerName": producer_name,
                            }
                            derived_map.setdefault(ep_proj, {}).setdefault(
                                "addresses", []
                            ).append(addr_def)
                            derived_map.setdefault(ep_proj, {}).setdefault(
                                "forwardingRules", []
                            ).append(fr_def)

        # --- Service Connection Policy (SCP) derivation logic ---
        logger.info("Checking for SCP requirements.")
        producers_requiring_scp = []
        for project_data in all_projects:
            for producer in project_data.get("producers", []):
                if self._get_resolved_connectivity_mode(producer) == "scp":
                    producers_requiring_scp.append(producer)

        if producers_requiring_scp:
            logger.info(
                "Found producers requiring SCP. Configuring Service Connection Policy."
            )
            for producer in producers_requiring_scp:
                subnet_name = producer.get("subnet")
                if not subnet_name:
                    continue

                subnet_object = None
                owner_project = None
                for p in all_projects:
                    found_subnet = next(
                        (
                            s
                            for s in p.get("subnets", [])
                            if s.get("name") == subnet_name
                        ),
                        None,
                    )
                    if found_subnet:
                        subnet_object = found_subnet
                        owner_project = p
                        break

                if subnet_object and owner_project:
                    network_name = subnet_object.get("network")
                    if network_name:
                        target_vpc = next(
                            (
                                v
                                for v in owner_project.get("vpc", [])
                                if v.get("name") == network_name
                            ),
                            None,
                        )
                        if target_vpc:
                            logger.info(
                                "Enabling SCP on VPC '%s' for subnet '%s'",
                                target_vpc.get("name"),
                                subnet_name,
                            )
                            target_vpc["createScpPolicy"] = True
                            subnets_for_policy = target_vpc.setdefault(
                                "subnetsForScpPolicy", []
                            )
                            if subnet_name not in subnets_for_policy:
                                subnets_for_policy.append(subnet_name)

        # --- Merge the derived resources (NAT, PSC, Firewall) back into the config ---
        if derived_map:
            logger.info("Merging derived resources into configuration...")
            for target_project_id, categories in derived_map.items():
                for proj_cfg in all_projects:
                    if proj_cfg.get("projectId") == target_project_id:
                        for category, resources in categories.items():
                            processed_resources = [
                                self._process_resource_instance(r) for r in resources
                            ]
                            proj_cfg.setdefault(category, []).extend(
                                filter(None, processed_resources)
                            )
                            logger.info(
                                "Added %d '%s' resource(s) to project '%s'",
                                len(processed_resources),
                                category,
                                target_project_id,
                            )

        return config_copy

    def _extract_nested_resources(self, config: dict[str, Any]) -> dict[str, Any]:
        """Finds and moves nested resources from their parents to a top-level list.

        This is a critical step that "flattens" the configuration, making it
        easier for subsequent processing phases to find and operate on all
        resources of a given type.

        Args:
          config: The configuration dictionary to process.

        Returns:
          A new configuration dictionary with a flattened resource structure.
        """
        logger.info("Intermediate Phase: Processing Nested Resources.")
        config_copy = copy.deepcopy(config)

        for project in config_copy.get("projects", []):
            newly_extracted: dict[str, List[dict[str, Any]]] = {}
            for resource_list in project.values():
                if not isinstance(resource_list, list):
                    continue
                for resource in resource_list:
                    resource_type = resource.get("type")
                    resource_info = (self._supported_resources or {}).get(
                        resource_type, {}
                    )

                    for list_key, nested_type in resource_info.get(
                        "nestedResources", {}
                    ).items():
                        if list_key in resource:
                            logger.info(
                                "Extracting %s from %s", list_key, resource.get("name")
                            )

                            # Get the parent resource's name to create the link.
                            parent_resource_name = resource.get("name")

                            # Special case for VPCs to disable auto-creation of subnets
                            if resource_type == "vpc":
                                resource["autoCreateSubnetworks"] = False

                            for nested_item_data in resource.pop(list_key, []):
                                nested_item_data["type"] = nested_type

                                if nested_type == "subnetwork":
                                    nested_item_data["network"] = parent_resource_name

                                processed_item = self._process_resource_instance(
                                    nested_item_data
                                )
                                if processed_item:
                                    newly_extracted.setdefault(list_key, []).append(
                                        processed_item
                                    )

            # Merge the extracted resources back into the project
            for category, new_resources in newly_extracted.items():
                project.setdefault(category, []).extend(new_resources)

        return config_copy

    def _resolve_all_resource_references(
        self, config: dict[str, Any]
    ) -> dict[str, Any]:
        """Final pass to replace all short-name resource references with their full URIs.

        Args:
          config: The configuration dictionary to process.

        Returns:
          The final configuration dictionary with all resource references resolved.
        """
        logger.info("Phase 3: Resolving Resource References.")
        uri_map: dict[Tuple[str, str], str] = {}
        config_copy = copy.deepcopy(config)

        # First pass: build a map of all resource URIs. This part is correct.
        for project in config_copy.get("projects", []):
            project_id = project.get("projectId")
            for resource_list in project.values():
                if not isinstance(resource_list, list):
                    continue
                for resource in resource_list:
                    res_type = resource.get("type")
                    res_name = resource.get("name")
                    res_info = (self._supported_resources or {}).get(res_type, {})
                    if not all(
                        [project_id, res_type, res_name, res_info.get("uriTemplate")]
                    ):
                        continue
                    try:
                        uri_params = resource.copy()
                        uri_params["projectId"] = project_id
                        uri = res_info["uriTemplate"].format(**uri_params)
                        uri_map[(res_type, res_name)] = uri
                        resource["selfLink"] = uri
                    except KeyError as e:
                        logger.warning(
                            "Could not format URI for %s: missing key %s", res_name, e
                        )

        # Second pass: resolve references using a smarter recursive resolver.
        def _recursive_resolver(current_obj: Any, parent_type: str | None = None):
            """Recursively traverses, passing parent context to resolve URIs."""
            if isinstance(current_obj, list):
                for item in current_obj:
                    _recursive_resolver(item, parent_type)
                return

            if not isinstance(current_obj, dict):
                return

            current_type = current_obj.get("type") or parent_type
            resource_info = (self._supported_resources or {}).get(current_type, {})

            ref_fields = resource_info.get("referenceFields", [])

            for key, value in current_obj.items():
                if (
                    key in ref_fields
                    and isinstance(value, str)
                    and value
                    and "/" not in value
                ):
                    was_resolved = False  # Use a simple flag
                    for (_, match_name), uri in uri_map.items():
                        if match_name == value:
                            current_obj[key] = uri
                            was_resolved = True  # Set the flag on success
                            break

                    # Only log a warning if the flag was never set
                    if not was_resolved:
                        logger.warning(
                            "Could not resolve reference for field '%s' with value '%s'. "
                            "Check for typos or ensure the target resource is defined.",
                            key,
                            value,
                        )
                else:
                    _recursive_resolver(value, current_type)

        _recursive_resolver(config_copy)
        return config_copy

    def _populate_psc_allowed_consumer_projects(self, config: MutableMapping[str, Any]):
        """Populates 'allowedConsumerProjects' list based on matching tags."""
        logger.info("Populating PSC allow-lists based on tags.")
        all_projects = config.get("projects", [])
        all_tagged_consumers = []
        for proj_cfg in all_projects:
            project_id = proj_cfg.get("projectId")
            if not project_id:
                continue
            for consumer in proj_cfg.get("consumers", []):
                tags_obj = consumer.get("tags")
                tag_list = []
                if isinstance(tags_obj, dict):
                    tag_list = tags_obj.get("items", [])
                elif isinstance(tags_obj, list):
                    tag_list = tags_obj
                if tag_list:
                    all_tagged_consumers.append(
                        {"projectId": project_id, "tags": set(tag_list)}
                    )

        if not all_tagged_consumers:
            return

        for proj_cfg in all_projects:
            for producer in proj_cfg.get("producers", []):
                if self._get_resolved_connectivity_mode(producer) != "psc":
                    continue
                if not (allowed_tags := set(producer.get("allowedConsumersTags", []))):
                    continue

                settings = producer.setdefault("settings", {})
                ip_config = settings.setdefault("ipConfiguration", {})
                psc_config = ip_config.setdefault("pscConfig", {})
                current_allowed = set(psc_config.get("allowedConsumerProjects", []))
                for tagged_consumer in all_tagged_consumers:
                    if not allowed_tags.isdisjoint(tagged_consumer["tags"]):
                        current_allowed.add(tagged_consumer["projectId"])
                psc_config["allowedConsumerProjects"] = sorted(list(current_allowed))
