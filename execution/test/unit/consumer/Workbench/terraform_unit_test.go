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
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Workbench")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/consumer/Workbench/config")
)

var (
	// Terraform variables to be passed to the module.
	tfVars = map[string]any{
		"config_folder_path": configFolderPath,
	}
)

// TestInitAndPlanRunWithTfVars verifies that 'terraform init' and 'terraform plan'
// execute successfully with the provided tfVars and checks the expected exit code.
func TestInitAndPlanRunWithTfVars(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	// Run 'terraform init' and 'terraform plan', get the exit code.
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect changes to be applied
	got := planExitCode

	// Check if the actual exit code matches the expected one.
	if got != want {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}
}

// TestResourcesCount verifies the number of resources to be added by the Terraform plan.
func TestResourcesCount(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	// Initialize and create a plan, then parse the resource count.
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)

	if got, want := resourceCount.Add, 1; got != want {
		t.Errorf("Test Resource Count Add = %v, want = %v", got, want)
	}
}

// TestTerraformModuleWorkbenchResourceAddressListMatch verifies that the Terraform plan contains the expected module addresses for workbench instances based on the configuration YAML files.
func TestTerraformModuleWorkbenchResourceAddressListMatch(t *testing.T) {
	expectedModuleAddresses := make(map[string]struct{}) // Use a map to ensure uniqueness

	// Read the YAML files in the configuration folder.
	yamlFiles, err := os.ReadDir(configFolderPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Extract instance names from the YAML files.
	for _, file := range yamlFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			// Read YAML content for the instance name
			yamlData, err := os.ReadFile(filepath.Join(configFolderPath, file.Name()))
			if err != nil {
				t.Fatal(err.Error())
			}

			var config struct {
				InstanceName string `yaml:"name"` // Adjust this field according to your YAML structure
			}

			err = yaml.Unmarshal(yamlData, &config)
			if err != nil {
				t.Fatal(err.Error())
			}

			expectedModuleAddresses[fmt.Sprintf("module.workbench_instance[\"%s\"]", config.InstanceName)] = struct{}{}
		}
	}

	// Log the expected module addresses for debugging
	t.Logf("Expected module addresses: %+v", expectedModuleAddresses)

	// Initialize Terraform and generate a plan.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	if err != nil {
		t.Fatal(err.Error())
	}

	actualModuleAddresses := make(map[string]struct{})
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.workbench_instance") {
			actualModuleAddresses[element.ModuleAddress] = struct{}{}
		}
	}

	// Log the actual module addresses for debugging
	t.Logf("Actual module addresses found in plan: %+v", actualModuleAddresses)

	expectedSlice := make([]string, 0, len(expectedModuleAddresses))
	for address := range expectedModuleAddresses {
		expectedSlice = append(expectedSlice, address)
	}

	actualSlice := make([]string, 0, len(actualModuleAddresses))
	for address := range actualModuleAddresses {
		actualSlice = append(actualSlice, address)
	}

	if len(expectedSlice) > 0 {
		// Compare expected and actual addresses
		assert.ElementsMatch(t, actualSlice, expectedSlice)
	} else {
		// If no modules expected, check if any actual addresses were found (should be none)
		if len(actualSlice) > 0 {
			t.Errorf("Unexpected module addresses found: %v", actualSlice)
		} else {
			t.Log("No modules expected, and none found in plan.")
		}
	}
}
