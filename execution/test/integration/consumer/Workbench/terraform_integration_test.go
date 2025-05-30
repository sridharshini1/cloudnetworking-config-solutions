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
package integrationtest

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

// Test configuration (adjust as needed)
var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Workbench")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Workbench/config")
)

var (
	projectID            = os.Getenv("TF_VAR_project_id") // Ensure TF_VAR_project_id is set in your environment
	region               = "us-central1"
	zone                 = "us-central1-a"
	vpcName              = fmt.Sprintf("testing-net-workbench-%d", rand.Intn(100000000))
	subnetName           = fmt.Sprintf("testing-subnet-workbench-%d", rand.Intn(100000000))
	workbenchName        = fmt.Sprintf("workbench-%d", rand.Intn(100000000))
	yaml_file_name       = "instance.yaml"
	connectivityTestName = fmt.Sprintf("workbench-bq-test-%d", rand.Intn(100000000))
)

// WorkbenchConfig struct to match the YAML structure
type WorkbenchConfig struct {
	Name      string `yaml:"name"`
	ProjectID string `yaml:"project_id"`
	Location  string `yaml:"location"`
	GceSetup  struct {
		NetworkInterfaces []struct {
			Network string `yaml:"network"`
			Subnet  string `yaml:"subnet"`
		} `yaml:"network_interfaces"`
	} `yaml:"gce_setup"`
}

// TestWorkbenchWithBigQueryConnectivity verifies the end-to-end connectivity between a Google Cloud Workbench instance and a BigQuery API within a specified VPC.
// The test provisions required infrastructure using Terraform, asserts the correct creation and configuration of the Workbench instance (including network and proxy settings),
// checks that the instance does not have a public IP, retrieves its internal IP, and finally tests connectivity to BigQuery.
// Resources are cleaned up after the test completes.
func TestWorkbenchWithBigQueryConnectivity(t *testing.T) {
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable is not set.")
	}
	createConfigYAML(t, filepath.Join(configFolderPath, yaml_file_name))

	tfVars := map[string]interface{}{
		"config_folder_path": configFolderPath,
	}

	// Terraform Options
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars:         tfVars,
		TerraformDir: terraformDirectoryPath,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	createVPC(t, projectID, vpcName)
	time.Sleep(30 * time.Second) // Allow time for VPC/subnet to be fully ready
	time.Sleep(30 * time.Second) // Allow time for firewall rule to propagate

	// Defer cleanup operations
	defer deleteVPC(t, projectID, vpcName)       // Then delete VPC
	defer terraform.Destroy(t, terraformOptions) // Finally, terraform destroy

	// Apply Terraform
	terraform.InitAndApply(t, terraformOptions)

	allWorkbenchOutputs := terraform.OutputJson(t, terraformOptions, "")
	workbenchInstanceIds := gjson.Get(allWorkbenchOutputs, "workbench_instance_ids.value").Map()
	workbenchProxyUris := gjson.Get(allWorkbenchOutputs, "workbench_instance_proxy_uris.value").Map()

	workbenchInstanceID := workbenchInstanceIds[workbenchName].String()
	workbenchSSHInstanceName := filepath.Base(workbenchInstanceID) // This is the short name

	// --- Assertions for Workbench instance properties ---
	assert.Contains(t, workbenchInstanceIds, workbenchName, "Workbench instance ID map should contain the instance name key")
	assert.NotEmpty(t, workbenchInstanceID, "Workbench instance ID should not be empty")
	t.Logf("Asserted: Workbench instance name '%s' is present in Terraform outputs.", workbenchName)

	expectedLocationSubstr := fmt.Sprintf("/locations/%s/instances/%s", zone, workbenchSSHInstanceName)
	assert.Contains(t, workbenchInstanceID, expectedLocationSubstr, "Workbench instance ID should contain expected zone")
	t.Logf("Asserted: Workbench instance location (zone) is '%s'.", zone)

	proxyURIResult := workbenchProxyUris[workbenchName]
	assert.False(t, proxyURIResult.Exists() && proxyURIResult.String() != "",
		fmt.Sprintf("Workbench proxy URI was expected to be empty/null but was '%s'", proxyURIResult.String()))
	t.Logf("Asserted: Workbench proxy URI is empty/null as expected: '%s'", proxyURIResult.String())

	t.Logf("Checking for public IP on Workbench instance: %s", workbenchSSHInstanceName)
	workbenchDescribeCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "describe", workbenchSSHInstanceName,
			"--zone=" + zone,
			"--project=" + projectID,
			"--format=json",
		},
	}
	describeOutput, err := shell.RunCommandAndGetOutputE(t, workbenchDescribeCmd)
	if err != nil {
		t.Errorf("Failed to describe Workbench instance '%s' to check for public IP: %v. Output: %s", workbenchSSHInstanceName, err, describeOutput)
	}
	accessConfigs := gjson.Get(describeOutput, "networkInterfaces.0.accessConfigs").Array()
	assert.Empty(t, accessConfigs, "Workbench instance should not have any public IP access configurations")
	t.Logf("Asserted: Workbench instance has no public IP.")

	// Get internal IP
	workbenchInternalIP, err := getInternalIP(t, workbenchSSHInstanceName, projectID, zone) // Using the refined function
	if err != nil {
		t.Errorf("Failed to get internal IP for Workbench instance: %v", err)
	}
	assert.NotEmpty(t, workbenchInternalIP, "Internal IP must be retrievable for subsequent tests")
	t.Logf("Retrieved Workbench instance internal IP: %s", workbenchInternalIP)

	testConnectivity(t, projectID, workbenchInternalIP, vpcName)
}

