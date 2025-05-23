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
	terraformSslModulePath = "../../../../../03-security/Compute-SSL-Certs/Google-Managed"
	projectID              = os.Getenv("TF_VAR_project_id")
)

func TestCreateAndAttachSslCertificate(t *testing.T) {
	t.Parallel() // Mark test as parallelizable

	// Ensure projectID is set
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set.")
	}

	// Correctly seed the random number generator for unique IDs in parallel tests
	// Use a new source for each test if necessary, or ensure unique names via other means for parallel terraform applies
	// For simplicity in this example, ensure TF_VAR_ssl_certificate_name is unique if running parallel tests modifying same global cert names
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	uniqueIDC := r.Int()
	sslCertificateName := fmt.Sprintf("test-ssl-cert-%d", uniqueIDC)
	domainName := fmt.Sprintf("test%d.example.com", uniqueIDC)

	// Variables for the SSL Certificate Terraform module
	tfSslVars := map[string]interface{}{
		"project_id":           projectID,
		"ssl_certificate_name": sslCertificateName,
		"ssl_certificate_type": "MANAGED", // Explicitly set to MANAGED
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
	t.Logf("Applying Terraform configuration for SSL certificate: %s", sslCertificateName)
	terraform.InitAndApply(t, terraformSslOptions)

	// Verify SSL certificate creation and get its self_link
	createdSslCertSelfLink := terraform.Output(t, terraformSslOptions, "managed_ssl_certificate_self_link")
	assert.NotEmpty(t, createdSslCertSelfLink, "SSL certificate self_link should not be empty")

	// Wait for the certificate to be in a PROVISIONING or ACTIVE state
	// Google Managed Certificates can take a while.
	// Statuses: MANAGED_CERTIFICATE_STATUS_UNSPECIFIED, PROVISIONING, FAILED_NOT_VISIBLE, FAILED_CAA_CHECKING, FAILED_CAA_FORBIDDEN, FAILED_RATE_LIMITED, ACTIVE
	t.Logf("Waiting for SSL certificate %s to be PROVISIONING or ACTIVE...", sslCertificateName)
	_, errCertStatus := retry.DoWithRetryE(t, "Check SSL Certificate Status", 15, 30*time.Second, func() (string, error) { // Retry for up to 7.5 mins
		status, errStatus := getSslCertificateStatus(t, projectID, sslCertificateName)
		if errStatus != nil {
			return "", fmt.Errorf("error getting SSL certificate status: %w", errStatus)
		}
		t.Logf("Current status of %s: %s", sslCertificateName, status)
		if status == "PROVISIONING" || status == "ACTIVE" {
			return "Certificate is PROVISIONING or ACTIVE", nil
		}
		// Add specific FAILED states that mean we should stop retrying.
		failedStates := []string{"FAILED_NOT_VISIBLE", "FAILED_CAA_CHECKING", "FAILED_CAA_FORBIDDEN", "FAILED_RATE_LIMITED"}
		for _, failedState := range failedStates {
			if status == failedState {
				return "", retry.FatalError{Underlying: fmt.Errorf("certificate entered a FAILED state: %s", status)}
			}
		}
		return "", fmt.Errorf("certificate %s is still %s, not yet PROVISIONING or ACTIVE", sslCertificateName, status)
	})
	assert.NoError(t, errCertStatus, "SSL certificate did not reach PROVISIONING or ACTIVE state in time or entered a failed state.")

	// --- Load Balancer Setup ---
	// Unique names for LB components
	ipAddressName := fmt.Sprintf("test-lb-ip-%d", uniqueIDC)
	healthCheckName := fmt.Sprintf("test-hc-%d", uniqueIDC)
	backendServiceName := fmt.Sprintf("test-bs-%d", uniqueIDC)
	urlMapName := fmt.Sprintf("test-um-%d", uniqueIDC)
	targetProxyName := fmt.Sprintf("test-thp-%d", uniqueIDC)
	forwardingRuleName := fmt.Sprintf("test-fw-rule-%d", uniqueIDC)

	// Defer cleanup of LB components (in reverse order of creation)
	defer deleteGlobalIpAddress(t, projectID, ipAddressName)           // Runs 6th (last among gcloud defers)
	defer deleteHealthCheck(t, projectID, healthCheckName)             // Runs 5th
	defer deleteBackendService(t, projectID, backendServiceName)       // Runs 4th
	defer deleteUrlMap(t, projectID, urlMapName)                       // Runs 3rd
	defer deleteTargetHttpsProxy(t, projectID, targetProxyName)        // Runs 2nd
	defer deleteGlobalForwardingRule(t, projectID, forwardingRuleName) // Runs 1st (first among gcloud defers)

	// Create LB components
	var err error // Declare err to be used by helper functions
	t.Logf("Creating Global IP Address: %s", ipAddressName)
	_, err = createGlobalIpAddress(t, projectID, ipAddressName)
	assert.NoError(t, err, "Failed to create Global IP Address")

	t.Logf("Creating Health Check: %s", healthCheckName)
	_, err = createHealthCheck(t, projectID, healthCheckName)
	assert.NoError(t, err, "Failed to create Health Check")

	t.Logf("Creating Backend Service: %s", backendServiceName)
	_, err = createBackendService(t, projectID, backendServiceName, healthCheckName)
	assert.NoError(t, err, "Failed to create Backend Service")

	t.Logf("Creating URL Map: %s", urlMapName)
	_, err = createUrlMap(t, projectID, urlMapName, backendServiceName)
	assert.NoError(t, err, "Failed to create URL Map")

	t.Logf("Creating Target HTTPS Proxy: %s with SSL cert: %s", targetProxyName, sslCertificateName)
	// gcloud command expects certificate name for --ssl-certificates flag.
	_, err = createTargetHttpsProxy(t, projectID, targetProxyName, urlMapName, sslCertificateName)
	assert.NoError(t, err, "Failed to create Target HTTPS Proxy")

	t.Logf("Creating Global Forwarding Rule: %s", forwardingRuleName)
	_, err = createGlobalForwardingRule(t, projectID, forwardingRuleName, ipAddressName, targetProxyName)
	assert.NoError(t, err, "Failed to create Global Forwarding Rule")

	// Allow some time for resources to be fully provisioned and associated
	t.Log("Waiting 60 seconds for LB components to stabilize...")
	time.Sleep(60 * time.Second)

	// Confirm SSL certificate attachment to Target HTTPS Proxy
	t.Logf("Verifying SSL certificate attachment to Target HTTPS Proxy: %s", targetProxyName)
	attachedCerts, err := getTargetHttpsProxySslCertificates(t, projectID, targetProxyName)
	assert.NoError(t, err, "Failed to get SSL certificates from Target HTTPS Proxy")

	// The output from gcloud for attached certs are self-links
	found := false
	for _, certSelfLink := range attachedCerts {
		if certSelfLink == createdSslCertSelfLink { // Compare with the self_link from Terraform output
			found = true
			break
		}
	}
	assert.True(t, found, fmt.Sprintf("SSL certificate %s (self-link: %s) not found in Target HTTPS Proxy %s. Found: %v", sslCertificateName, createdSslCertSelfLink, targetProxyName, attachedCerts))

	t.Logf("SSL certificate %s successfully created and attached to Target HTTPS Proxy %s.", sslCertificateName, targetProxyName)
}

