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
	projectRoot, _ = filepath.Abs("../../../../../../../")

	// Path to the Terraform directory for the Network Load Balancer
	terraformDirectoryPathNLB = filepath.Join(projectRoot, "execution/07-consumer-load-balancing/Network/Passthrough/External")

	// Path to the YAML configuration folder for NLB tests
	configFolderPathNLB = filepath.Join(projectRoot, "execution/test/unit/consumer-load-balancing/Network/Passthrough/External/config")
)

// Terraform variables to be passed to the NLB module.
var tfVarsNLB = map[string]interface{}{
	"config_folder_path": configFolderPathNLB,
}

/*
TestInitAndPlanRunWithTfVarsNLB tests Terraform initialization and planning
for the Network Load Balancer module with specified variables.
It expects an exit code of 2, indicating that changes are planned.
If the exit code differs, it logs an error.
*/
func TestInitAndPlanRunWithTfVarsNLB(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathNLB,
		Vars:         tfVarsNLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-nlb", // Use a distinct plan file name
		NoColor:      true,
	})

	// Run 'terraform init' and 'terraform plan', get the exit code.
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect changes to be applied (exit code 2 means plan has changes)
	got := planExitCode

	// Check if the actual exit code matches the expected one.
	if got != want {
		t.Errorf("TestInitAndPlanRunWithTfVarsNLB: Plan Exit Code = %v, want = %v", got, want)
	}
}

/*
TestResourcesCountNLB verifies the number of resources planned by Terraform for the NLB module.
It initializes Terraform with specified variables, creates a plan, and checks
that the total resource count to be added matches the expected value.
An error is logged if the count differs.

The expected count depends on the number of NLBs defined in your test YAML files
and the resources created per NLB by the module (typically a backend service,
a health check, and at least one forwarding rule).
If you have one YAML file defining one NLB, and it creates these 3 main resources,
the count would be 3.
*/
func TestResourcesCountNLB(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathNLB,
		Vars:         tfVarsNLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-nlb", // Use a distinct plan file name
		NoColor:      true,
	})

	// Initialize and create a plan, then parse the resource count.
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)

	// --- Determine the expected resource count ---
	// Count the number of NLB configurations in your test YAML files.
	yamlFiles, err := os.ReadDir(configFolderPathNLB)
	if err != nil {
		t.Fatalf("Failed to read NLB config directory %s: %v", configFolderPathNLB, err)
	}

	numberOfNLBs := 0
	for _, file := range yamlFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			// Basic check: ensure file is not empty or a "hidden" config file
			if !strings.HasPrefix(file.Name(), "_") {
				numberOfNLBs++
			}
		}
	}
	if numberOfNLBs == 0 {
		t.Logf("No YAML configuration files found in %s. Ensure your test setup is correct.", configFolderPathNLB)
	}

	// Expected resources per NLB instance (typically 1 backend service, 1 health check, 1 forwarding rule by default)
	resourcesPerNLB := 3
	expectedResourceAddCount := numberOfNLBs * resourcesPerNLB

	if got, want := resourceCount.Add, expectedResourceAddCount; got != want {
		// For debugging, show the plan output
		planJSON := terraform.Show(t, terraformOptions)
		t.Logf("Plan output: %s", planJSON) // Log the plan for inspection
		t.Errorf("TestResourcesCountNLB: Resource Count Add = %v, want = %v (based on %d NLB configs)", got, want, numberOfNLBs)
	}
	if got, want := resourceCount.Change, 0; got != want {
		t.Errorf("TestResourcesCountNLB: Resource Count Change = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Destroy, 0; got != want {
		t.Errorf("TestResourcesCountNLB: Resource Count Destroy = %v, want = %v", got, want)
	}
}

