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

"""This script orchestrates the generation of Terraform .tfvars and YAML files from a complete JSON configuration.

It acts as the primary orchestrator for the Terraform Config Generator,
supporting multiple modes of operation via command-line flags.
"""

import argparse
import json
import os
import platform
import shutil
import subprocess
import sys
from typing import Any, Dict, List, Tuple

from utilities import json_generator
from utilities import config_generator_engine

generate_config_from_path = json_generator.generate_config_from_path


# --- Configuration Constants ---
# SCRIPT_DIR is the directory where the main script lives
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# PROJECT_ROOT_DIR is one level up from the script's directory. This is the main project folder.
PROJECT_ROOT_DIR = os.path.dirname(SCRIPT_DIR)

# The main output directory for all generated files.
MAIN_OUTPUT_DIR = os.path.join(PROJECT_ROOT_DIR, "configuration")

# The directory containing the input JSON files.
ARCHITECTURE_SPEC_DIR = os.path.join(
    PROJECT_ROOT_DIR, "config-generator-beta/architecture-spec"
)

# Paths to the data files, which live at the project root.
DEFAULTS_DIR = os.path.join(PROJECT_ROOT_DIR, "config-generator-beta/.defaults")
SCHEMA_DIR = os.path.join(PROJECT_ROOT_DIR, "config-generator-beta/.schema")
SUPPORTED_RESOURCES_PATH = os.path.join(SCHEMA_DIR, "supported_resources.json")

# The path to the Jinja2 templates (the source templates).
TEMPLATES_BASE_DIR = os.path.join(SCRIPT_DIR, "configuration")

# Constants for generated filenames
ORGANIZATION_TFVARS_FILENAME = "organization.tfvars"
NETWORKING_TFVARS_FILENAME = "networking.tfvars"
PRODUCER_CONNECTIVITY_TFVARS_FILENAME = "producer-connectivity.tfvars"

_GENERATED_CONFIG_FILES: List[str] = []
STATIC_DIRECTORIES_TO_COPY = [
    "security/Firewall",
    "security/Certificates",
    "security/SecurityProfile",
    "consumer-load-balancing",
    "networking/FirewallEndpoint",
    "networking/ncc",
]

# --- Utility and Workflow Functions ---


def print_red(text: str):
    """Prints text to console in red color."""
    print(f"\033[91m{text}\033[0m")


def print_yellow(text: str):
    """Prints text to console in a lighter yellow color."""
    print(f"\033[38;5;226m{text}\033[0m")


def load_json_file(file_path: str, is_critical: bool = True) -> Dict[str, Any] | None:
    """Loads a JSON file with error handling."""
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print_red(f"Error loading JSON file {file_path}: {e}")
        if is_critical:
            raise
        return None


def _write_file(output_path: str, content: str | None) -> bool:
    """Writes content to a file, creating directories if needed."""
    if content is None:
        print_red(f"Error: Received no content to write to {output_path}.")
        return False
    if not content.strip():
        print(f"  Skipping empty file: {output_path}")
        return True
    try:
        dir_path = os.path.dirname(output_path)
        if dir_path:
            os.makedirs(dir_path, exist_ok=True)
        with open(output_path, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"  Successfully generated {output_path}")
        _GENERATED_CONFIG_FILES.append(output_path)
        return True
    except IOError as e:
        print_red(f"Error writing to file {output_path}: {e}")
        return False


def _open_file_in_default_app(file_path: str):
    """Opens a file using the default application for the current OS."""
    if not os.path.exists(file_path):
        print(f"Warning: Cannot open '{file_path}'. File not found.", file=sys.stderr)
        return
    system_name = platform.system()
    try:
        if system_name == "Windows":
            os.startfile(file_path)  # pytype: disable=module-attr
        elif system_name == "Darwin":
            subprocess.Popen(["open", file_path])
        else:
            subprocess.Popen(["xdg-open", file_path])
        print(f"Opened '{file_path}' in default application.")
    except (subprocess.CalledProcessError, OSError) as e:
        print(f"Error opening '{file_path}': {e}", file=sys.stderr)