// --- Helper functions for gcloud commands ---

func getSslCertificateStatus(t *testing.T, projectID, certName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "ssl-certificates", "describe", certName, "--project=" + projectID, "--global", "--format=json"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		// If describe fails, it might be because the cert is not yet fully registered.
		t.Logf("Warning: 'gcloud compute ssl-certificates describe %s' failed. Output: %s, Error: %v", certName, output, err)
		return "UNKNOWN_OR_NOT_FOUND", nil // Return a status that forces retry
	}
	// Check for 'type' field first; if 'SELF_MANAGED', it's 'ACTIVE' if it exists.
	// This test specifically targets 'MANAGED' type from the Terraform module.
	certType := gjson.Get(output, "type").String()
	var status string
	if certType == "MANAGED" {
		status = gjson.Get(output, "managed.status").String()
	} else if certType == "SELF_MANAGED" { // Should not happen with this module config
		status = "ACTIVE" // Self-managed are implicitly active if they exist
	} else {
		// Fallback if type is missing or unexpected, try to get a general status
		status = gjson.Get(output, "status").String()
	}

	if status == "" {
		t.Logf("Warning: Could not determine status for certificate %s from output: %s. Full JSON: %s", certName, status, output)
		return "UNKNOWN_EMPTY_STATUS", nil // Return a status that forces retry
	}
	return status, nil
}

func createGlobalIpAddress(t *testing.T, projectID, ipName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "addresses", "create", ipName, "--project=" + projectID, "--global", "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteGlobalIpAddress(t *testing.T, projectID, ipName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "addresses", "delete", ipName, "--project=" + projectID, "--global", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting Global IP Address %s: %v. Manual cleanup might be required.", ipName, err)
	} else {
		t.Logf("Successfully deleted Global IP Address: %s", ipName)
	}
}

func createHealthCheck(t *testing.T, projectID, hcName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "health-checks", "create", "http", hcName, "--project=" + projectID, "--global", "--port=80", "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteHealthCheck(t *testing.T, projectID, hcName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "health-checks", "delete", hcName, "--project=" + projectID, "--global", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting Health Check %s: %v. Manual cleanup might be required.", hcName, err)
	} else {
		t.Logf("Successfully deleted Health Check: %s", hcName)
	}
}

