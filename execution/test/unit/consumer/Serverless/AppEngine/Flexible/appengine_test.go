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
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../../../") // As provided
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Flexible")
	configFolderPath       = filepath.Join(projectRoot, "test/unit/consumer/Serverless/AppEngine/Flexible/config")
)

// Create the resources using the terrafrom consumer code
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

// Count the number of resources created expected vs actual
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
	expectedAddCount := 8

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

// Match Expected resources to Created resources to ensure correct resource creation
func TestTerraformModuleResourceAddressListMatch(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": configFolderPath,
		},
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	// Define the *expected* resources within the flexible module.  This is crucial.
	// We're testing the *flexible* module, so we list resources *it* creates.
	expectedResources := []string{
		"module.flexible_app_engine_instance[\"project-id2_test-service2\"].google_app_engine_flexible_app_version.flexible[\"test-service2\"]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_application.app[0]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_application_url_dispatch_rules.dispatch[0]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_domain_mapping.mapping[\"0\"]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_firewall_rule.firewall[\"0\"]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_flexible_app_version.flexible[\"test-service1\"]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_service_network_settings.network_settings[\"test-service1\"]",
		"module.flexible_app_engine_instance[\"project-id1_test-service1\"].google_app_engine_service_split_traffic.split_traffic[\"test-service1\"]",
	}

	planJSON := terraform.InitAndPlanAndShow(t, terraformOptions)
	plan, err := terraform.ParsePlanJSON(planJSON)
	if err != nil {
		t.Fatalf("Failed to parse plan JSON: %v", err)
	}

	// Get the actual resources from the plan.
	actualResources := make([]string, 0)
	for resourceAddress := range plan.ResourceChangesMap { // Corrected: use ResourceChangesMap
		actualResources = append(actualResources, resourceAddress)
	}

	// Use ElementsMatch for a set comparison (order doesn't matter).
	assert.ElementsMatch(t, expectedResources, actualResources, "Planned resources should match expected resources")
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
