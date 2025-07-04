// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
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
	terraformSSLCertificateDirectoryPath = "../../../../../../03-security/Certificates/Compute-SSL-Certs/Google-Managed/"
	sslCertificateName                   = "test-managed-ssl-cert"
)

var (
	sslProjectID = "dummy-ssl-project-id"
	sslTfVars    = map[string]any{
		"project_id":                  sslProjectID,
		"ssl_certificate_name":        sslCertificateName,
		"ssl_certificate_description": "Test SSL certificate managed by Terraform",
		"ssl_managed_domains": []any{
			map[string]any{
				"domains": []string{"test.example.com", "www.test.example.com"},
			},
		},
	}
)

// TestSSLCertInitAndPlanRunWithTfVars tests terraform init & plan with valid TF variables.
// Expects a successful plan with changes (exit code 2).
func TestSSLCertInitAndPlanRunWithTfVars(t *testing.T) {

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformSSLCertificateDirectoryPath,
		Vars:         sslTfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan", // Use a distinct plan file path
		NoColor:      true,
	})

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // 0=no changes, 1=error, 2=changes
	got := planExitCode
	if got != want {
		t.Errorf("TestSSLCertInitAndPlanRunWithTfVars: Plan Exit Code = %v, want = %v", got, want)
	}
}

// TestSSLCertInitAndPlanRunWithoutTfVarsExpectFailure tests terraform init & plan
// without required TF variables. Expects a failure (exit code 1).
func TestSSLCertInitAndPlanRunWithoutTfVarsExpectFailure(t *testing.T) {

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformSSLCertificateDirectoryPath,
		// Vars:         // Intentionally omit tfVars
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan", // Use a distinct plan file path
		NoColor:      true,
	})

	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 1
	got := planExitCode
	if !cmp.Equal(got, want) { // Using cmp.Equal is fine, direct got != want is also fine for integers
		t.Errorf("TestSSLCertInitAndPlanRunWithoutTfVarsExpectFailure: Plan Exit Code = %v, want = %v", got, want)
	}
}

// TestSSLCertResourcesCount tests the number of resources to be added, changed, or destroyed.
// Assumes the module creates one primary SSL certificate resource.
func TestSSLCertResourcesCount(t *testing.T) {

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformSSLCertificateDirectoryPath,
		Vars:         sslTfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan", // Use a distinct plan file path
		NoColor:      true,
	})

	planData := terraform.InitAndPlan(t, terraformOptions) // Runs InitAndPlan and returns *terraform.PlanStruct
	resourceCount := terraform.GetResourceCount(t, planData)

	if got, want := resourceCount.Add, 1; got != want {
		t.Errorf("TestSSLCertResourcesCount: Resource Count Add = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Change, 0; got != want {
		t.Errorf("TestSSLCertResourcesCount: Resource Count Change = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Destroy, 0; got != want {
		t.Errorf("TestSSLCertResourcesCount: Resource Count Destroy = %v, want = %v", got, want)
	}
}

// TestSSLCertTerraformModuleResourceAddressListMatch tests if the expected SSL module address
// is present in the Terraform plan.
func TestSSLCertTerraformModuleResourceAddressListMatch(t *testing.T) {

	expectedModulesAddress := []string{"module.ssl_certificate"}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformSSLCertificateDirectoryPath,
		Vars:         sslTfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan",
		NoColor:      true,
	})

	planOutputJSON := terraform.InitAndPlanAndShow(t, terraformOptions) // This returns a string (JSON output)
	planData, err := terraform.ParsePlanJSON(planOutputJSON)            // Parse the string to *terraform.PlanStruct
	if err != nil {
		t.Fatalf("TestSSLCertTerraformModuleResourceAddressListMatch: Failed to parse plan JSON: %v", err)
	}

	actualModuleAddresses := make([]string, 0)
	for _, rc := range planData.ResourceChangesMap { // Use planData here
		if rc.ModuleAddress != "" && !slices.Contains(actualModuleAddresses, rc.ModuleAddress) {
			actualModuleAddresses = append(actualModuleAddresses, rc.ModuleAddress)
		}
	}

	slices.Sort(actualModuleAddresses)
	slices.Sort(expectedModulesAddress)

	if !cmp.Equal(actualModuleAddresses, expectedModulesAddress) {
		t.Errorf("TestSSLCertTerraformModuleResourceAddressListMatch: Module Address Mismatch = diff: %v\nGot: %v\nWant: %v", cmp.Diff(actualModuleAddresses, expectedModulesAddress), actualModuleAddresses, expectedModulesAddress)
	}
}

