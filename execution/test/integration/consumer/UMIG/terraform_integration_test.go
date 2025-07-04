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

package integrationtest

import (
	"fmt"
	"math/rand"
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
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/UMIG") // Updated for UMIG
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/UMIG/config")
)

var (
	projectID     = os.Getenv("TF_VAR_project_id")
	umigName      = fmt.Sprintf("umig-%d", rand.Intn(10000)) // Renamed for UMIG
	region        = "us-central1"
	zone          = "us-central1-a"
	vpcName       = "testing-net-umig"
	subnetName    = "testing-subnet-umig"
	instanceNames = []string{"umig-instance-1", "umig-instance-2"}
)

// NamedPortConfig struct to match the YAML structure
type NamedPortConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

var (
	yaml_file_name = "instance.yaml"
)

// UMIGConfig struct to match the new YAML structure
type UMIGConfig struct {
	Name        string            `yaml:"name"`
	ProjectID   string            `yaml:"project_id"`
	Zone        string            `yaml:"zone"`
	Network     string            `yaml:"network"`
	Description string            `yaml:"description"`
	Instances   []string          `yaml:"instances"`
	NamedPorts  []NamedPortConfig `yaml:"named_ports"`
}

/*
TestUMIGs tests the creation and configuration of Unmanaged Instance Groups (UMIGs) in Google Cloud Platform
using Terraform. It verifies that the resources are correctly provisioned, checks the status of VM instances,
and ensures that configurations such as instance group names, zones, and named ports match expected values.
*/
func TestUMIGs(t *testing.T) {
	// Ensure config folder exists
	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory at %s: %v", configFolderPath, err)
	}

	createConfigYAML(t) // Create the UMIG configuration YAML

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

	// --- Pre-Terraform resource creation ---
	createVPC(t, projectID, vpcName)
	time.Sleep(10 * time.Second) // Give VPC time to provision
	createSubnet(t, projectID, vpcName, subnetName, region, "10.0.0.0/24")
	time.Sleep(10 * time.Second) // Give subnet time to provision
	time.Sleep(10 * time.Second) // Give firewall rule time to provision
	createVMInstances(t, projectID, zone, vpcName, subnetName, instanceNames)
	time.Sleep(30 * time.Second) // Give VM instances time to provision and become ready

	// --- Deferred cleanup ---
	defer deleteVPC(t, projectID, vpcName)
	defer deleteSubnet(t, projectID, subnetName, region)
	defer deleteVMInstances(t, projectID, zone, instanceNames)
	defer terraform.Destroy(t, terraformOptions)
	defer deleteUMIGConfigYAML(t)

	// Apply Terraform
	terraform.InitAndApply(t, terraformOptions)

	// Retrieve outputs
	umigSelfLinksOutput := terraform.OutputJson(t, terraformOptions, "umig_self_links")
	umigInstancesOutput := terraform.OutputJson(t, terraformOptions, "umig_instances")

	t.Logf("UMIG Self Links Output: %s", umigSelfLinksOutput)
	t.Logf("UMIG Instances Output: %s", umigInstancesOutput)

	// Parse outputs
	umigSelfLinks := gjson.Parse(umigSelfLinksOutput).Map()
	umigInstances := gjson.Parse(umigInstancesOutput).Map()

	// Check if UMIG self links are present
	var umigModuleKey string
	for k := range umigSelfLinks {
		umigModuleKey = k
		break
	}
	if umigModuleKey == "" {
		t.Fatalf("Could not find key in umig_self_links output. Output was: %s", umigSelfLinksOutput)
	}
	t.Logf("Found UMIG module key: %s", umigModuleKey)

	// Verify the created UMIG
	yamlFilePath := filepath.Join(configFolderPath, yaml_file_name)
	yamlFile, err := os.ReadFile(yamlFilePath)
	if err != nil {
		t.Fatalf("Error reading YAML file at %s: %s", yamlFilePath, err)
	}

	var expectedUMIG UMIGConfig
	err = yaml.Unmarshal(yamlFile, &expectedUMIG)
	if err != nil {
		t.Fatalf("Error unmarshaling YAML from %s: %s", yamlFilePath, err)
	}

	t.Logf("Verifying UMIG: %s", expectedUMIG.Name)

	maxRetries := 5
	retryInterval := 15 * time.Second
	var actualUMIGInfo gjson.Result
	var lastGcloudOutput string
	succeeded := false

	for i := 0; i < maxRetries; i++ {
		gcloudRawOutput := shell.RunCommandAndGetOutput(t, shell.Command{
			Command: "gcloud",
			Args:    []string{"compute", "instance-groups", "unmanaged", "describe", expectedUMIG.Name, "--zone", expectedUMIG.Zone, "--project=" + projectID, "--format=json"},
		})
		lastGcloudOutput = gcloudRawOutput // Store the last output for debugging

		jsonStartIndex := strings.Index(gcloudRawOutput, "{")
		gcloudJSONOutput := ""
		if jsonStartIndex != -1 {
			gcloudJSONOutput = gcloudRawOutput[jsonStartIndex:]
		}

		actualUMIGInfo = gjson.Parse(gcloudJSONOutput)

		// Check if the required fields are present in the gcloud output
		if actualUMIGInfo.Get("name").Exists() && actualUMIGInfo.Get("network").Exists() && actualUMIGInfo.Get("zone").Exists() && actualUMIGInfo.Get("namedPorts").Exists() {
			t.Logf("gcloud describe returned complete data after %d retry/retries.", i+1)
			succeeded = true
			break
		}
		t.Logf("gcloud describe output for %s not yet complete. Retrying in %v...", expectedUMIG.Name, retryInterval)
		time.Sleep(retryInterval)
	}

	if !succeeded {
		t.Fatalf("Failed to get complete UMIG info for '%s' after %d retries. Last gcloud output:\n%s", expectedUMIG.Name, maxRetries, lastGcloudOutput)
	}

	t.Log("========= Verify Unmanaged Instance Group =========")

	// Verify Name
	actualName := actualUMIGInfo.Get("name").String()
	assert.Equal(t, expectedUMIG.Name, actualName, "UMIG name mismatch")
	t.Logf("Confirmed UMIG name matches: %s", actualName)

	// Verify Zone
	actualZoneURL := actualUMIGInfo.Get("zone").String()
	actualZoneName := filepath.Base(actualZoneURL)
	assert.Equal(t, expectedUMIG.Zone, actualZoneName, "UMIG zone mismatch")
	t.Logf("Confirmed UMIG zone matches: %s", actualZoneName)

	// Verify Network
	actualNetworkURL := actualUMIGInfo.Get("network").String()
	expectedNetworkURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", projectID, expectedUMIG.Network)
	assert.Equal(t, expectedNetworkURL, actualNetworkURL, "UMIG network mismatch")
	t.Logf("Confirmed UMIG network matches: %s", actualNetworkURL)

	// Verify Named Ports
	actualNamedPorts := actualUMIGInfo.Get("namedPorts").Array()
	assert.Len(t, actualNamedPorts, len(expectedUMIG.NamedPorts), "Number of named ports mismatch")
	for _, expectedPort := range expectedUMIG.NamedPorts {
		found := false
		for _, actualPort := range actualNamedPorts {
			if actualPort.Get("name").String() == expectedPort.Name && int(actualPort.Get("port").Int()) == expectedPort.Port {
				found = true
				break
			}
		}
		assert.True(t, found, "Named port %s:%d not found in actual UMIG", expectedPort.Name, expectedPort.Port)
	}
	t.Log("Confirmed named ports match.")

	// Verify Instances (from Terraform output)
	actualUMIGInstances := umigInstances[umigModuleKey].Array()
	assert.Len(t, actualUMIGInstances, len(expectedUMIG.Instances), "Number of instances in UMIG output mismatch")
	for _, expectedInstanceName := range expectedUMIG.Instances {
		found := false
		for _, actualInstanceSelfLink := range actualUMIGInstances {
			if strings.HasSuffix(actualInstanceSelfLink.String(), "/"+expectedInstanceName) {
				found = true
				break
			}
		}
		assert.True(t, found, "Instance %s not found in UMIG's instances list", expectedInstanceName)
	}
	t.Log("Confirmed instances in UMIG output match.")
}

