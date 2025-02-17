// # Copyright 2025 Google LLC
// #
// # Licensed under the Apache License, Version 2.0 (the "License");
// # you may not use this file except in compliance with the License.
// # You may obtain a copy of the License at
// #
// #     http://www.apache.org/licenses/LICENSE-2.0
// #
// # Unless required by applicable law or agreed to in writing, software
// # distributed under the License is distributed on an "AS IS" BASIS,
// # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// # See the License for the specific language governing permissions and
// # limitations under the License.

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
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// Test configuration (adjust as needed)
var (
	projectRoot, _         = filepath.Abs("../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "07-consumer-load-balancing/Application/External")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer-load-balancing/Application/External/config/")
)

var (
	projectID        = os.Getenv("TF_VAR_project_id")
	instanceName     = fmt.Sprintf("lb-%d", rand.Int())
	region           = "us-central1"
	networkName      = fmt.Sprintf("vpc-%s-test", instanceName)
	subnetName       = fmt.Sprintf("%s-subnet", networkName)
	migName          = fmt.Sprintf("mig-%s", instanceName)               // Name for the Managed Instance Group
	templateName     = fmt.Sprintf("%s-instance-template", instanceName) // Name for the Instance Template
	firewallRuleName = fmt.Sprintf("%s-firewall-rule", networkName)
)

const (
	defaultHCLBName = "load-balancer-default-hc"
	customHCLBName  = "load-balancer-custom-hc"
)

// LoadBalancerConfig struct
type LoadBalancerConfig struct {
	Name      string         `yaml:"name"`
	ProjectID string         `yaml:"project_id"`
	Network   string         `yaml:"network"`
	Backends  BackendsConfig `yaml:"backends"`
}

// MIGInstanceConfig struct
type MIGInstanceConfig struct {
	Name       string `yaml:"name"`
	ProjectID  string `yaml:"project_id"`
	Region     string `yaml:"region"`
	Image      string `yaml:"image"`
	Network    string `yaml:"network"`
	Subnetwork string `yaml:"subnetwork"`
}

type BackendsConfig struct {
	Default BackendConfig `yaml:"default"`
}

type BackendConfig struct {
	Protocol    string        `yaml:"protocol"`
	Port        int           `yaml:"port"`
	PortName    string        `yaml:"port_name"`
	TimeoutSec  int           `yaml:"timeout_sec"`
	EnableCdn   bool          `yaml:"enable_cdn"`
	HealthCheck HealthCheck   `yaml:"health_check"`
	LogConfig   LogConfig     `yaml:"log_config"`
	Groups      []GroupConfig `yaml:"groups"`
}

type HealthCheck struct {
	RequestPath string `yaml:"request_path"`
	Port        int    `yaml:"port"`
}

type LogConfig struct {
	Enable     bool    `yaml:"enable"`
	SampleRate float64 `yaml:"sample_rate"`
}

type GroupConfig struct {
	Group  string `yaml:"group"`
	Region string `yaml:"region"`
}

/*
TestCreateLoadBalancers tests the creation of load balancers by generating YAML
configurations, initializing Terraform, and applying the configuration. It creates
necessary infrastructure components like VPCs and instance templates, then verifies
the existence and configuration of backend services associated with the load balancers.
Resources are cleaned up after the test to ensure no lingering infrastructure remains.
*/

