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
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../../../") // As provided
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Standard")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/consumer/Serverless/AppEngine/Standard/config")
)

func TestInitAndPlan(t *testing.T) {
	// Construct the terraform options.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath, // Use the YAML configs
		},
		Reconfigure:  true, // Important for switching between test cases
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	// Run 'terraform init' and 'terraform plan', get the exit code.
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect no  changes to be applied
	got := planExitCode

	// Check if the actual exit code matches the expected one.
	if got != want {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}

}
func TestResourcesCount(t *testing.T) {
	// Construct the terraform options.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath, // Use the YAML configs
		},
		Reconfigure:  true, // Important for switching between test cases
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	// Initialize and create a plan, then parse the resource count.
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)
	// Initialize expectedAddCount, expectedChangeCount, and expectedDestroyCount to zero
	expectedAddCount := 6

	// Read and process YAML files to calculate expected counts
	files, err := ioutil.ReadDir(configFolderPath)
	if err != nil {
		t.Fatalf("Failed to read config folder: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			yamlFilePath := filepath.Join(configFolderPath, file.Name())
			yamlFile, err := ioutil.ReadFile(yamlFilePath)
			if err != nil {
				t.Fatalf("Failed to read YAML file %s: %v", yamlFilePath, err)
			}

			var config map[string]interface{}
			err = yaml.Unmarshal(yamlFile, &config)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML file %s: %v", yamlFilePath, err)
			}
			// Check if the service needs to be created
			if createVersion, ok := config["create_app_version"].(bool); ok && !createVersion {
				continue // Skip incrementing counts if create_app_version is false
			}
		}
	}
	//check for count mismatch
	if got, want := resourceCount.Add, expectedAddCount; got != want {
		t.Errorf("Test Resource Count Add = %v, want = %v", got, want)
	}
}

func TestTerraformModuleResourceAddressListMatch(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath, // Use the YAML configs
		},
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	localServiceMap := make(map[string]map[string]interface{}) // Initialize the map
	err := filepath.Walk(configFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".yaml" {
			// Read the YAML file
			yamlFile, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// Unmarshal YAML into a map
			var serviceData map[string]interface{}
			err = yaml.Unmarshal(yamlFile, &serviceData)
			if err != nil {
				return err
			}

			// Extract the service name and add to the map
			serviceName := filepath.Base(path) // Assumes filename is the instance name
			serviceName = strings.TrimSuffix(serviceName, filepath.Ext(serviceName))
			localServiceMap[serviceName] = serviceData
		}
		return nil
	})
	if err != nil {
		t.Errorf("Error reading YAML files: %s", err)
	}

	if len(localServiceMap) == 0 { // Check if the map is empty
		t.Error("No services found in YAML files. Make sure the files exist and are correctly formatted.")
	}
	expectedModuleAddresses := []string{} // Start with an empty slice
	expectedModuleAddresses = append(expectedModuleAddresses, "module.appengine_standard_instance[\"instance1\"]", "module.appengine_standard_instance[\"instance2\"]")

	planStruct := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planStruct)
	if err != nil {
		t.Errorf("Error parsing plan JSON: %s", err) // Fail fast if parsing errors occur
	}

	actualModuleAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if strings.HasPrefix(element.ModuleAddress, "module.appengine_standard") &&
			!slices.Contains(actualModuleAddresses, element.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, element.ModuleAddress)
		}
	}

	assert.ElementsMatch(t, expectedModuleAddresses, actualModuleAddresses)
}

// TestInitAndPlanFailure tests for failure scenarios with invalid inputs.
func TestInitAndPlanFailure(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath,           // Use ALL configs, including invalid
			"project_id":         "dummy-project-id-failure", //Required for plan to fail
		},
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	exitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	assert.Equal(t, 1, exitCode, "Expected Terraform to fail with exit code 1")
}
