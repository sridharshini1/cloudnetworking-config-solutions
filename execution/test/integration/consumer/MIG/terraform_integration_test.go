/**
 * Copyright 2024 Google LLC
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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// Test configuration (adjust as needed)
var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/MIG")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/MIG/config")
)

var (
	projectID        = os.Getenv("TF_VAR_project_id")
	migName          = fmt.Sprintf("mig-%d", rand.Int())
	region           = "us-central1"
	zone             = "us-central1-a"
	vpcName          = "testing-net-mig"
	subnetName       = "testing-subnet-mig"
	firewallRuleName = "fw-allow-health-check"
)

// AutoscalerConfig struct to match the YAML structure
type AutoscalerConfig struct {
	MaxReplicas int `yaml:"max_replicas"`
	MinReplicas int `yaml:"min_replicas"`
}

var (
	yaml_file_name = "instance.yaml"
)

// MIGConfig struct to match the new YAML structure
type MIGConfig struct {
	Name             string           `yaml:"name"`
	ProjectID        string           `yaml:"project_id"`
	Location         string           `yaml:"location"`
	Zone             string           `yaml:"zone"`
	VPCName          string           `yaml:"vpc_name"`
	SubnetworkName   string           `yaml:"subnetwork_name"`
	TargetSize       int              `yaml:"target_size"`
	AutoscalerConfig AutoscalerConfig `yaml:"autoscaler_config"`
}

/*
TestMIGs tests the creation and configuration of Managed Instance Groups (MIGs) in Google Cloud Platform
using Terraform. It verifies that the resources are correctly provisioned, checks the status of VM instances,
and ensures that configurations such as instance group names, zones, and autoscaler settings match expected values.
*/
func TestMIGs(t *testing.T) {
	createConfigYAML(t) // Use the updated createConfigYAML for MIG

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

	createVPC(t, projectID, vpcName) // Create VPC before applying Terraform
	time.Sleep(30 * time.Second)
	createFirewallRule(t, projectID, vpcName) // Create Firewall rule before applying Terraform
	time.Sleep(30 * time.Second)

	defer deleteVPC(t, projectID, vpcName)                   // Delete VPC after test
	defer deleteFirewallRule(t, projectID, firewallRuleName) // Delete Firewall rule after test
	defer terraform.Destroy(t, terraformOptions)             // Destroy resources after test

	// Apply Terraform
	terraform.InitAndApply(t, terraformOptions)

	// Retrieve outputs
	autoscalerOutput := terraform.OutputJson(t, terraformOptions, "autoscaler")
	groupManagerOutput := terraform.OutputJson(t, terraformOptions, "group_manager")
	vmInstances := gjson.Parse(groupManagerOutput).Map()
	t.Logf(autoscalerOutput, groupManagerOutput, vmInstances)

	maxRetries := 5
	retryInterval := 20 * time.Second

	for _, instanceDetails := range vmInstances {
		instanceName := instanceDetails.Get("name").String()
		t.Logf("Checking instance details for '%s' in 60 seconds", instanceName)
		time.Sleep(6 * time.Second)

		for i := 0; i < maxRetries; i++ {
			statusOutput := shell.RunCommandAndGetOutput(t, shell.Command{
				Command: "gcloud",
				Args:    []string{"compute", "instance-groups", "managed", "list-instances", instanceName, "--region", region, "--project=" + projectID, "--format=json"},
			})

			t.Logf("Status Output: %s", statusOutput)
			// Parse the JSON output
			statusJSON := gjson.Parse(statusOutput)

			// Access elements using array indexing where needed
			status := statusJSON.Get("0.instanceStatus").String() // Accessing first element of array for instanceStatus
			instanceURL := gjson.Get(statusOutput, "0.instance").String()

			if status == "RUNNING" {
				t.Logf("Instance group '%s' is stable (Status: '%s'). Proceeding with assertions...", instanceName, status)

				gcloudOutput := shell.RunCommandAndGetOutput(t, shell.Command{
					Command: "gcloud",
					Args:    []string{"compute", "instance-groups", "managed", "describe", instanceName, "--region", region, "--project=" + projectID, "--format=json"},
				})

				yamlFilePath := filepath.Join(configFolderPath, yaml_file_name)
				yamlFile, err := os.ReadFile(yamlFilePath)
				if err != nil {
					t.Fatalf("Error reading YAML file at %s: %s", yamlFilePath, err)
				} else {
					t.Logf("Read YAML file correctly.")
				}

				var expectedInstance MIGConfig
				err = yaml.Unmarshal(yamlFile, &expectedInstance)
				if err != nil {
					t.Fatalf("Error unmarshaling YAML from %s: %s", yamlFilePath, err)
				} else {
					t.Logf("Read MIG Config correctly.")
				}

				t.Log("========= Verify Instance Group =========")
				actualInstanceInfo := gjson.Parse(gcloudOutput)

				actualName := actualInstanceInfo.Get("name").String()
				t.Logf("Actual Instance Group Name: %s, Expected: %s", actualName, expectedInstance.Name)
				if actualName != expectedInstance.Name {
					t.Errorf("Instance group name mismatch: actual=%s, expected=%s", actualName, expectedInstance.Name)
				} else {
					t.Logf("Confirmed instance group name matches: %s", expectedInstance.Name)
				}

				actualZone := extractZoneFromInstanceURL(instanceURL, t)
				distributionPolicyZones := gjson.Get(groupManagerOutput, fmt.Sprintf("%s.distribution_policy_zones", instanceName)).Array()
				zoneStrings := make([]string, len(distributionPolicyZones))
				for i, z := range distributionPolicyZones {
					zoneStrings[i] = z.String()
				}

				if assert.Contains(t, zoneStrings, actualZone, "Actual zone not found in distribution policy zones") {
					t.Logf("Zone assertion successful: Actual zone '%s' found in distribution policy zones.", actualZone)
				}

				// Verify Autoscaler Configuration
				autoscalerConfig := gjson.Parse(autoscalerOutput).Map()
				for migName := range vmInstances {
					t.Logf("Verifying autoscaler for MIG key: %s", migName)

					maxReplicas := autoscalerConfig[migName].Get("autoscaling_policy.0.max_replicas").Int()
					minReplicas := autoscalerConfig[migName].Get("autoscaling_policy.0.min_replicas").Int()

					if assert.Greater(t, maxReplicas, int64(0), "Max replicas should be greater than 0 for %s", migName) {
						t.Logf("Max replicas assertion successful for %s: %d > 0", migName, maxReplicas)
					}

					if assert.GreaterOrEqual(t, minReplicas, int64(1), "Min replicas should be at least 1 for %s", migName) {
						t.Logf("Min replicas assertion successful for %s: %d >= 1", migName, minReplicas)
					}
				}

				break // Exit retry loop if successful
			}

			t.Logf("Instance group '%s' not yet stable (Status: '%s'). Retrying...", instanceName, status)
			time.Sleep(retryInterval)
		}
	}

}