/*
TestTerraformModuleNLBResourceAddressListMatch checks that the module addresses for
Network Load Balancer resources in the Terraform plan match the expected addresses
derived from YAML configuration files. It reads NLB names from YAML files,
initializes Terraform, generates a plan, and compares the expected and actual
module addresses for 'module.nlb_passthrough_ext'.
An error is logged if there are mismatches or unexpected addresses.
*/
func TestTerraformModuleNLBResourceAddressListMatch(t *testing.T) {
	expectedModuleAddresses := make(map[string]struct{}) // Use a map to ensure uniqueness and simplify lookups

	// Read the YAML files in the configuration folder.
	yamlFiles, err := os.ReadDir(configFolderPathNLB)
	if err != nil {
		t.Fatalf("Failed to read NLB config directory %s: %v", configFolderPathNLB, err)
	}

	foundYAML := false
	// Extract NLB names from the YAML files.
	for _, file := range yamlFiles {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			if strings.HasPrefix(file.Name(), "_") { // Skip template/hidden files
				continue
			}
			foundYAML = true
			yamlData, err := os.ReadFile(filepath.Join(configFolderPathNLB, file.Name()))
			if err != nil {
				t.Fatalf("Failed to read YAML file %s: %v", file.Name(), err)
			}

			var config struct {
				Name string `yaml:"name"` // Assumes your YAML structure has a 'name' field for the NLB
			}

			err = yaml.Unmarshal(yamlData, &config)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML from file %s: %v", file.Name(), err)
			}
			if config.Name == "" {
				t.Logf("Warning: NLB name is empty in YAML file %s. This might lead to unexpected module addresses.", file.Name())
				// Handle as per your requirements: skip, error, or allow if empty names are valid keys in your for_each
			}

			// Construct the expected module address for the Network Load Balancer module
			expectedModuleAddresses[fmt.Sprintf("module.nlb_passthrough_ext[\"%s\"]", config.Name)] = struct{}{}
		}
	}

	if !foundYAML {
		t.Logf("No YAML configuration files found in %s for TestTerraformModuleNLBResourceAddressListMatch. "+
			"If this is unexpected, check your configFolderPathNLB and test setup.", configFolderPathNLB)
	}

	// Initialize Terraform and generate a plan.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPathNLB,
		Vars:         tfVarsNLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-nlb-addressmatch", // Use a distinct plan file name
		NoColor:      true,
	})

	planJSON := terraform.InitAndPlanAndShow(t, terraformOptions) // Shows JSON output of plan
	content, err := terraform.ParsePlanJSON(planJSON)             // planStruct was planJSON string, content needs to be parsed from it
	if err != nil {
		t.Fatalf("Failed to parse plan JSON: %v", err)
	}

	actualModuleAddresses := make(map[string]struct{}) // Use a map for actual addresses as well
	for _, resourceChange := range content.ResourceChangesMap {
		moduleAddr := resourceChange.ModuleAddress

		if moduleAddr != "" { // Ensure it's not a root module resource
			parts := strings.Split(moduleAddr, ".")
			// We are looking for module instances like "module.nlb_passthrough_ext[\"some_key\"]"
			// Such an address will be split into two parts: "module" and "nlb_passthrough_ext[\"some_key\"]"
			if len(parts) == 2 && parts[0] == "module" && strings.HasPrefix(parts[1], "nlb_passthrough_ext[") {
				// This is a direct instance of the module we are interested in.
				actualModuleAddresses[moduleAddr] = struct{}{}
			}
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

	if len(expectedSlice) == 0 && len(actualSlice) == 0 && !foundYAML { // Adjusted condition slightly for clarity
		t.Log("No NLB YAMLs found and no NLB module instances found in plan. Test passes.")
		return
	}

	if !assert.ElementsMatch(t, actualSlice, expectedSlice) {
		// Log for more detailed debugging if assert fails
		t.Logf("Expected Module Addresses (from YAMLs): %v", expectedSlice)
		t.Logf("Actual Module Addresses (from Plan): %v", actualSlice)
		t.Logf("Full Plan JSON for TestTerraformModuleNLBResourceAddressListMatch: %s", planJSON)
		t.Errorf("TestTerraformModuleNLBResourceAddressListMatch: Mismatch in module instance addresses.\nExpected: %v\nActual:   %v", expectedSlice, actualSlice)
	}
}
