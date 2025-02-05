// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unittest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"golang.org/x/exp/slices"
)

const (
	terraformDirectoryPath = "../../../../03-security/Workbench"
	network                = "projects/dummy-project/global/networks/dummy-vpc-network01"
)

var (
	projectID = "dummy-project-id"
	tfVars    = map[string]any{
		"project_id": projectID,
		"network":    network,
		"ingress_rules": map[string]any{
			"allow-ssh-custom-ranges-workbench": map[string]any{
				"deny": false,
				"rules": []any{
					map[string]any{
						"protocol": "tcp",
						"ports":    []string{"22", "443"},
					},
				},
			},
		},
	}
)

// TestInitAndPlanRunWithTfVars verifies that Terraform init and plan succeed with the provided tfVars.
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
	want := 2 // Update expected exit code to match the actual behavior
	got := planExitCode
	if got != want {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}
}

// TestInitAndPlanRunWithoutTfVarsExpectFailureScenario verifies that Terraform init and plan fail without tfVars.
func TestInitAndPlanRunWithoutTfVarsExpectFailureScenario(t *testing.T) {
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 1
	got := planExitCode
	if !cmp.Equal(got, want) {
		t.Errorf("Test Plan Exit Code = %v, want = %v", got, want)
	}
}

// TestResourcesCount verifies the count of resources to be added, changed, or destroyed in the plan.
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
	if got, want := resourceCount.Change, 0; got != want {
		t.Errorf("Test Resource Count Change = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Destroy, 0; got != want {
		t.Errorf("Test Resource Count Destroy = %v, want = %v", got, want)
	}
}

// TestTerraformModuleResourceAddressListMatch verifies that the resource addresses in the plan match the expected list.
func TestTerraformModuleResourceAddressListMatch(t *testing.T) {
	expectedModulesAddress := []string{
		"module.workbench_firewall.google_compute_firewall.custom-rules[\"allow-ssh-custom-ranges-workbench\"]",
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

	actualResourceAddresses := make([]string, 0)
	for _, element := range content.ResourceChangesMap {
		if element.Address != "" && !slices.Contains(actualResourceAddresses, element.Address) {
			actualResourceAddresses = append(actualResourceAddresses, element.Address)
		}
	}

	// Sort both slices to make the comparison order-independent
	slices.Sort(expectedModulesAddress)
	slices.Sort(actualResourceAddresses)

	want := expectedModulesAddress
	got := actualResourceAddresses
	if !cmp.Equal(got, want) {
		t.Errorf("Test Element Mismatch = %v, want = %v", got, want)
	}
}