// createConfigYAML creates the configuration YAML file for a UMIG instance.
func createConfigYAML(t *testing.T) {
	t.Log("========= Creating UMIG YAML File =========")

	umigInstance := UMIGConfig{
		Name:        umigName,
		ProjectID:   projectID,
		Zone:        zone,
		Network:     vpcName,
		Description: "Integration test unmanaged instance group.",
		Instances:   instanceNames, // Use the pre-defined instance names
		NamedPorts: []NamedPortConfig{
			{Name: "http", Port: 80},
			{Name: "https", Port: 443},
		},
	}

	yamlData, err := yaml.Marshal(&umigInstance)
	if err != nil {
		t.Errorf("Error while marshaling YAML: %v", err)
	}

	// Construct file path
	filePath := filepath.Join(configFolderPath, yaml_file_name)

	t.Logf("Created YAML config at %s with content:\n%s", filePath, string(yamlData))

	err = os.WriteFile(filePath, []byte(yamlData), 0644) // Use 0644 for file permissions
	if err != nil {
		t.Fatalf("Unable to write data into the file: %v", err)
	}
}

// deleteUMIGConfigYAML deletes the generated UMIG configuration YAML file.
func deleteUMIGConfigYAML(t *testing.T) {
	filePath := filepath.Join(configFolderPath, yaml_file_name)
	if err := os.Remove(filePath); err != nil {
		t.Logf("Warning: Failed to delete UMIG config YAML file %s: %v", filePath, err)
	} else {
		t.Logf("Successfully deleted UMIG config YAML file: %s", filePath)
	}
}