def select_basic_json() -> Tuple[str, str]:
    """Prompts user to select an architecture from the 'architecture-spec' dir."""
    print("\n--- Step 1: Select an Architecture ---")
    if not os.path.isdir(ARCHITECTURE_SPEC_DIR):
        print(
            f"Error: Input directory '{ARCHITECTURE_SPEC_DIR}' not found.",
            file=sys.stderr,
        )
        sys.exit(1)

    json_files = []
    for dirpath, _, filenames in os.walk(ARCHITECTURE_SPEC_DIR):
        for filename in filenames:
            if filename.endswith(".json"):
                full_path = os.path.join(dirpath, filename)
                relative_path = os.path.relpath(full_path, ARCHITECTURE_SPEC_DIR)
                json_files.append(relative_path)

    if not json_files:
        print(
            f"Error: No architecture files found in '{ARCHITECTURE_SPEC_DIR}'."
            " Exiting.",
            file=sys.stderr,
        )
        sys.exit(1)

    json_files.sort()
    print("Available architectures ready for deployment:")
    for i, filename in enumerate(json_files):
        display_name = os.path.splitext(filename)[0]
        print(f"  {i+1}. {display_name}")

    while True:
        try:
            choice_str = input(f"Enter file number (1-{len(json_files)}): ")
            choice = int(choice_str)
            if 1 <= choice <= len(json_files):
                selected_relative_path = json_files[choice - 1]
                path = os.path.join(ARCHITECTURE_SPEC_DIR, selected_relative_path)
                name = os.path.splitext(os.path.basename(selected_relative_path))[0]
                print(f"Selected: {name}")
                return path, name
            else:
                print_red(
                    f"Invalid number. Please enter a number between 1 and {len(json_files)}."
                )
        except (ValueError, IndexError):
            print_red("Invalid input. Please enter a number.")


def prompt_for_json_review(file_path: str, architecture_spec_name: str) -> bool:
    """Opens the complete.json file and asks for user confirmation."""
    print("\n--- Step 3 & 4: Reviewing complete.json ---")
    _open_file_in_default_app(file_path)
    print_yellow(
        f"\nComplete architecture spec for '{architecture_spec_name}' has been"
        " generated."
    )
    prompt_text = (
        "\033[92mConfirm to generate all Terraform configuration files? (yes/y):"
        " \033[0m"
    )
    return input(prompt_text).strip().lower() in ["yes", "y"]


def open_and_advise_terraform_review():
    """Opens all generated Terraform files and advises the user to review them."""
    if not _GENERATED_CONFIG_FILES:
        print(
            "\nWarning: No new TFVARS/YAML files were generated to review.",
            file=sys.stderr,
        )
        return
    print("\n--- Step 6: Review Generated Files ---")
    for f_path in _GENERATED_CONFIG_FILES:
        _open_file_in_default_app(f_path)
    print(
        f"\nAll {len(_GENERATED_CONFIG_FILES)} generated Terraform/YAML files"
        " have been opened."
    )


def prompt_for_final_deployment() -> bool:
    """Shows the final warning and asks for deployment confirmation."""
    print("\n--- Step 7: Final Deployment Confirmation ---")
    print_red("\n" + "=" * 80)
    print_red(
        "WARNING: Deploying this architecture requires OWNER permissions on the"
        " GCP projects."
    )
    print_red("Ensure you have the necessary permissions before proceeding.")
    print(
        "\nAlternatively, use run.sh automation at the folder level based on your"
        " roles/permissions."
    )
    print_red("=" * 80)
    prompt_text = "\033[92mConfirm to proceed with deployment? (yes/y): \033[0m"
    return input(prompt_text).strip().lower() in ["yes", "y"]


def run_terraform_fmt(output_dir: str):
    """Runs 'terraform fmt --recursive' on the specified directory."""
    print("\n--- Step 5.1: Running terraform fmt ---")
    try:
        subprocess.run(
            ["terraform", "fmt", "-recursive"],
            cwd=output_dir,
            check=True,
            capture_output=True,
        )
        print("Formatting complete.")
    except (FileNotFoundError, subprocess.CalledProcessError) as e:
        print_yellow(f"\nWarning: Could not run 'terraform fmt'. Error: {e}")