// TestSSLCertificateAttributes tests specific attributes of the planned SSL certificate.
func TestSSLCertificateAttributes(t *testing.T) {

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformSSLCertificateDirectoryPath,
		Vars:         sslTfVars,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan", // This plan file will be created by InitAndPlanAndShow
		NoColor:      true,
	})

	// Step 1: Get the plan output as a JSON string using InitAndPlanAndShow.
	// This function runs init, then plan, then show, and returns the 'show' output as a string.
	planOutputJSON := terraform.InitAndPlanAndShow(t, terraformOptions)

	// Step 2: Parse the JSON string into a *terraform.PlanStruct.
	// The 'planData' variable will hold the structured plan information.
	planData, err := terraform.ParsePlanJSON(planOutputJSON)
	if err != nil {
		// If parsing fails, the test cannot proceed with attribute checks.
		t.Fatalf("TestSSLCertificateAttributes: Failed to parse plan JSON: %v", err)
	}

	// Step 3: Now use 'planData' (which is *terraform.PlanStruct) for all subsequent operations.
	resourceAddress := "module.ssl_certificate.google_compute_managed_ssl_certificate.ssl_cert"

	// Pass 'planData' (the *terraform.PlanStruct) to this function.
	// This should resolve: "cannot use plan (variable of type string) as *terraform.PlanStruct value"
	terraform.RequirePlannedValuesMapKeyExists(t, planData, resourceAddress)

	// Access ResourcePlannedValuesMap from 'planData' (the *terraform.PlanStruct).
	// This should resolve: "plan.ResourcePlannedValuesMap undefined (type string has no field or method ResourcePlannedValuesMap)"
	certificateResource := planData.ResourcePlannedValuesMap[resourceAddress]
	if certificateResource == nil {
		t.Fatalf("TestSSLCertificateAttributes: Resource %s not found in planned values map of the planData", resourceAddress)
	}

	// Check "name" attribute
	gotName, nameOk := certificateResource.AttributeValues["name"].(string)
	if !nameOk {
		if certificateResource.AttributeValues["name"] == nil {
			t.Fatalf("TestSSLCertificateAttributes: 'name' attribute is nil for resource %s. Expected a string.", resourceAddress)
		}
		t.Fatalf("TestSSLCertificateAttributes: Could not assert 'name' attribute as string for resource %s. Actual type: %T, Value: %v", resourceAddress, certificateResource.AttributeValues["name"], certificateResource.AttributeValues["name"])
	}
	if gotName != sslCertificateName {
		t.Errorf("TestSSLCertificateAttributes: SSL Certificate name = %q, want = %q", gotName, sslCertificateName)
	}

	// Check "managed" block and "domains"
	managedBlockAny, managedBlockOk := certificateResource.AttributeValues["managed"].([]any)
	if !managedBlockOk || len(managedBlockAny) == 0 {
		if !managedBlockOk {
			val := certificateResource.AttributeValues["managed"]
			t.Fatalf("TestSSLCertificateAttributes: 'managed' attribute is not []any for resource %s. Actual type: %T, Value: %v", resourceAddress, val, val)
		}
		t.Fatalf("TestSSLCertificateAttributes: 'managed' block not found or empty for resource %s", resourceAddress)
	}

	managedConfig, configOk := managedBlockAny[0].(map[string]any)
	if !configOk {
		val := managedBlockAny[0]
		t.Fatalf("TestSSLCertificateAttributes: 'managed' block content [0] is not map[string]any for resource %s. Actual type: %T, Value: %v", resourceAddress, val, val)
	}

	domainsAny, domainsOk := managedConfig["domains"].([]any)
	if !domainsOk {
		val := managedConfig["domains"]
		t.Fatalf("TestSSLCertificateAttributes: 'domains' not found or not a list ([]any) in managed block for resource %s. Actual type: %T, Value: %v", resourceAddress, val, val)
	}

	var actualDomains []string
	for _, d := range domainsAny {
		ds, ok := d.(string)
		if !ok {
			t.Fatalf("TestSSLCertificateAttributes: A domain in the 'domains' list is not a string. Value: %v, Type: %T", d, d)
		}
		actualDomains = append(actualDomains, ds)
	}

	expectedDomainsStringSlice := []string{"test.example.com", "www.test.example.com"}
	if !cmp.Equal(actualDomains, expectedDomainsStringSlice) {
		t.Errorf("TestSSLCertificateAttributes: SSL Certificate domains mismatch (-got +want):\n%s", cmp.Diff(actualDomains, expectedDomainsStringSlice))
	}
}