// createConfigYAML creates the configuration YAML file for a Workbench instance.
// filePath is the absolute path where the YAML file should be written.
func createConfigYAML(t *testing.T, filePath string) {
	t.Log("========= YAML File Creation =========")

	workbenchInstance := WorkbenchConfig{
		Name:      workbenchName,
		ProjectID: projectID,
		Location:  zone, // Use zone for Workbench instance location
		GceSetup: struct {
			NetworkInterfaces []struct {
				Network string `yaml:"network"`
				Subnet  string `yaml:"subnet"`
			} `yaml:"network_interfaces"`
		}{
			NetworkInterfaces: []struct {
				Network string `yaml:"network"`
				Subnet  string `yaml:"subnet"`
			}{
				{
					Network: fmt.Sprintf("projects/%s/global/networks/%s", projectID, vpcName),
					Subnet:  fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectID, region, subnetName),
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(&workbenchInstance)
	if err != nil {
		t.Fatalf("Error while marshaling YAML for WorkbenchConfig: %v", err)
	}

	// Ensure the directory for the filePath exists (in case it's nested)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("Failed to create directory for YAML file %s: %v", filePath, err)
	}

	t.Logf("Creating YAML config at %s with content:\n%s", filePath, string(yamlData))
	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		t.Fatalf("Unable to write data into the file %s: %v", filePath, err)
	}
}

// Function to get the internal IP from the instance
func getInternalIP(t *testing.T, instanceName, projectID, zone string) (string, error) {
	describeCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "describe", instanceName,
			"--project", projectID,
			"--zone", zone,
			"--format=value(networkInterfaces[0].networkIP)",
		},
	}
	stdout, stderr, err := shell.RunCommandAndGetStdOutErrE(t, describeCmd)
	if err != nil {
		return "", fmt.Errorf("failed to get internal IP for instance %s: %w, stderr: %s", instanceName, err, stderr)
	}

	internalIP := strings.TrimSpace(stdout)
	if internalIP == "" {
		return "", fmt.Errorf("failed to extract internal IP from stdout, stderr: %s", stderr)
	}
	t.Logf("Internal IP for instance %s: %s", instanceName, internalIP)
	return internalIP, nil
}

func createVPC(t *testing.T, projectID string, vpcName string) {
	t.Logf("Creating VPC: %s and Subnet: %s", vpcName, subnetName)
	cmdCreateNet := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "create", vpcName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmdCreateNet)
	if err != nil {
		t.Fatalf("Error creating VPC network %s: %v", vpcName, err)
	}

	cmdCreateSubnet := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--region=" + region,
			"--range=10.0.0.0/24", // Make sure this range doesn't overlap if running multiple tests in parallel on same project
			"--enable-private-ip-google-access",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmdCreateSubnet)
	if err != nil {
		t.Fatalf("Error creating subnet %s in VPC %s: %v", subnetName, vpcName, err)
	}
	t.Logf("Successfully created VPC '%s' with PGA-enabled subnet '%s'.", vpcName, subnetName)
}

