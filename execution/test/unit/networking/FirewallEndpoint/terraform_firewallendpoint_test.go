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

package unittest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var (
	// ADAPT THESE PATHS TO YOUR PROJECT STRUCTURE
	projectRootFE, _         = filepath.Abs("../../../../")
	terraformDirectoryPathFE = filepath.Join(projectRootFE, "02-networking/FirewallEndpoint")               // Path to the firewall endpoint consumer code
	configFolderPathFE       = filepath.Join(projectRootFE, "test/unit/networking/FirewallEndpoint/config") // Path to the TEST YAML files for firewall endpoints
)

var (
	// Terraform variables to be passed to the consumer.
	tfVarsFE = map[string]any{
		"config_folder_path": configFolderPathFE,
	}
)

// TestFirewallEndpointPlanExitCode verifies that the plan exits with a code of 2, indicating changes are planned.
func TestFirewallEndpointPlanExitCode(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathFE,
		Vars:         tfVarsFE,
		Reconfigure:  true,
		PlanFilePath: "./plan_fe",
		NoColor:      true,
	})

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 2, planExitCode, "Test Plan Exit Code: Expected changes to be applied")
}

// TestFirewallEndpointResourcesCount verifies the number of resources to be added by the plan.
func TestFirewallEndpointResourcesCount(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathFE,
		Vars:         tfVarsFE,
		Reconfigure:  true,
		PlanFilePath: "./plan_fe",
		NoColor:      true,
	})

	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)

	// ASSUMPTION: Your test config folder contains YAML files that together create 3 resources.
	// Adjust this number based on your actual test files.
	expectedResourceCount := 3
	assert.Equal(t, expectedResourceCount, resourceCount.Add, "Test Resource Count Add: Unexpected number of resources to be created")
}

// TestFirewallEndpointModuleAddressListMatch verifies that a module instance is planned for each YAML config file.
func TestFirewallEndpointModuleAddressListMatch(t *testing.T) {
	t.Parallel()
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathFE,
		Vars:         tfVarsFE,
		Reconfigure:  true,
		PlanFilePath: "./plan_fe",
		NoColor:      true,
	})

	// Get the expected module keys directly from the filenames.
	expectedModuleKeys := []string{}
	files, err := os.ReadDir(configFolderPathFE)
	assert.NoError(t, err, "Error reading config directory")

	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
				key := strings.TrimSuffix(filename, ".yaml")
				key = strings.TrimSuffix(key, ".yml")
				expectedModuleKeys = append(expectedModuleKeys, key)
			}
		}
	}
	assert.NotEmpty(t, expectedModuleKeys, "No YAML files found in the test config directory")

	// Build the expected list of module addresses from the keys.
	expectedModuleAddresses := []string{}
	for _, key := range expectedModuleKeys {
		expectedModuleAddresses = append(expectedModuleAddresses, fmt.Sprintf("module.firewall_endpoints[\"%s\"]", key))
	}

	// Parse the plan and find the actual module addresses being created.
	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	assert.NoError(t, err, "Error parsing plan JSON")

	actualModuleAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.firewall_endpoints") &&
			!slices.Contains(actualModuleAddresses, element.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, element.ModuleAddress)
		}
	}

	assert.ElementsMatch(t, expectedModuleAddresses, actualModuleAddresses, "The planned module addresses do not match the expected addresses from YAML files.")
}