def run_deployment():
    """Executes the run.sh script from the correct directory."""
    print("\n--- Step 8: Running Deployment ---")
    execution_dir = os.path.join(SCRIPT_DIR, "../execution")
    run_sh_path = os.path.join(execution_dir, "run.sh")
    if not os.path.exists(run_sh_path):
        print_red(f"Error: Deployment script not found at {run_sh_path}.")
        sys.exit(1)
    try:
        command = ["bash", "./run.sh", "-s", "all", "-t", "init-apply-auto-approve"]

        print(
            f"Executing deployment script from within the '{execution_dir}' directory..."
        )
        print(f"Command: {' '.join(command)}")

        subprocess.run(
            command,
            check=True,
            cwd=execution_dir,
        )

        print("\nDeployment completed successfully!")
    except (subprocess.CalledProcessError, OSError) as e:
        print_red(f"Error during deployment: {e}")
        sys.exit(1)


# --- Main Orchestration Function ---


def generate_all_tf_files(complete_config_path: str, output_dir: str) -> bool:
    """Orchestrates the generation of all Terraform .tfvars and YAML files."""
    print("\n--- Step 5: Creating Terraform & YAML Files ---")
    _GENERATED_CONFIG_FILES.clear()

    complete_config = load_json_file(complete_config_path)
    supported_resources = load_json_file(SUPPORTED_RESOURCES_PATH)
    if not complete_config or not supported_resources:
        return False

    # 1. Instantiate the engine, providing the base path for templates
    engine = config_generator_engine.TerraformArtifactGenerator(
        complete_config=complete_config,
        supported_resources=supported_resources,
        templates_base_dir=TEMPLATES_BASE_DIR,
    )

    # 2. Get the content for all the main tfvars files
    org_tfvars = engine.generate_organisation_tfvars()
    net_tfvars = engine.generate_networking_tfvars()
    psc_tfvars = engine.generate_producer_connectivity_tfvars()

    # 3. Get the content for all the individual resource files
    resource_files = engine.generate_all_resource_files()

    if resource_files is None:
        return False

    # 4. Use the engine's own write_file method to save the files
    success = True
    if not engine.write_file(
        os.path.join(output_dir, ORGANIZATION_TFVARS_FILENAME), org_tfvars
    ):
        success = False
    if not engine.write_file(
        os.path.join(output_dir, NETWORKING_TFVARS_FILENAME), net_tfvars
    ):
        success = False
    if not engine.write_file(
        os.path.join(output_dir, PRODUCER_CONNECTIVITY_TFVARS_FILENAME),
        psc_tfvars,
    ):
        success = False

    for file_path, content in resource_files.items():
        if not engine.write_file(os.path.join(output_dir, file_path), content):
            success = False
    _GENERATED_CONFIG_FILES.extend(engine.get_generated_files())
    return success


def find_existing_configuration(require_complete_json: bool) -> Tuple[str, str | None]:
    """Finds the existing configuration directory and validates its contents.

    This function checks for the existence of the main output directory. If
    required, it also verifies that a single '*-complete.json' file is present.
    It will exit the script with a specific error message if validation fails.

    Args:
        require_complete_json: If True, the function will fail if a
          '-complete.json' file is not found inside the directory.

    Returns:
        A tuple containing:
          - The absolute path to the 'configuration' directory.
          - The absolute path to the 'complete.json' file (or None).
    """
    print("\n--- Step 1: Finding Existing Configuration ---")
    if not os.path.isdir(MAIN_OUTPUT_DIR):
        print_red(
            f"Error: The configuration directory does not exist at the expected path."
        )
        print_red(f"Path: {MAIN_OUTPUT_DIR}")
        print_yellow(
            "\nPlease run with '--all' or '--full-spec' to generate a new configuration, or use --help for more info."
        )
        sys.exit(1)

    print(f"Found configuration directory: {MAIN_OUTPUT_DIR}")
    complete_json_path = None

    if require_complete_json:
        try:
            found_files = [
                f for f in os.listdir(MAIN_OUTPUT_DIR) if f.endswith("-complete.json")
            ]
            if not found_files:
                raise FileNotFoundError
            # In a valid state, there should only ever be one complete.json
            complete_json_path = os.path.join(MAIN_OUTPUT_DIR, found_files[0])
            print(f"Found spec file: {complete_json_path}")
        except FileNotFoundError:
            print_red(f"Error: No '*-complete.json' file found in '{MAIN_OUTPUT_DIR}'.")
            print_yellow(
                "This file is required for the '--terraform' and '--terraform-apply' flags."
            )
            sys.exit(1)

    return MAIN_OUTPUT_DIR, complete_json_path