func deleteVPC(t *testing.T, projectID string, vpcName string) {
	t.Logf("Deleting Subnet: %s and VPC: %s", subnetName, vpcName)
	// It can take time for dependent resources (like workbench instance) to be fully deleted
	// before a subnet/network can be. The sleeps are attempts to mitigate this.
	// Ideally, ensure all dependent resources are gone first. Terraform destroy should handle the instance.

	cmdDeleteSubnet := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "subnets", "delete", subnetName, "--project=" + projectID, "--region=" + region, "--quiet"},
	}
	// Retry subnet deletion as it often fails if resources are still attached
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		output, err := shell.RunCommandAndGetOutputE(t, cmdDeleteSubnet)
		if err == nil {
			t.Logf("Successfully deleted subnet %s.", subnetName)
			break
		}
		t.Logf("Error deleting subnet %s (attempt %d/%d): %v. Output: %s", subnetName, i+1, maxRetries, err, output)
		if i < maxRetries-1 {
			time.Sleep(30 * time.Second)
		} else {
			t.Errorf("Failed to delete subnet %s after %d attempts. Last error: %v", subnetName, maxRetries, err)
			// Continue to attempt VPC deletion anyway
		}
	}

	// Wait longer before attempting network deletion
	t.Logf("Waiting before deleting network %s...", vpcName)
	time.Sleep(60 * time.Second)

	cmdDeleteNet := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "delete", vpcName, "--project=" + projectID, "--quiet"},
	}
	for i := 0; i < maxRetries; i++ {
		output, err := shell.RunCommandAndGetOutputE(t, cmdDeleteNet)
		if err == nil {
			t.Logf("Successfully deleted VPC network %s.", vpcName)
			return
		}
		t.Logf("Error deleting VPC network %s (attempt %d/%d): %v. Output: %s", vpcName, i+1, maxRetries, err, output)
		if i < maxRetries-1 {
			time.Sleep(30 * time.Second)
		} else {
			t.Errorf("Failed to delete VPC network %s after %d attempts. Last error: %v", vpcName, maxRetries, err)
		}
	}
}

// testConnectivity tests the connectivity from the Workbench instance to the BigQuery endpoint.
// It creates a connectivity test using the Network Management API, waits for the analysis to complete,
// and asserts that the Workbench instance can reach the BigQuery endpoint.
func testConnectivity(t *testing.T, projectID, sourceIP, networkName string) {
	t.Log("========= Network Management Connectivity Test =========")

	bigqueryEndpoint := "bigquery.googleapis.com"

	resolvedIP, err := resolveHostname(t, bigqueryEndpoint, projectID, networkName)
	if err != nil {
		t.Fatalf("Failed to resolve hostname '%s': %v", bigqueryEndpoint, err)
	}
	t.Logf("Resolved IP for %s: %s", bigqueryEndpoint, resolvedIP)

	createCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"network-management", "connectivity-tests", "create", connectivityTestName,
			"--project=" + projectID,
			"--source-ip-address=" + sourceIP,
			"--destination-ip-address=" + resolvedIP,
			"--protocol=TCP",
			"--destination-port=443",
			"--source-network=projects/" + projectID + "/global/networks/" + networkName,
			"--format=json",
			"--verbosity=debug",
		},
	}

	createOutput, err := shell.RunCommandAndGetOutputE(t, createCmd)
	if err != nil {
		t.Fatalf("Error creating connectivity test '%s': %v. Output: %s", connectivityTestName, err, createOutput)
	}
	t.Logf("Connectivity test '%s' creation initiated. Waiting for analysis...", connectivityTestName)

	defer func() {
		t.Logf("Deleting connectivity test: %s", connectivityTestName)
		deleteCmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"network-management", "connectivity-tests", "delete", connectivityTestName, "--project=" + projectID, "--quiet"},
		}
		delOutput, delErr := shell.RunCommandAndGetOutputE(t, deleteCmd)
		if delErr != nil {
			t.Errorf("Error deleting connectivity test %s: %v. Output: %s", connectivityTestName, delErr, delOutput)
		} else {
			t.Logf("Connectivity test '%s' deleted.", connectivityTestName)
		}
	}()

	maxPollRetries := 3
	pollInterval := 20 * time.Second
	testStatus := ""
	finalDescribeOutput := ""
	var reachabilityResult gjson.Result

	for i := 0; i < maxPollRetries; i++ {
		time.Sleep(pollInterval)
		describeCmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"network-management", "connectivity-tests", "describe", connectivityTestName, "--project=" + projectID, "--format=json"},
		}
		describeOutput, describeErr := shell.RunCommandAndGetOutputE(t, describeCmd)
		finalDescribeOutput = describeOutput
		if describeErr != nil {
			t.Logf("Error describing connectivity test (attempt %d/%d): %v. Output: %s", i+1, maxPollRetries, describeErr, describeOutput)
			continue
		}

		// Extract only the JSON part from the output (skip warnings/logs)
		jsonStart := strings.Index(describeOutput, "{")
		if jsonStart != -1 {
			describeOutput = describeOutput[jsonStart:]
		}
		parsedOutput := gjson.Parse(describeOutput)
		reachabilityResult = parsedOutput.Get("reachabilityDetails.result")
		if reachabilityResult.Exists() {
			testStatus = reachabilityResult.String()
			t.Logf("Connectivity test status: %s (Attempt %d/%d)", testStatus, i+1, maxPollRetries)
			if testStatus == "REACHABLE" || testStatus == "UNREACHABLE" || testStatus == "AMBIGUOUS" {
				break
			}
		} else {
			t.Logf("Connectivity test status not yet available (attempt %d/%d). State: %s. Output: %s", i+1, maxPollRetries, parsedOutput.Get("state").String(), describeOutput)
		}
	}

	// Assert the final testStatus after the loop
	assert.Equal(t, "REACHABLE", testStatus, fmt.Sprintf("Connectivity test to BigQuery endpoint should be REACHABLE. Final describe output:\n%s", finalDescribeOutput))
	if testStatus == "REACHABLE" {
		t.Logf("Connectivity test '%s' PASSED: Workbench can reach BigQuery endpoint.", connectivityTestName)
	} else {
		t.Errorf("Connectivity test '%s' FAILED. Final status: %s. Review test details in GCP console. Final describe output:\n%s", connectivityTestName, testStatus, finalDescribeOutput)
	}
}

