/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package unittest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const (
	// The relative path from this unit test to the Terraform module under test.
	terraformDirectoryPath = "../../../../../../../execution/07-consumer-load-balancing/Network/Passthrough/Internal"
)

// TestInitAndValidate checks if the module is syntactically valid.
func TestInitAndValidate(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		NoColor:      true,
	})

	// Run "terraform init" and "terraform validate".
	// Terratest will fail the test if there are any errors.
	terraform.InitAndValidate(t, terraformOptions)
}

// TestPlanFailsWithInvalidConfig tests that the Terraform plan fails when a YAML
// config is missing required top-level attributes that are accessed in locals.tf.
func TestPlanFailsWithInvalidConfig(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for our invalid config.
	tempDir, err := os.MkdirTemp("", "test-invalid-config-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up the temp directory after the test.

	// Define an invalid YAML content that is missing the 'project', 'network', and 'subnetwork' keys
	// which are accessed directly in your locals.tf file.
	invalidYAML := []byte(`
name: invalid-lb
region: us-central1
# This config is invalid because required keys are missing.
`)

	// Write the invalid YAML to a file in our temporary directory.
	invalidFilePath := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidFilePath, invalidYAML, 0644)
	assert.NoError(t, err)

	// Define Terraform options, pointing config_folder_path to our temporary directory.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars: map[string]interface{}{
			"config_folder_path": tempDir,
		},
		Reconfigure: true,
		Lock:        true,
		NoColor:     true,
	})

	// Run `terraform plan` and expect it to fail with an exit code of 1
	// because the module will try to access attributes that do not exist in the invalid YAML.
	exitCode := terraform.PlanExitCode(t, terraformOptions)

	wantCode := 1
	if exitCode != wantCode {
		t.Errorf("Expected plan to fail with exit code %d due to invalid config, but got %d", wantCode, exitCode)
	}
}