// createConfigYAML creates the configuration YAML file for a MIG instance.
func createConfigYAML(t *testing.T) {
	t.Log("========= YAML File =========")

	migInstance := MIGConfig{
		Name:           migName,
		ProjectID:      projectID,
		Location:       region,
		Zone:           zone,
		VPCName:        vpcName,
		SubnetworkName: subnetName,
		TargetSize:     1, // setting it to minimal for testing
		AutoscalerConfig: AutoscalerConfig{
			MaxReplicas: 3,
			MinReplicas: 1,
		},
	}

	yamlData, err := yaml.Marshal(&migInstance)
	if err != nil {
		t.Errorf("Error while marshaling: %v", err)
	}

	configDir := "config" // Specify a directory for config files (adjust if needed)

	// Construct file path
	filePath := filepath.Join(configDir, yaml_file_name)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	t.Logf("Created YAML config at %s with content:\n%s", filePath, string(yamlData))

	err = os.WriteFile(filePath, []byte(yamlData), 0644) // Use 0644 for file permissions
	if err != nil {
		t.Fatalf("Unable to write data into the file: %v", err)
	}
}

/*
extractZoneFromInstanceURL extracts the zone from a given instance URL in Google Cloud Platform.
It parses the URL, splits the path, and searches for the "zones" segment to retrieve the corresponding zone value.
If the zone cannot be extracted, it logs an error and fails the test.
*/
func extractZoneFromInstanceURL(instanceURL string, t *testing.T) string {
	u, err := url.Parse(instanceURL)
	if err != nil {
		return "" // Or handle the error more appropriately
	}

	parts := strings.Split(u.Path, "/")

	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "zones" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if zone == "" {
		t.Logf("Could not extract zone from instance URL: %s, Path: %s", instanceURL, u.Path)
		assert.FailNow(t, "Could not extract zone. Check instance URL.")
	}
	return zone
}

