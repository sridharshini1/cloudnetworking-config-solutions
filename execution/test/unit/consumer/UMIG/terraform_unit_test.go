/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * You may not use this file except in compliance with the License.
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
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/UMIG")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/consumer/UMIG/config")
)

var (
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

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect changes to be applied
	got := planExitCode

	if got != want {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}
}

func TestResourcesCount(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)

	if got, want := resourceCount.Add, 1; got != want {
		t.Errorf("Test Resource Count Add = %v, want = %v", got, want)
	}
}

func TestTerraformModuleUMIGResourceAddressListMatch(t *testing.T) {
	expectedModuleAddresses := make(map[string]struct{})

	// Read the YAML files in the configuration folder.
	yamlFiles, err := os.ReadDir(configFolderPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Use an incrementing index to match the for_each key in Terraform.
	idx := 0
	for _, file := range yamlFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			yamlData, err := os.ReadFile(filepath.Join(configFolderPath, file.Name()))
			if err != nil {
				t.Fatal(err.Error())
			}

			var config struct {
				Name string `yaml:"name"`
			}

			err = yaml.Unmarshal(yamlData, &config)
			if err != nil {
				t.Fatal(err.Error())
			}

			expectedModuleAddresses[fmt.Sprintf(`module.umig["umig-%d"]`, idx)] = struct{}{}
			idx++
		}
	}

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
		if strings.HasPrefix(element.ModuleAddress, "module.umig") {
			actualModuleAddresses[element.ModuleAddress] = struct{}{}
		}
	}

	expectedSlice := make([]string, 0, len(expectedModuleAddresses))
	for address := range expectedModuleAddresses {
		expectedSlice = append(expectedSlice, address)
	}

	actualSlice := make([]string, 0, len(actualModuleAddresses))
	for address := range actualModuleAddresses {
		actualSlice = append(actualSlice, address)
	}

	if len(expectedSlice) > 0 {
		if !assert.ElementsMatch(t, actualSlice, expectedSlice) {
			t.Errorf("Test Element Mismatch = %v, want = %v", actualSlice, expectedSlice)
		}
	} else {
		if len(actualSlice) > 0 {
			t.Errorf("Unexpected module addresses found: %v", actualSlice)
		} else {
			t.Log("No modules expected, and none found in plan.")
		}
	}
}

// TestInitAndPlanFailure tests for failure scenarios with invalid inputs.
func TestInitAndPlanFailure(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath,           // Use ALL configs, including invalid
			"project_id":         "dummy-project-id-failure", //Required for plan
		},
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	exitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 1, exitCode, "Expected Terraform to fail with exit code 1")
}