func createBackendService(t *testing.T, projectID, bsName, hcName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "backend-services", "create", bsName, "--project=" + projectID, "--protocol=HTTP", "--health-checks=" + hcName, "--global", "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteBackendService(t *testing.T, projectID, bsName string) {
	// Retry deletion as backend services can be slow to release from other resources
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("Deleting Backend Service %s", bsName), 5, 30*time.Second, func() (string, error) {
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"compute", "backend-services", "delete", bsName, "--project=" + projectID, "--global", "--quiet"},
		}
		output, errCmd := shell.RunCommandAndGetOutputE(t, cmd) // Renamed err to errCmd to avoid conflict with outer err
		if errCmd != nil {
			if strings.Contains(output, "being used by resource") || strings.Contains(errCmd.Error(), "being used by resource") {
				t.Logf("Backend Service %s is still in use, retrying delete...", bsName)
				return "", fmt.Errorf("backend service %s still in use: %w", bsName, errCmd) // Retryable error
			}
			// For other errors, fail fast (non-retryable)
			// Corrected line:
			return "", retry.FatalError{Underlying: fmt.Errorf("non-retryable error deleting backend service %s. Output: %s. Error: %w", bsName, output, errCmd)}
		}
		return "Backend service deleted", nil
	})

	if err != nil {
		t.Logf("Failed to delete Backend Service %s after retries: %v. Manual cleanup might be required.", bsName, err)
	} else {
		t.Logf("Successfully deleted Backend Service: %s", bsName)
	}
}

func createUrlMap(t *testing.T, projectID, umName, defaultServiceName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "url-maps", "create", umName, "--project=" + projectID, "--default-service=" + defaultServiceName, "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteUrlMap(t *testing.T, projectID, umName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "url-maps", "delete", umName, "--project=" + projectID, "--global", "--quiet"}, // Added --global for consistency
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting URL Map %s: %v. Manual cleanup might be required.", umName, err)
	} else {
		t.Logf("Successfully deleted URL Map: %s", umName)
	}
}

func createTargetHttpsProxy(t *testing.T, projectID, proxyName, umName, certName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "target-https-proxies", "create", proxyName, "--project=" + projectID, "--url-map=" + umName, "--ssl-certificates=" + certName, "--global", "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteTargetHttpsProxy(t *testing.T, projectID, proxyName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "target-https-proxies", "delete", proxyName, "--project=" + projectID, "--global", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting Target HTTPS Proxy %s: %v. Manual cleanup might be required.", proxyName, err)
	} else {
		t.Logf("Successfully deleted Target HTTPS Proxy: %s", proxyName)
	}
}

func createGlobalForwardingRule(t *testing.T, projectID, ruleName, ipAddressName, proxyName string) (string, error) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "forwarding-rules", "create", ruleName, "--project=" + projectID, "--address=" + ipAddressName, "--global", "--target-https-proxy=" + proxyName, "--ports=443", "--quiet"},
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

func deleteGlobalForwardingRule(t *testing.T, projectID, ruleName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "forwarding-rules", "delete", ruleName, "--project=" + projectID, "--global", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting Global Forwarding Rule %s: %v. Manual cleanup might be required.", ruleName, err)
	} else {
		t.Logf("Successfully deleted Global Forwarding Rule: %s", ruleName)
	}
}

func getTargetHttpsProxySslCertificates(t *testing.T, projectID, proxyName string) ([]string, error) {
	var attachedCerts []string
	// Retry fetching proxy details as it might take time for certificate attachment to reflect
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("Get SSL certs for Target HTTPS Proxy %s", proxyName), 3, 20*time.Second, func() (string, error) {
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"compute", "target-https-proxies", "describe", proxyName, "--project=" + projectID, "--global", "--format=json"},
		}
		output, errCmd := shell.RunCommandAndGetOutputE(t, cmd)
		if errCmd != nil {
			return "", fmt.Errorf("failed to describe target HTTPS proxy %s. Output: %s. Error: %w", proxyName, output, errCmd)
		}

		result := gjson.Get(output, "sslCertificates")
		currentCerts := []string{}
		if result.Exists() && result.IsArray() {
			for _, cert := range result.Array() {
				currentCerts = append(currentCerts, cert.String())
			}
		}
		attachedCerts = currentCerts // Update the outer scope variable
		if len(attachedCerts) > 0 {  // Consider success if any certs are attached, detailed check later
			return "Certificates retrieved", nil
		}
		return "", fmt.Errorf("no SSL certificates found attached yet to proxy %s", proxyName)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get SSL certificates from Target HTTPS Proxy %s after retries: %w", proxyName, err)
	}
	return attachedCerts, nil
}