func TestCreateLoadBalancers(t *testing.T) {
	createLoadBalancerYAML(t) // Create YAML configurations

	tfVars := map[string]interface{}{
		"config_folder_path": configFolderPath,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars:                 tfVars,
		TerraformDir:         terraformDirectoryPath,
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})

	createVPC(t, projectID, networkName)
	time.Sleep(60 * time.Second)

	defer deleteVPC(t, projectID, networkName)
	defer deleteInstanceTemplate(t)              // Delete Instance Template after the test
	defer deleteManagedInstanceGroup(t)          // Delete MIG after the test
	defer terraform.Destroy(t, terraformOptions) // Ensure resources are cleaned up after the test

	createFirewallRule(t, projectID, networkName) // Create firewall rule
	defer deleteFirewallRule(t, projectID)        // Delete firewall rule
	createInstanceTemplate(t)                     // Create MIG Instance Template
	createManagedInstanceGroup(t)                 // Create MIG using gcloud command

	// In your TestCreateLoadBalancers function
	if _, err := terraform.InitAndApplyE(t, terraformOptions); err != nil {
		t.Fatalf("Failed to apply Terraform configuration: %v", err)
	}

	loadBalancersOutput := terraform.OutputJson(t, terraformOptions, "load_balancers") // Fetch load balancer output
	loadBalancers := gjson.Parse(loadBalancersOutput).Map()

	maxRetries := 5
	retryInterval := 15 * time.Second

	lbNameToYaml := map[string]string{
		defaultHCLBName: "instance1.yaml",
		customHCLBName:  "instance2.yaml",
	}

	for lbName := range loadBalancers {
		backendServiceName := fmt.Sprintf("%s-backend-default", lbName)

		for i := 0; i < maxRetries; i++ {
			gcloudOutput := shell.RunCommandAndGetOutput(t, shell.Command{
				Command: "gcloud",
				Args:    []string{"compute", "backend-services", "describe", backendServiceName, "--global", "--project", projectID},
			})

			if strings.Contains(gcloudOutput, "was not found") {
				t.Logf("Backend service '%s' not found yet. Retrying...", backendServiceName)
				time.Sleep(retryInterval)
				continue
			}

			t.Logf("Backend service '%s' exists.", backendServiceName)
			verifyLoadBalancerConfiguration(t, lbName, lbNameToYaml, terraformOptions)
			break
		}
	}
}

/*
createInstanceTemplate creates a Google Cloud instance template using the gcloud command.
It configures the template with specified machine type, image, network, and other parameters.
An error is logged if the creation fails.
*/

func createInstanceTemplate(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"instance-templates",
			"create",
			templateName,
			"--machine-type=n1-standard-1",
			"--image-family=ubuntu-2204-lts",
			"--image-project=ubuntu-os-cloud",
			"--network", networkName,
			"--subnet", subnetName,
			"--region", region,
			"--tags", "http-server",
			"--project", projectID,
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to create Instance Template: %v", err)
	}
}

/*
createManagedInstanceGroup creates a Managed Instance Group (MIG) in Google Cloud using
the gcloud command. It sets the base instance name, size, and template for the group.
An error is logged if the creation fails.
*/

func createManagedInstanceGroup(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"instance-groups",
			"managed",
			"create",
			migName,
			"--base-instance-name", fmt.Sprintf("%s-instance", instanceName),
			"--size", "3",
			"--template", templateName,
			"--region", region,
			"--project", projectID,
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to create Managed Instance Group: %v", err)
	}
}

/*
createLoadBalancerYAML generates YAML configuration files for health checks associated
with a Managed Instance Group. It creates both minimal and maximal health check configurations
and writes them to specified files. Errors during marshaling or file operations are logged.
*/

func createLoadBalancerYAML(t *testing.T) {
	t.Log("========= YAML Files for Health Checks =========")

	minimalHC := struct {
		Name     string `yaml:"name"`
		Project  string `yaml:"project"`
		Network  string `yaml:"network"`
		Backends struct {
			Default struct {
				Groups []struct {
					Group  string `yaml:"group"`
					Region string `yaml:"region"`
				} `yaml:"groups"`
			} `yaml:"default"`
		} `yaml:"backends"`
	}{
		Name:    "load-balancer-default-hc",
		Project: projectID,
		Network: networkName,
	}
	minimalHC.Backends.Default.Groups = append(minimalHC.Backends.Default.Groups, struct {
		Group  string `yaml:"group"`
		Region string `yaml:"region"`
	}{Group: migName, Region: region})

	yamlMinimalData, err := yaml.Marshal(&minimalHC)
	if err != nil {
		t.Errorf("Error while marshaling minimal health check: %v", err)
		return
	}

	// Create maximal health check configuration
	maximalHC := struct {
		Name     string `yaml:"name"`
		Project  string `yaml:"project"`
		Network  string `yaml:"network"`
		Backends struct {
			Default BackendConfig `yaml:"default"` // Ensure this matches the expected type
		} `yaml:"backends"`
	}{
		Name:    "load-balancer-custom-hc",
		Project: projectID,
		Network: networkName,
		Backends: struct {
			Default BackendConfig `yaml:"default"` // Match the expected type with tag
		}{
			Default: BackendConfig{
				Protocol:   "HTTP",
				Port:       80,
				PortName:   "http",
				TimeoutSec: 30,
				EnableCdn:  false,
				HealthCheck: HealthCheck{
					RequestPath: "/healthz",
					Port:        80,
				},
				LogConfig: LogConfig{
					Enable:     true,
					SampleRate: 0.5,
				},
				Groups: []GroupConfig{{Group: migName, Region: region}},
			},
		},
	}

	yamlMaximalData, err := yaml.Marshal(&maximalHC)
	if err != nil {
		t.Errorf("Error while marshaling maximal health check: %v", err)
		return
	}

	configDir := "config"
	minimalFilePath := filepath.Join(configDir, "instance1.yaml")
	maximalFilePath := filepath.Join(configDir, "instance2.yaml")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Errorf("Failed to create config directory: %v", err)
		return
	}

	if err := os.WriteFile(minimalFilePath, yamlMinimalData, 0644); err != nil {
		t.Errorf("Unable to write minimal health check data into the file: %v", err)
		return
	}
	t.Logf("Created YAML config at %s with content:\n%s", minimalFilePath, string(yamlMinimalData))

	if err := os.WriteFile(maximalFilePath, yamlMaximalData, 0644); err != nil {
		t.Errorf("Unable to write maximal health check data into the file: %v", err)
		return
	}
	t.Logf("Created YAML config at %s with content:\n%s", maximalFilePath, string(yamlMaximalData))
}

