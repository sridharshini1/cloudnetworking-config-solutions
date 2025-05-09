/**
 * Copyright 2024 Google LLC
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

// Package for comparison operations
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform" // Terraform testing library
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/MIG")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/consumer/MIG/config")
)

var (

	// Terraform variables to be passed to the module.
	tfVars = map[string]any{
		"config_folder_path": configFolderPath,
	}
)

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

	if got, want := resourceCount.Add, 3; got != want {
		t.Errorf("Test Resource Count Add = %v, want = %v", got, want)
	}
}

func TestTerraformModuleMIGResourceAddressListMatch(t *testing.T) {
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

			expectedModuleAddresses[fmt.Sprintf("module.mig[\"%s\"]", config.InstanceName)] = struct{}{}          // Use the instance name
			expectedModuleAddresses[fmt.Sprintf("module.mig-template[\"%s\"]", config.InstanceName)] = struct{}{} // Use the instance name
		}
	}

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

	actualModuleAddresses := make(map[string]struct{}) // Use a map to ensure uniqueness
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.mig") ||
			strings.HasPrefix(element.ModuleAddress, "module.mig-template") {
			actualModuleAddresses[element.ModuleAddress] = struct{}{}
		}
	}

	fmt.Printf("Actual module addresses found in plan: %+v\n", actualModuleAddresses)

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
		if !assert.ElementsMatch(t, actualSlice, expectedSlice) {
			t.Errorf("Test Element Mismatch = %v, want = %v", actualSlice, expectedSlice)
		}
	} else {
		// If no modules expected, check if any actual addresses were found (should be none)
		if len(actualSlice) > 0 {
			t.Errorf("Unexpected module addresses found: %v", actualSlice)
		} else {
			t.Log("No modules expected, and none found in plan.")
		}
	}
}