def _copy_static_files(
    complete_config: dict,
    supported_resources: dict,
    templates_dir: str,
    output_dir: str,
) -> bool:
    """
    Copies static .tfvars files and entire static directories from the
    templates directory to the final output directory.
    """
    print("\n--- Step 5.1: Copying static files and directories ---")
    copied_sources = set()  #
    success = True

    # First, copy the entire directories specified in the list
    for dir_name in STATIC_DIRECTORIES_TO_COPY:
        source_path = os.path.join(templates_dir, dir_name)
        dest_path = os.path.join(output_dir, dir_name)

        if os.path.isdir(source_path):
            try:
                # Use shutil.copytree to recursively copy the directory
                shutil.copytree(source_path, dest_path, dirs_exist_ok=True)
                print(f"  Successfully copied directory: {dir_name}")
            except (IOError, shutil.Error) as e:
                print_red(f"Error copying directory {source_path}: {e}")
                success = False  # Mark as failure if a directory copy fails
        else:
            print_yellow(
                f"  Warning: Directory to copy not found, skipping: {source_path}"
            )

    # --- The existing logic for copying individual static files follows ---
    if not complete_config or not supported_resources:
        print_red("Cannot copy static files due to missing config.")
        return False

    for project in complete_config.get("projects", []):
        for category in ["producers", "consumers"]:
            for item in project.get(category, []):
                res_type = item.get("type")
                res_info = supported_resources.get(res_type, {})
                gen_config = res_info.get("generationConfig", {})
                folderName = gen_config.get("folderName")

                if not folderName:
                    continue

                tfvars_filename = gen_config.get("staticTfvarsFilename")
                if not tfvars_filename:
                    tfvars_filename = f"{folderName.split('/')[-1].lower()}.tfvars"

                source_path_file = os.path.join(
                    templates_dir, category, folderName, tfvars_filename
                )
                singular_category = category.rstrip("s")
                dest_path_file = os.path.join(
                    output_dir, singular_category, folderName, tfvars_filename
                )

                if (
                    os.path.exists(source_path_file)
                    and source_path_file not in copied_sources
                ):
                    try:
                        os.makedirs(os.path.dirname(dest_path_file), exist_ok=True)
                        shutil.copy(source_path_file, dest_path_file)
                        print(f"  Copied static file: {tfvars_filename}")
                        copied_sources.add(source_path_file)
                    except IOError as e:
                        print_red(f"Error copying static file {source_path_file}: {e}")
                        success = False
    return success