/*
verifyLoadBalancerConfiguration checks the configuration of a specified load balancer
against expected values defined in YAML files. It reads the appropriate YAML file based
on the load balancer name, unmarshals its content, and verifies the existence and
properties of the load balancer and its associated backend services in Terraform output.
It also checks health checks, self-links, and Managed Instance Groups (MIGs) for
correctness, logging any discrepancies found.
*/

func verifyLoadBalancerConfiguration(t *testing.T, lbName string, lbNameToYaml map[string]string, terraformOptions *terraform.Options) {
	// Determine which YAML file to use based on lbName
	yamlFileName, ok := lbNameToYaml[lbName]
	if !ok {
		t.Errorf("No matching YAML configuration for Load Balancer: %s", lbName)
		return
	}

	yamlFilePath := filepath.Join(configFolderPath, yamlFileName)
	t.Logf("Reading the YAML for %s from file %s", lbName, yamlFilePath)

	yamlFile, err := os.ReadFile(yamlFilePath)
	if err != nil {
		t.Errorf("Error reading YAML file %s: %s", yamlFilePath, err)
		return
	}

	var expectedLB LoadBalancerConfig
	err = yaml.Unmarshal(yamlFile, &expectedLB)
	if err != nil {
		t.Errorf("Error unmarshaling YAML for Load Balancer %s: %s", lbName, err)
		return
	}

	t.Logf("Verifying Load Balancer configuration for %s...", lbName)

	loadBalancersOutput := terraform.OutputJson(t, terraformOptions, "load_balancers")
	actualLBDetails := gjson.Parse(loadBalancersOutput).Get(lbName)

	// Check if actualLBDetails exists
	if !actualLBDetails.Exists() {
		t.Errorf("Load Balancer %s does not exist in output", lbName)
		return
	} else {
		t.Logf("Load Balancer %s does correctly exist in the output", lbName)
	}

	actualBackendServiceName := actualLBDetails.Get("backend_services.0.name").String()
	if actualBackendServiceName == "" {
		t.Errorf("Backend service name is empty for %s", lbName)
		return
	} else {
		t.Logf("Backend service %s does correctly exist in the output", actualBackendServiceName)
	}

	gcloudDescribeCmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "backend-services", "describe", actualBackendServiceName, "--global", "--project", projectID, "--format=json"},
	}
	backendServiceJSON := shell.RunCommandAndGetOutput(t, gcloudDescribeCmd)

	actualHealthChecks := gjson.Parse(backendServiceJSON).Get("healthChecks")
	if actualHealthChecks.IsArray() && len(actualHealthChecks.Array()) == 0 {
		t.Errorf("Health checks are empty for Load Balancer %s", lbName)
		return
	} else {
		t.Logf("Health checks %s correctly exist in the output", actualHealthChecks)
	}

	actualSelfLink := gjson.Parse(backendServiceJSON).Get("selfLink").String()
	expectedSelfLink := actualLBDetails.Get("backend_services.0.self_link").String()

	if actualSelfLink != expectedSelfLink {
		t.Errorf("Self link mismatch for Load Balancer %s: actual=%s, expected=%s", lbName, actualSelfLink, expectedSelfLink)
	} else {
		t.Logf("Self link verification successful for Load Balancer %s.", lbName)
	}

	// Check for Managed Instance Groups (MIG) in backends from backendServiceJSON
	migFound := false

	for _, backend := range gjson.Get(backendServiceJSON, "backends").Array() {
		groupLink := backend.Get("group").String()
		if groupLink != "" {
			migFound = true
			// Extract MIG name from the group URL
			parts := strings.Split(groupLink, "/")
			migName := parts[len(parts)-1] // Last part of the URL is the MIG name

			// Extract region from group link (part before instanceGroups)
			regionPart := parts[len(parts)-3] // The region is the third last part of the URL

			t.Logf("Managed Instance Group (MIG) found: %s in region: %s", migName, regionPart)

			// Compare extracted region with expected region
			expectedRegion := region
			if regionPart != expectedRegion {
				t.Errorf("Region mismatch for Load Balancer %s: actual=%s, expected=%s", lbName, regionPart, expectedRegion)
			} else {
				t.Logf("Region verification successful for Load Balancer %s: %s.", lbName, regionPart)
			}
		}
	}

	if !migFound {
		t.Logf("No Managed Instance Groups (MIG) found in Load Balancer %s configuration.", lbName)
	}
}