/*
createVPC creates the VPC before the test execution.
*/
func createVPC(t *testing.T, projectID string, vpcName string) {
	t.Logf("Creating VPC '%s' in project '%s'...", vpcName, projectID)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "create", vpcName, "--project=" + projectID, "--format=json", "--subnet-mode=custom"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating VPC '%s': %s", vpcName, err)
	}
	t.Logf("Successfully created VPC '%s'.", vpcName)
}

/*
createSubnet creates a subnet within the specified VPC.
*/
func createSubnet(t *testing.T, projectID, vpcName, subnetName, region, ipRange string) {
	t.Logf("Creating subnet '%s' in VPC '%s', region '%s'...", subnetName, vpcName, region)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--region=" + region,
			"--range=" + ipRange,
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating subnet '%s': %s", subnetName, err)
	}
	t.Logf("Successfully created subnet '%s'.", subnetName)
}

/*
createVMInstances creates a set of compute instances to be used with the UMIG.
*/
func createVMInstances(t *testing.T, projectID, zone, network, subnet string, instances []string) {
	t.Logf("Creating VM instances in project '%s', zone '%s'...", projectID, zone)
	for _, instName := range instances {
		t.Logf("Creating instance: %s", instName)
		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "instances", "create", instName,
				"--project=" + projectID,
				"--zone=" + zone,
				"--machine-type=e2-micro", // Or any suitable machine type
				"--network=" + network,
				"--subnet=" + subnet,
				"--image-family=ubuntu-2204-lts",
				"--image-project=ubuntu-os-cloud",
				"--boot-disk-size=10GB",
				"--no-address",
				"--format=json",
			},
		}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Fatalf("Error creating VM instance '%s': %s", instName, err)
		}
		t.Logf("VM instance '%s' created.", instName)
	}
	t.Logf("All VM instances created.")
}

/*
deleteVMInstances deletes the specified compute instances.
*/
func deleteVMInstances(t *testing.T, projectID, zone string, instances []string) {
	t.Logf("Deleting VM instances in project '%s', zone '%s'...", projectID, zone)
	for _, instName := range instances {
		t.Logf("Deleting instance: %s", instName)
		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "instances", "delete", instName,
				"--project=" + projectID,
				"--zone=" + zone,
				"--quiet",
			},
		}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Logf("Error deleting VM instance '%s': %s (may already be deleted)", instName, err)
		} else {
			t.Logf("VM instance '%s' deleted.", instName)
		}
	}
	t.Logf("All VM instances deleted.")
}

/*
deleteSubnet deletes the specified subnet.
*/
func deleteSubnet(t *testing.T, projectID, subnetName, region string) {
	t.Logf("Deleting subnet '%s' in region '%s'...", subnetName, region)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "networks", "subnets", "delete", subnetName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Logf("Error deleting subnet '%s': %s (may already be deleted)", subnetName, err)
	} else {
		t.Logf("Successfully deleted subnet '%s'.", subnetName)
	}
}

/*
deleteVPC deletes the VPC after the test.
*/
func deleteVPC(t *testing.T, projectID string, vpcName string) {
	// Give some time for resources to detach from VPC before deletion
	time.Sleep(30 * time.Second)
	t.Logf("Deleting VPC '%s'...", vpcName)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "delete", vpcName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Logf("Error deleting VPC '%s': %s (may already be deleted or still have dependencies)", vpcName, err)
	} else {
		t.Logf("Successfully deleted VPC '%s'.", vpcName)
	}
}