/*
createVPC creates the VPC and subnet before the test execution and enables Cloud NAT.
*/
func createVPC(t *testing.T, projectID string, vpcName string) {
	text := "compute"

	// Create VPC
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", vpcName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
		return // Exit if VPC creation fails
	}

	// Create Subnet
	cmd = shell.Command{ // Re-use cmd variable
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--region=" + region,
			"--range=10.0.0.0/24",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
		return // Exit if subnet creation fails
	}

	// Enable Cloud NAT for the created VPC

	routerName := vpcName + "-router"
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "routers", "create", routerName,
			"--project=" + projectID,
			"--region=" + region,
			"--network=" + vpcName,
			"--asn=65001",
			"--format=json",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while creating router for NAT: %s", err, routerName)
		return // Exit if router creation fails
	}

	natName := vpcName + "-nat" // Name for the NAT gateway
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "routers", "nats", "create", natName,
			"--router=" + routerName,
			"--project=" + projectID,
			"--region=" + region,
			"--nat-all-subnet-ip-ranges",
			"--auto-allocate-nat-external-ips",
			"--format=json",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while creating router for NAT: %s", err, natName)
		return // Exit if router creation fails
	}

	t.Logf("Successfully created VPC '%s' with subnet '%s' and enabled Cloud NAT.", vpcName, subnetName)
}

/*
createFirewallRule creates a firewall rule for the specified VPC network.
*/
func createFirewallRule(t *testing.T, projectID string, vpcName string) {
	text := "compute"

	// Create Firewall Rule
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "firewall-rules", "create", firewallRuleName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--allow=tcp:80",
			"--source-ranges=130.211.0.0/22,35.191.0.0/16",
			"--description=Allow health checks",
			"--priority=1000",
			"--enable-logging",
			"--quiet",
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while creating firewall rule: %s", err, firewallRuleName)
		return // Exit if firewall rule creation fails
	}

	t.Logf("Successfully created firewall rule '%s' for VPC '%s'.", firewallRuleName, vpcName)
}

/*
deleteFirewallRule deletes the specified firewall rule from the VPC network.
*/
func deleteFirewallRule(t *testing.T, projectID string, firewallRuleName string) {
	text := "compute"

	// Delete Firewall Rule
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "firewall-rules", "delete", firewallRuleName,
			"--project=" + projectID,
			"--quiet",
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting firewall rule: %s", err, firewallRuleName)
		return // Exit if firewall rule deletion fails
	}

	t.Logf("Successfully deleted firewall rule '%s'.", firewallRuleName)
}

/*
deleteVPC deletes the VPC, subnet, and Cloud NAT configuration after the test.
*/
func deleteVPC(t *testing.T, projectID string, vpcName string) {
	text := "compute"
	time.Sleep(60 * time.Second) // Wait for resources to be in a deletable state

	// Delete Cloud NAT mappings
	natName := vpcName + "-nat" // Assuming NAT was created with this naming convention.
	routerName := vpcName + "-router"

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "routers", "nats", "delete", natName,
			"--project=" + projectID,
			"--region=" + region,
			"--router=" + routerName,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting NAT router: %s", err, natName)
	}

	// Delete the router itself after deleting the NAT
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "routers", "delete", routerName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting NAT router: %s", err, routerName)
	}

	// Delete Subnet
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "delete", subnetName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting subnet: %s", err, subnetName)
	}

	time.Sleep(150 * time.Second) // Wait for firewall deletion to complete

	// Delete VPC
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", vpcName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting VPC: %s", err, vpcName)
	}
}