def main():
    """Parses arguments and controls the full interactive generation and deployment workflow."""
    help_description = """
Generates Terraform .tfvars and YAML configuration files, and optionally deploys GCP resources.

Operation Modes:
  --all: Full cycle = generate complete.json, generate TF/YAML configs, prompt for review, and optionally deploy.
         Requires user interaction to select basic JSON and confirm steps.
  
  --full-spec: Only generate the complete.json specification file.
               Requires user interaction to select basic JSON.

  --terraform: Only generate Terraform files from an existing complete.json.
               Requires user interaction to select an existing configuration directory.

  --terraform-apply: Generate Terraform .tfvars and YAML configuration files from an existing complete.json, then deploy.
                     Requires user interaction to select an existing configuration directory and confirm deployment.

  --apply: Only run the deployment script for an already generated configuration.
           Requires user interaction to select an existing configuration directory and confirm deployment.
"""
    parser = argparse.ArgumentParser(
        description=help_description,
        formatter_class=argparse.RawTextHelpFormatter,
        usage=(
            "config_generator.py [-h] [--all] [--full-spec] [--terraform]"
            " [--terraform-apply] [--apply]"
        ),
    )

    parser.add_argument(
        "--all",
        action="store_true",
        help="Run the full generation and optional deployment cycle.",
    )
    parser.add_argument(
        "--full-spec",
        action="store_true",
        help="Only generate the complete.json specification file.",
    )
    parser.add_argument(
        "--terraform",
        action="store_true",
        help="Only generate Terraform files from an existing complete.json.",
    )
    parser.add_argument(
        "--terraform-apply",
        action="store_true",
        help="Generate Terraform files and then optionally deploy.",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help=(
            "Only run the deployment script for an already generated" " configuration."
        ),
    )

    args = parser.parse_args()
    modes = [k for k, v in vars(args).items() if v]
    if not modes or len(modes) > 1:
        parser.print_help()
        sys.exit(1)

    if args.all or args.full_spec:
        if os.path.exists(MAIN_OUTPUT_DIR):
            print_red(
                f"Error: Config Generator has detected a configuration/ folder inside this repository at '{MAIN_OUTPUT_DIR}'."
            )
            print_red(
                "Destroy any existing infrastructure created using the CNCS repository and delete configuration/ folder before proceeding."
            )
            sys.exit(1)

        print("=" * 60)
        print(" CNCS Configuration Generator ".center(60))
        print("=" * 60)

        basic_path, config_name = select_basic_json()

        print("\n--- Step 2: Creating output directory and complete.json ---")
        os.makedirs(MAIN_OUTPUT_DIR, exist_ok=True)
        shutil.copy(
            basic_path, os.path.join(MAIN_OUTPUT_DIR, os.path.basename(basic_path))
        )

        complete_json_path = generate_config_from_path(
            basic_path,
            config_name,
            MAIN_OUTPUT_DIR,
            SUPPORTED_RESOURCES_PATH,
            SCHEMA_DIR,
            DEFAULTS_DIR,
        )

        if not complete_json_path:
            print_red("\nJSON Generation failed. Aborting.")
            shutil.rmtree(MAIN_OUTPUT_DIR, ignore_errors=True)
            sys.exit(1)
        print(f"\nJSON Generation successful. Output: {complete_json_path}")

        if args.full_spec:
            _open_file_in_default_app(complete_json_path)
            print("\n--full-spec mode selected. `complete.json` has been generated.")
            return

        if args.all:
            if not prompt_for_json_review(complete_json_path, config_name):
                print("\nUser aborted after reviewing complete.json. Exiting.")
                shutil.rmtree(MAIN_OUTPUT_DIR, ignore_errors=True)
                sys.exit(0)

            if generate_all_tf_files(complete_json_path, MAIN_OUTPUT_DIR):
                print(
                    "\n--- All Terraform configuration files have been successfully generated ---"
                )
                _copy_static_files(
                    load_json_file(complete_json_path),
                    load_json_file(SUPPORTED_RESOURCES_PATH),
                    TEMPLATES_BASE_DIR,
                    MAIN_OUTPUT_DIR,
                )
                run_terraform_fmt(MAIN_OUTPUT_DIR)
                open_and_advise_terraform_review()
                if prompt_for_final_deployment():
                    run_deployment()
                else:
                    print("\nDeployment skipped by user.")
            else:
                print_red("\nErrors during Terraform file generation. Aborting.")
                shutil.rmtree(MAIN_OUTPUT_DIR, ignore_errors=True)
                sys.exit(1)

    elif args.terraform:
        # Generate TF files from an existing config, with validation.
        config_dir, complete_json_path = find_existing_configuration(
            require_complete_json=True
        )

        if generate_all_tf_files(complete_json_path, config_dir):
            _copy_static_files(
                load_json_file(complete_json_path),
                load_json_file(SUPPORTED_RESOURCES_PATH),
                TEMPLATES_BASE_DIR,
                config_dir,
            )
            run_terraform_fmt(config_dir)
            print("\nTerraform file generation complete.")
        else:
            print_red("\nErrors during Terraform file generation.")
            sys.exit(1)

    elif args.terraform_apply:
        # Generate TF files and then deploy, with validation.
        config_dir, complete_json_path = find_existing_configuration(
            require_complete_json=True
        )

        if generate_all_tf_files(complete_json_path, config_dir):
            _copy_static_files(
                load_json_file(complete_json_path),
                load_json_file(SUPPORTED_RESOURCES_PATH),
                TEMPLATES_BASE_DIR,
                config_dir,
            )
            run_terraform_fmt(config_dir)
            open_and_advise_terraform_review()
            if prompt_for_final_deployment():
                run_deployment()
            else:
                print("\nDeployment skipped by user.")
        else:
            print_red("\nErrors during Terraform file generation.")
            sys.exit(1)

    elif args.apply:
        # Only deploy an existing config, with validation.
        config_dir, _ = find_existing_configuration(require_complete_json=False)

        print(f"Preparing to deploy configuration from: {config_dir}")
        if prompt_for_final_deployment():
            run_deployment()
        else:
            print("\nDeployment skipped by user.")


if __name__ == "__main__":
    main()
