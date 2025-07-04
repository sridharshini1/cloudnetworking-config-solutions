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

package integrationtest

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

var (
	// Path to the Terraform configuration that uses the SSL certificate module
	terraformSslModulePath    = "../../../../../../03-security/Certificates/Compute-SSL-Certs/Google-Managed/"
	projectID                 = os.Getenv("TF_VAR_project_id")
	sslCertificateTypeManaged = "MANAGED"
	sslCertificatePathPrefix  = "/global/sslCertificates/"
)

// CertificateDetails struct to hold relevant information from gcloud describe output
type CertificateDetails struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ManagedDomains    []string `json:"managed.domains"`
	SelfLink          string   `json:"selfLink"`
	ManagedStatus     string   `json:"managed.status"`
	CreationTimestamp string   `json:"creationTimestamp"`
}

func TestCreateGoogleManagedSslCertificate(t *testing.T) {
	t.Parallel() // Mark test as parallelizable

	// Ensure projectID is set
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set.")
	}

	uniqueID := rand.Int() // Use uniqueID to ensure unique resource names
	sslCertificateName := fmt.Sprintf("test-managed-cert-%d", uniqueID)
	domainName := fmt.Sprintf("terratest-managed-cert-%d.example.com", uniqueID)

	// Variables for the SSL Certificate Terraform module
	tfSslVars := map[string]interface{}{
		"project_id":           projectID,
		"ssl_certificate_name": sslCertificateName,
		"ssl_certificate_type": sslCertificateTypeManaged,
		"ssl_managed_domains": []map[string]interface{}{
			{
				"domains": []string{domainName},
			},
		},
	}

	terraformSslOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:         terraformSslModulePath,
		Vars:                 tfSslVars,
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})

	// Defer destruction of SSL certificate created by Terraform
	defer terraform.Destroy(t, terraformSslOptions)

	// Create SSL certificate
	t.Logf("Applying Terraform configuration for Google-Managed SSL certificate: %s", sslCertificateName)
	terraform.InitAndApply(t, terraformSslOptions)

	// 1. Verify SSL certificate creation and get its self_link from Terraform output
	createdSslCertSelfLink := terraform.Output(t, terraformSslOptions, "managed_ssl_certificate_self_link")
	assert.NotEmpty(t, createdSslCertSelfLink, "SSL certificate self_link should not be empty")
	t.Logf("Created SSL Certificate Self Link: %s", createdSslCertSelfLink)

	// 2. Get detailed certificate properties using gcloud describe
	// We'll retry this in case the resource isn't immediately describe-able after creation.
	var certDetails *CertificateDetails
	_, err := retry.DoWithRetryE(t, "Get SSL Certificate Details", 5, 10*time.Second, func() (string, error) {
		details, describeErr := getSslCertificateDetails(t, projectID, sslCertificateName)
		if describeErr != nil {
			return "", fmt.Errorf("failed to describe cert: %w", describeErr)
		}
		certDetails = details // Assign to outer variable
		return "Successfully described certificate", nil
	})
	assert.NoError(t, err, "Failed to get SSL certificate details after retries")
	assert.NotNil(t, certDetails, "Certificate details should not be nil after retrieval")

	// 3. Assert on key properties of the created certificate
	// Properties that should be present immediately after creation (even in PROVISIONING)

	t.Logf("Expected: %s", sslCertificateName)
	t.Logf("Actual: %s", certDetails.Name)
	assert.Equal(t, sslCertificateName, certDetails.Name, "Assertion Failed: Certificate name mismatch")
	t.Logf("Validation complete for property: Name")

	t.Logf("Expected: MANAGED")
	t.Logf("Actual: %s", certDetails.Type)
	assert.Equal(t, sslCertificateTypeManaged, certDetails.Type, "Assertion Failed: Certificate type should be MANAGED")
	t.Logf("Validation complete for property: Type")

	t.Logf("Expected: %s (to be contained within ManagedDomains)", domainName)
	t.Logf("Actual: %v", certDetails.ManagedDomains)
	assert.Contains(t, certDetails.ManagedDomains, domainName, "Assertion Failed: Certificate should contain the specified managed domain")
	t.Logf("Validation complete for property: ManagedDomains")

	expectedSelfLinkPrefix := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/", projectID)
	t.Logf("Expected (prefix): %s", expectedSelfLinkPrefix)
	t.Logf("Actual: %s", certDetails.SelfLink)
	assert.True(t, strings.HasPrefix(certDetails.SelfLink, expectedSelfLinkPrefix), "Assertion Failed: Self link should start with correct project path")
	t.Logf("Validation complete for property: SelfLink prefix")

	t.Logf("Expected: Not empty")
	t.Logf("Actual: %s", certDetails.CreationTimestamp)
	assert.NotEmpty(t, certDetails.CreationTimestamp, "Assertion Failed: CreationTimestamp should not be empty")
	t.Logf("Validation complete for property: CreationTimestamp")

	t.Logf("Expected: Not empty")
	t.Logf("Actual: %s", certDetails.ManagedStatus)
	assert.NotEmpty(t, certDetails.ManagedStatus, "Assertion Failed: ManagedStatus should not be empty")
	t.Logf("Validation complete for property: ManagedStatus")

	t.Logf("Expected: SelfLink to contain '/global/sslCertificates/'")
	t.Logf("Actual: %s", certDetails.SelfLink)
	assert.True(t, strings.Contains(certDetails.SelfLink, sslCertificatePathPrefix), "Assertion Failed: Self link should indicate global SSL certificate. Expected to contain: %s, Actual: %s", sslCertificatePathPrefix, certDetails.SelfLink)
	t.Logf("Validation complete for property: SelfLink global indicator (Expected to contain: %s, Actual: %s)", sslCertificatePathPrefix, certDetails.SelfLink)

	t.Logf("Initial managed status of %s: %s", sslCertificateName, certDetails.ManagedStatus)

	// 4. Wait for the certificate to be in a PROVISIONING or ACTIVE state
	t.Logf("Waiting for SSL certificate %s (Type: %s) to be PROVISIONING or ACTIVE...", sslCertificateName, certDetails.Type)
	_, errCertStatus := retry.DoWithRetryE(t, "Check SSL Certificate Managed Status", 15, 30*time.Second, func() (string, error) {
		status, errStatus := getSslCertificateManagedStatus(t, projectID, sslCertificateName)
		if errStatus != nil {
			t.Logf("Error getting SSL certificate managed status: %v", errStatus)
			return "", errStatus
		}
		t.Logf("Current managed status of %s: %s", sslCertificateName, status)
		if status == "PROVISIONING" || status == "ACTIVE" {
			return "Certificate is PROVISIONING or ACTIVE", nil
		}
		failedStates := []string{"FAILED_NOT_VISIBLE", "FAILED_CAA_CHECKING", "FAILED_CAA_FORBIDDEN", "FAILED_RATE_LIMITED"}
		for _, failedState := range failedStates {
			if status == failedState {
				return "", retry.FatalError{Underlying: fmt.Errorf("certificate entered a FAILED state: %s", status)}
			}
		}
		return "", fmt.Errorf("certificate %s is still %s, not yet PROVISIONING or ACTIVE", sslCertificateName, status)
	})
	assert.NoError(t, errCertStatus, "Assertion Failed: SSL certificate did not reach PROVISIONING or ACTIVE state in time or entered a failed state.")

	t.Logf("Google-Managed SSL certificate %s successfully created and its properties are validated.", sslCertificateName)
}