/*
createVPC creates a Virtual Private Cloud (VPC) network and a subnet in Google Cloud
using the gcloud command. It sets the VPC to custom subnet mode and specifies routing
options. Errors encountered during the execution of the commands are logged.
*/

func createVPC(t *testing.T, projectID string, networkName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{"compute", "networks", "create", networkName,
			"--project=" + projectID,
			"--format=json",
			"--bgp-routing-mode=global",
			"--subnet-mode=custom"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Logf("===Error %s Encountered while executing gcloud command to create VPC.", err)
	}

	subnetName := fmt.Sprintf("%s-subnet", networkName)
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{"compute", "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--region=" + region,
			"--range=10.0.0.0/24"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Logf("===Error %s Encountered while executing gcloud command to create subnet.", err)
	}
}

/*
deleteVPC deletes a Virtual Private Cloud (VPC) network and its associated subnet
in Google Cloud using the gcloud command. It includes delays to ensure resources
are fully released before deletion. Errors encountered during the execution of the
commands are logged.
*/

func deleteVPC(t *testing.T, projectID string, networkName string) {
	time.Sleep(120 * time.Second)

	subnetName := fmt.Sprintf("%s-subnet", networkName)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{"compute", "networks", "subnets", "delete", subnetName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while executing gcloud command to delete subnet.", err)
	}

	time.Sleep(150 * time.Second)

	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while executing gcloud command to delete VPC.", err)
	}
}

/*
deleteManagedInstanceGroup deletes a Managed Instance Group (MIG) in Google Cloud
using the gcloud command. It logs an error if the deletion fails.
*/

func deleteManagedInstanceGroup(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"instance-groups",
			"managed",
			"delete",
			migName,
			"--region=" + region,
			"--project=" + projectID,
			"--quiet",
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to delete Managed Instance Group %s: %v", migName, err)
	}
}

/*
deleteInstanceTemplate deletes an instance template in Google Cloud using the gcloud
command. An error is logged if the deletion fails.
*/

func deleteInstanceTemplate(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"instance-templates",
			"delete",
			"--project=" + projectID,
			templateName,
			"--quiet",
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to delete Instance Template: %v", err)
	}
}

func createFirewallRule(t *testing.T, projectID string, networkName string) {
	healthCheckIPRanges := []string{"130.211.0.0/22", "35.191.0.0/16"}

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"firewall-rules",
			"create",
			firewallRuleName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--allow=tcp:80", // Allow TCP on port 80
			"--source-ranges=" + strings.Join(healthCheckIPRanges, ","),
			"--target-tags=http-server", // Apply to instances with the http-server tag
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to create firewall rule: %v", err)
	}
	t.Log("Successfully created firewall rule to allow health check traffic.")
}

func deleteFirewallRule(t *testing.T, projectID string) {

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute",
			"firewall-rules",
			"delete",
			firewallRuleName,
			"--project=" + projectID,
			"--quiet",
		},
	}

	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Failed to delete firewall rule: %v", err)
	}
	t.Log("Successfully deleted firewall rule.")
}