// resolveHostname resolves a given hostname to its IP address.
// Function to resolve a hostname to an IP address
func resolveHostname(t *testing.T, hostname, projectID, networkName string) (string, error) {
	// 1. Try Go's net.LookupIP first
	t.Logf("Attempting to resolve hostname '%s' using net.LookupIP...", hostname)
	ips, err := net.LookupIP(hostname)
	if err == nil && len(ips) > 0 {
		for _, ip := range ips {
			if ip.To4() != nil { // Prefer IPv4
				t.Logf("Successfully resolved '%s' to %s using net.LookupIP.", hostname, ip.String())
				return ip.String(), nil
			}
		}
	}
	t.Logf("net.LookupIP failed or returned no IPv4: %v. Falling back to external tools.", err)

	workbenchInstanceName := workbenchName
	networkCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "describe", workbenchInstanceName,
			"--project", projectID,
			"--zone", zone,
			"--format=value(networkInterfaces[0].network)",
		},
	}
	networkOutput, networkErr := shell.RunCommandAndGetOutputE(t, networkCmd)
	if networkErr != nil {
		t.Logf("Warning: Failed to get network for instance '%s': %v, Output: %s. Proceeding with hostname resolution.", workbenchInstanceName, networkErr, networkOutput)
	} else if !strings.Contains(networkOutput, networkName) {
		t.Logf("Warning: The provided network name '%s' does not match the network of the instance '%s'. Output: %s. Proceeding with hostname resolution.", networkName, workbenchInstanceName, networkOutput)
	}

	t.Logf("Attempting to resolve hostname '%s' using 'dig'...", hostname)
	digCmd := shell.Command{
		Command: "dig",
		Args: []string{
			"+short", hostname,
		},
	}
	stdout, stderr, err := shell.RunCommandAndGetStdOutErrE(t, digCmd)
	if err == nil {
		resolvedIP := strings.TrimSpace(stdout)
		if resolvedIP != "" {
			t.Logf("Successfully resolved '%s' to %s using 'dig'.", hostname, resolvedIP)
			return resolvedIP, nil
		}
	}
	t.Logf("dig failed: %v, stderr: %s. Falling back to 'nslookup'.", err, stderr)

	t.Logf("Attempting to resolve hostname '%s' using 'nslookup'...", hostname)
	nslookupCmd := shell.Command{
		Command: "nslookup",
		Args: []string{
			hostname,
		},
	}
	stdout, stderr, err = shell.RunCommandAndGetStdOutErrE(t, nslookupCmd)
	if err == nil {
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "Address:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					ip := strings.TrimSpace(parts[1])
					if net.ParseIP(ip) != nil {
						t.Logf("Successfully resolved '%s' to %s using 'nslookup'.", hostname, ip)
						return ip, nil
					}
				}
			}
		}
	}
	t.Logf("nslookup failed: %v, stderr: %s.", err, stderr)

	return "", fmt.Errorf("failed to resolve hostname '%s' using net.LookupIP, dig, or nslookup", hostname)
}