// getSslCertificateDetails fetches all relevant details of an SSL certificate using gcloud describe
// and maps them to the CertificateDetails struct.
func getSslCertificateDetails(t *testing.T, projectID, certName string) (*CertificateDetails, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "ssl-certificates", "describe", certName, "--project=" + projectID, "--global", "--format=json", "--verbosity=error"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		return nil, fmt.Errorf("error describing SSL certificate %s: %w", certName, err)
	}

	result := gjson.Parse(output)
	if !result.Exists() {
		return nil, fmt.Errorf("gjson failed to parse output or output is empty for cert %s. Output: %s", certName, output)
	}

	details := &CertificateDetails{
		Name:              result.Get("name").String(),
		Type:              result.Get("type").String(),
		SelfLink:          result.Get("selfLink").String(),
		ManagedStatus:     result.Get("managed.status").String(),
		CreationTimestamp: result.Get("creationTimestamp").String(),
	}

	result.Get("managed.domains").ForEach(func(_, value gjson.Result) bool {
		details.ManagedDomains = append(details.ManagedDomains, value.String())
		return true // continue iteration
	})

	return details, nil
}

// getSslCertificateManagedStatus specifically fetches the 'managed.status' for retry logic.
func getSslCertificateManagedStatus(t *testing.T, projectID, certName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "ssl-certificates", "describe", certName, "--project=" + projectID, "--global", "--format=json", "--verbosity=error"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Warning: 'gcloud compute ssl-certificates describe %s' failed. Error: %v. Output: %s", certName, err, output)
		return "UNKNOWN_ERROR_DURING_DESCRIBE", fmt.Errorf("gcloud describe failed: %w", err)
	}

	result := gjson.Parse(output)
	if !result.Exists() {
		t.Logf("Warning: gjson failed to parse output or output is empty for status check of cert %s. Output: %s", certName, output)
		return "INVALID_JSON_OUTPUT", fmt.Errorf("gjson failed to parse output for status check of cert %s. Full output: %s", certName, output)
	}

	status := result.Get("managed.status").String()
	if status == "" {
		t.Logf("Warning: 'managed.status' field not found or empty for certificate %s from output: %s.", certName, output)
		return "EMPTY_STATUS_FIELD_OR_PARSE_ERROR", fmt.Errorf("'managed.status' field is empty for cert %s. Full JSON: %s", certName, output)
	}
	return status, nil
}
