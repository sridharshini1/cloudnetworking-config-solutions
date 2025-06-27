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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// Test Configuration & Global Variables
var (
	// Find the project root dynamically. This assumes the test is run from its own directory.
	projectRoot, _ = filepath.Abs("../../../../../../../")

	// This points to the Terraform code that will be EXECUTED.
	ilbTerraformDirectoryPath = filepath.Join(projectRoot, "execution/07-consumer-load-balancing/Network/Passthrough/Internal")

	// This points to the 'config' folder where this test's YAML files are GENERATED.
	ilbConfigFolderPath = filepath.Join(projectRoot, "execution/test/integration/consumer-load-balancing/Network/Passthrough/Internal/config")
)

var tfVarsINLB = map[string]interface{}{
	"config_folder_path": ilbConfigFolderPath,
}

var (
	// ilbProjectID is set by the TF_VAR_project_id environment variable.
	ilbProjectID = os.Getenv("TF_VAR_project_id")
	// Dynamic names based on a random integer to ensure test isolation.
	ilbInstanceName = fmt.Sprintf("ilb-test-%d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000))
	ilbNamesToTest  = []string{
		fmt.Sprintf("lite-%s", ilbInstanceName),
		fmt.Sprintf("expanded-%s", ilbInstanceName),
	}
	ilbRegion            = "us-central1"
	ilbZone              = ilbRegion + "-a"
	ilbNetworkName       = fmt.Sprintf("vpc-%s", ilbInstanceName)
	ilbSubnetName        = fmt.Sprintf("%s-subnet", ilbNetworkName)
	ilbSubnetCidr        = "10.20.0.0/24" // Using a distinct CIDR for this test
	ilbMigName           = fmt.Sprintf("mig-%s-regional", ilbInstanceName)
	ilbTemplateName      = fmt.Sprintf("it-%s", ilbInstanceName)
	ilbFwHcRuleName      = fmt.Sprintf("%s-fw-hc", ilbNetworkName)
	ilbFwTrafficRuleName = fmt.Sprintf("%s-fw-traffic", ilbNetworkName)
	ilbTestVmName        = fmt.Sprintf("test-vm-%s", ilbInstanceName)
	ilbInstanceTag       = "ilb-backend-instance"
)

const (
	minimalILBYamlFile = "ilb-lite.yaml"
	maximalILBYamlFile = "ilb-expanded.yaml"
	apachePort         = "80"
)

// YAML Configuration Structs
type NetworkLoadBalancerConfig struct {
	Name            string                          `yaml:"name"`
	Project         string                          `yaml:"project"`
	Region          string                          `yaml:"region"`
	Network         string                          `yaml:"network"`
	Subnetwork      string                          `yaml:"subnetwork"`
	Description     string                          `yaml:"description,omitempty"`
	Backends        []BackendItemConfig             `yaml:"backends"`
	HealthCheck     *NetworkHealthCheckConfig       `yaml:"health_check,omitempty"`
	ForwardingRules map[string]ForwardingRuleConfig `yaml:"forwarding_rules"`
}

type BackendItemConfig struct {
	GroupName   string `yaml:"group_name"`
	GroupRegion string `yaml:"group_region,omitempty"`
	GroupZone   string `yaml:"group_zone,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type NetworkHealthCheckConfig struct {
	Name               string           `yaml:"name,omitempty"`
	Description        string           `yaml:"description,omitempty"`
	CheckIntervalSec   *int             `yaml:"check_interval_sec,omitempty"`
	TimeoutSec         *int             `yaml:"timeout_sec,omitempty"`
	HealthyThreshold   *int             `yaml:"healthy_threshold,omitempty"`
	UnhealthyThreshold *int             `yaml:"unhealthy_threshold,omitempty"`
	EnableLogging      *bool            `yaml:"enable_logging,omitempty"`
	TCP                *TCPHealthCheck  `yaml:"tcp,omitempty"`
	HTTP               *HTTPHealthCheck `yaml:"http,omitempty"`
}

type TCPHealthCheck struct {
	Port              *int   `yaml:"port,omitempty"`
	PortSpecification string `yaml:"port_specification,omitempty"`
}

type HTTPHealthCheck struct {
	Port        *int   `yaml:"port,omitempty"`
	RequestPath string `yaml:"request_path,omitempty"`
}

type ForwardingRuleConfig struct {
	Address     string   `yaml:"address,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Protocol    string   `yaml:"protocol,omitempty"`
}

/*
TestInitAndPlanRunWithTfVarsINLB tests Terraform initialization and planning
for the Internal Network Load Balancer module with specified variables.
It expects an exit code of 2, indicating that changes are planned.
*/
func TestInitAndPlanRunWithTfVarsINLB(t *testing.T) {
	createInternalLoadBalancerYAML(t)
	createVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName, ilbSubnetCidr)
	defer deleteVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ilbTerraformDirectoryPath,
		Vars:         tfVarsINLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-inlb", // Use a distinct plan file name
		NoColor:      true,
	})

	// Run 'terraform init' and 'terraform plan', get the exit code.
	planExitCode := terraform.InitAndPlanWithExitCode(t, terraformOptions)
	want := 2 // Expect changes to be applied (exit code 2 means plan has changes)
	got := planExitCode

	// Check if the actual exit code matches the expected one.
	if got != want {
		planJSON := terraform.Show(t, terraformOptions)
		t.Logf("Plan output for TestInitAndPlanRunWithTfVarsINLB: %s", planJSON)
		t.Errorf("TestInitAndPlanRunWithTfVarsINLB: Plan Exit Code = %v, want = %v", got, want)
	}
}

/*
TestResourcesCountINLB verifies the number of resources planned by Terraform for the INLB module.
It initializes Terraform, creates a plan, and checks that the total resource count
to be added matches the expected value based on the test YAML files.

The expected count is based on the resources created by the GoogleCloudPlatform/lb-internal/google module.
Typically, this includes:
- 1 google_compute_region_backend_service
- 1 google_compute_health_check
- 1 google_compute_forwarding_rule
Total = 3 resources per NLB instance.
*/
func TestResourcesCountINLB(t *testing.T) {
	createInternalLoadBalancerYAML(t)
	createVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName, ilbSubnetCidr)
	defer deleteVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ilbTerraformDirectoryPath,
		Vars:         tfVarsINLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-inlb", // Use a distinct plan file name
		NoColor:      true,
	})

	// Initialize and create a plan, then parse the resource count.
	planStruct := terraform.InitAndPlan(t, terraformOptions)
	resourceCount := terraform.GetResourceCount(t, planStruct)

	numberOfINLBs := len(ilbNamesToTest)
	resourcesPerINLB := 3
	expectedResourceAddCount := numberOfINLBs * resourcesPerINLB

	if got, want := resourceCount.Add, expectedResourceAddCount; got != want {
		planJSON := terraform.Show(t, terraformOptions)
		t.Logf("Plan output: %s", planJSON) // Log the plan for inspection
		t.Errorf("TestResourcesCountINLB: Resource Count Add = %v, want = %v (based on %d INLB configs)", got, want, numberOfINLBs)
	}
	if got, want := resourceCount.Change, 0; got != want {
		t.Errorf("TestResourcesCountINLB: Resource Count Change = %v, want = %v", got, want)
	}
	if got, want := resourceCount.Destroy, 0; got != want {
		t.Errorf("TestResourcesCountINLB: Resource Count Destroy = %v, want = %v", got, want)
	}
}

/*
TestTerraformModuleINLBResourceAddressListMatch checks that the module addresses for
Internal Load Balancer resources in the Terraform plan match the expected addresses
derived from YAML configuration files. It looks for module instances named 'module.inlb_passthrough'.
*/
func TestTerraformModuleINLBResourceAddressListMatch(t *testing.T) {
	createInternalLoadBalancerYAML(t)
	createVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName, ilbSubnetCidr)
	defer deleteVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName)

	expectedModuleAddresses := make(map[string]struct{})
	for _, name := range ilbNamesToTest {
		address := fmt.Sprintf("module.internal_passthrough_nlb[\"%s\"]", name)
		expectedModuleAddresses[address] = struct{}{}
	}

	// Initialize Terraform and generate a plan.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ilbTerraformDirectoryPath,
		Vars:         tfVarsINLB,
		Reconfigure:  true,
		Lock:         true,
		PlanFilePath: "./plan-inlb-addressmatch", // Use a distinct plan file name
		NoColor:      true,
	})

	planJSON := terraform.InitAndPlanAndShow(t, terraformOptions)
	content, err := terraform.ParsePlanJSON(planJSON)
	if err != nil {
		t.Fatalf("Failed to parse plan JSON: %v", err)
	}

	actualModuleAddresses := make(map[string]struct{})
	for _, resourceChange := range content.ResourceChangesMap {
		moduleAddr := resourceChange.ModuleAddress
		if strings.HasPrefix(moduleAddr, "module.internal_passthrough_nlb[") {
			actualModuleAddresses[moduleAddr] = struct{}{}
		}
	}

	if !assert.Equal(t, expectedModuleAddresses, actualModuleAddresses) {
		t.Logf("Full Plan JSON for TestTerraformModuleINLBResourceAddressListMatch: %s", planJSON)
	}
}

// TestCreateInternalLoadBalancer tests the full lifecycle of an Internal Passthrough Network Load Balancer.
func TestCreateInternalLoadBalancer(t *testing.T) {
	t.Parallel()

	if ilbProjectID == "" {
		t.Fatal("TF_VAR_project_id must be set as an environment variable.")
	}

	// 1. SETUP: Generate dynamic YAML configs for different test cases.
	createInternalLoadBalancerYAML(t)

	// 2. SETUP: Create all prerequisite cloud resources using gcloud commands.
	createVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName, ilbSubnetCidr)
	defer deleteVPC(t, ilbProjectID, ilbNetworkName, ilbRegion, ilbSubnetName)

	ilbFwIapRuleName := fmt.Sprintf("%s-fw-iap-ssh", ilbNetworkName)
	createFirewallRuleForIAP(t, ilbProjectID, ilbNetworkName, ilbFwIapRuleName, []string{"allow-iap-ssh"})
	defer deleteFirewallRule(t, ilbProjectID, ilbFwIapRuleName)

	createFirewallRuleForILBTraffic(t, ilbProjectID, ilbNetworkName, ilbFwTrafficRuleName, []string{apachePort}, []string{ilbInstanceTag}, ilbSubnetCidr)
	defer deleteFirewallRule(t, ilbProjectID, ilbFwTrafficRuleName)

	createFirewallRuleForNLBHealthChecks(t, ilbProjectID, ilbNetworkName, ilbFwHcRuleName, []string{ilbInstanceTag})
	defer deleteFirewallRule(t, ilbProjectID, ilbFwHcRuleName)

	createInstanceTemplate(t, ilbProjectID, ilbTemplateName, ilbNetworkName, ilbSubnetName, ilbRegion, []string{ilbInstanceTag, "allow-iap-ssh"})
	defer deleteInstanceTemplate(t, ilbProjectID, ilbTemplateName)

	createManagedInstanceGroup(t, ilbProjectID, ilbRegion, "", ilbMigName, ilbTemplateName, 2)
	defer deleteManagedInstanceGroup(t, ilbProjectID, ilbRegion, "", ilbMigName)
	setNamedPortsOnMIG(t, ilbProjectID, ilbRegion, "", ilbMigName, "http", apachePort)

	createTestVM(t, ilbProjectID, ilbZone, ilbTestVmName, ilbNetworkName, ilbSubnetName)
	defer deleteTestVM(t, ilbProjectID, ilbZone, ilbTestVmName)

	// 3. EXECUTION: Run terraform init and apply.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:         ilbTerraformDirectoryPath,
		Vars:                 map[string]interface{}{"config_folder_path": ilbConfigFolderPath},
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})
	defer terraform.Destroy(t, terraformOptions)

	_, err := terraform.InitAndApplyE(t, terraformOptions)
	if !assert.NoError(t, err, "Terraform apply failed for ILB") {
		t.FailNow()
	}

	// 4. VERIFICATION: Fetch Terraform outputs and run verification checks.
	ilbForwardingRuleAddresses := terraform.OutputJson(t, terraformOptions, "ilb_forwarding_rule_addresses")
	assert.True(t, gjson.Valid(ilbForwardingRuleAddresses), "Output 'ilb_forwarding_rule_addresses' is not valid JSON")

	loadBalancersToTest := gjson.Parse(ilbForwardingRuleAddresses).Map()
	if !assert.NotEmpty(t, loadBalancersToTest, "No load balancers found in the output") {
		t.FailNow()
	}

	// Get details needed for verification checks.
	testVmIp := getVmInternalIp(t, ilbProjectID, ilbZone, ilbTestVmName)
	backendInstances := getMigInstances(t, ilbProjectID, ilbRegion, ilbMigName)
	if !assert.NotEmpty(t, backendInstances, "Could not retrieve backend instances from MIG") {
		t.FailNow()
	}

	// Loop through each load balancer created by Terraform and verify it.
	for lbNameFromOutput, fwdRules := range loadBalancersToTest {
		t.Run(lbNameFromOutput, func(t *testing.T) {
			t.Logf("--- Starting Verification for Internal Load Balancer: %s ---", lbNameFromOutput)

			yamlFileName := getYamlFileForTest(lbNameFromOutput)
			if !assert.NotEmpty(t, yamlFileName, "Could not determine YAML file for LB: %s", lbNameFromOutput) {
				return
			}

			verifyInternalLoadBalancerConfiguration(t, lbNameFromOutput, yamlFileName, terraformOptions)

			// Loop through each forwarding rule associated with the load balancer.
			fwdRules.ForEach(func(ruleKey, ipAddress gjson.Result) bool {
				t.Logf("Verifying rule '%s' with IP: %s", ruleKey.String(), ipAddress.String())
				if !assert.NotEmpty(t, ipAddress.String(), "IP address for ILB %s (rule '%s') is empty", lbNameFromOutput, ruleKey.String()) {
					return true // continue to next rule
				}

				verifyPassthroughResponse(t, ilbProjectID, ilbZone, ilbTestVmName, testVmIp, ipAddress.String(), apachePort)
				return true // continue ForEach
			})
			t.Logf("--- Finished Verification for Internal Load Balancer: %s ---", lbNameFromOutput)
		})
	}
}

// YAML Generation Function
// createInternalLoadBalancerYAML generates the YAML config files for the ILB tests.
func createInternalLoadBalancerYAML(t *testing.T) {
	t.Log("========= Generating YAML Files for Internal Load Balancers =========")

	err := os.RemoveAll(ilbConfigFolderPath)
	assert.NoError(t, err, "Failed to remove existing ILB config directory")

	err = os.MkdirAll(ilbConfigFolderPath, 0755)
	assert.NoError(t, err, "Failed to create ILB config directory %s", ilbConfigFolderPath)

	minimalILBCfg := NetworkLoadBalancerConfig{
		Name:       ilbNamesToTest[0],
		Project:    ilbProjectID,
		Region:     ilbRegion,
		Network:    ilbNetworkName,
		Subnetwork: ilbSubnetName,
		Backends: []BackendItemConfig{
			{GroupName: ilbMigName, GroupRegion: ilbRegion},
		},
	}
	yamlMinimalData, err := yaml.Marshal(&minimalILBCfg)
	assert.NoError(t, err, "Error marshaling lite ILB config")
	minimalFilePath := filepath.Join(ilbConfigFolderPath, minimalILBYamlFile)
	err = os.WriteFile(minimalFilePath, yamlMinimalData, 0644)
	assert.NoError(t, err, "Unable to write lite ILB config")
	t.Logf("Created Lite ILB YAML config at %s", minimalFilePath)

	// Expanded ILB Configuration: A more complex setup with a custom health check.
	hcTimeout, hcCheckInterval := 5, 10
	hcEnableLogging := true

	maximalILBCfg := NetworkLoadBalancerConfig{
		Name:        ilbNamesToTest[1],
		Project:     ilbProjectID,
		Region:      ilbRegion,
		Network:     ilbNetworkName,
		Subnetwork:  ilbSubnetName,
		Description: "Expanded ILB with custom HC",
		Backends: []BackendItemConfig{
			{GroupName: ilbMigName, GroupRegion: ilbRegion},
		},
		HealthCheck: &NetworkHealthCheckConfig{
			Description:      "Custom Health Check for ILB",
			CheckIntervalSec: &hcCheckInterval,
			TimeoutSec:       &hcTimeout,
			EnableLogging:    &hcEnableLogging,
			TCP:              &TCPHealthCheck{PortSpecification: "USE_SERVING_PORT"},
		},
		ForwardingRules: map[string]ForwardingRuleConfig{
			"expanded-rule": {
				Protocol:    "TCP",
				Ports:       []string{apachePort},
				Description: "Forwarding rule for HTTP traffic",
			},
		},
	}
	yamlMaximalData, err := yaml.Marshal(&maximalILBCfg)
	assert.NoError(t, err, "Error marshaling expanded ILB config")
	maximalFilePath := filepath.Join(ilbConfigFolderPath, maximalILBYamlFile)
	err = os.WriteFile(maximalFilePath, yamlMaximalData, 0644)
	assert.NoError(t, err, "Unable to write expanded ILB config")
	t.Logf("Created Expanded ILB YAML config at %s", maximalFilePath)
}

// Verification Functions
func verifyInternalLoadBalancerConfiguration(t *testing.T, lbNameFromOutput string, yamlFileName string, opts *terraform.Options) {
	t.Logf("Verifying ILB configuration for: %s using YAML: %s", lbNameFromOutput, yamlFileName)

	frOutput := terraform.OutputJson(t, opts, "ilb_forwarding_rules")
	actualFwdRuleMap := gjson.Parse(frOutput).Get(lbNameFromOutput)

	var frSelfLink string
	actualFwdRuleMap.ForEach(func(key, value gjson.Result) bool {
		frSelfLink = value.String()
		return false
	})

	if !assert.NotEmpty(t, frSelfLink, "Could not find forwarding rule self_link in output for %s", lbNameFromOutput) {
		return
	}

	// Describe the resource using the self_link
	frPathParts := strings.Split(frSelfLink, "/")
	frName := frPathParts[len(frPathParts)-1]
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "forwarding-rules", "describe", frName, "--region", ilbRegion, "--project", ilbProjectID, "--format=json"}}
	detailsStr, errFr := shell.RunCommandAndGetOutputE(t, cmd)
	if !assert.NoError(t, errFr, "Failed to describe forwarding rule %s", frName) {
		return
	}

	frDetails := gjson.Parse(detailsStr)
	assert.Equal(t, "INTERNAL", frDetails.Get("loadBalancingScheme").String(), "ILB %s: FR scheme should be INTERNAL", lbNameFromOutput)
	t.Logf("Successfully verified Forwarding Rule '%s' for ILB '%s' is INTERNAL.", frName, lbNameFromOutput)
}

// verifyPassthroughResponse connects from a test VM to the load balancer and asserts that
// the response from the backend is the test VM's own internal IP address.
// This elegantly verifies both connectivity and client IP preservation in a single check.
func verifyPassthroughResponse(t *testing.T, projectID, zone, testVmName, testVmIp, lbIpAddress, port string) {
	t.Logf("Verifying passthrough from VM %s (%s) to ILB IP %s", testVmName, testVmIp, lbIpAddress)

	maxRetries := 3
	sleepBetweenRetries := 15 * time.Second

	// The command to run on the test VM
	sshCommand := fmt.Sprintf("curl -s --fail -m 10 http://%s:%s", lbIpAddress, port)

	// Use a retry loop to handle the time it takes for the LB and backends to become healthy.
	_, err := retry.DoWithRetryE(t, "Check for client IP in response", maxRetries, sleepBetweenRetries, func() (string, error) {
		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "ssh", testVmName,
				"--project=" + projectID, "--zone=" + zone, "--command=" + sshCommand,
				"--tunnel-through-iap", "--quiet",
			},
		}
		response, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			if strings.Contains(response, "gcloud crashed") {
				return "", fmt.Errorf("gcloud command crashed: %s", response)
			}
			return "", fmt.Errorf("command failed: %v. Output: %s", err, response)
		}

		// The gcloud ssh command includes warnings on stderr, which get combined with the response.
		// We need to extract the actual response from the last line of the output string.
		lines := strings.Split(strings.TrimSpace(response), "\n")
		actualResponse := ""
		if len(lines) > 0 {
			// The last line of the output will be the actual response from curl.
			actualResponse = lines[len(lines)-1]
		}

		// The response should be the client's (the test VM's) IP address.
		if strings.TrimSpace(actualResponse) != testVmIp {
			return "", fmt.Errorf("response did not match client IP. Expected: '%s', Got: '%s'. Full output: %s", testVmIp, actualResponse, response)
		}

		t.Logf("Success! Response from load balancer matched client IP: %s", actualResponse)
		return actualResponse, nil
	})

	assert.NoError(t, err, "Failed to verify passthrough response after multiple retries.")
}

// Prerequisite Infrastructure Helper Functions (gcloud Wrappers)
func createVPC(t *testing.T, projectID, networkName, region, subnetName, subnetCidr string) {
	t.Logf("Creating VPC '%s' and subnet '%s'", networkName, subnetName)
	vpcCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "create", networkName, "--project=" + projectID, "--subnet-mode=custom", "--bgp-routing-mode=global"}}
	_, err := shell.RunCommandAndGetOutputE(t, vpcCmd)
	assert.NoError(t, err, "Error creating VPC")

	subnetCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", subnetName, "--project=" + projectID, "--network=" + networkName, "--region=" + region, "--range=" + subnetCidr}}
	_, err = shell.RunCommandAndGetOutputE(t, subnetCmd)
	assert.NoError(t, err, "Error creating subnet")
	t.Logf("Successfully created VPC and Subnet.")
}

func deleteVPC(t *testing.T, projectID, networkName, region, subnetName string) {
	t.Logf("Deleting VPC '%s' and subnet '%s'", networkName, subnetName)

	// Delete subnet
	subnetCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", subnetName, "--project=" + projectID, "--region=" + region, "--quiet"}}
	if err := shell.RunCommandE(t, subnetCmd); err != nil {
		t.Logf("Warning: failed to delete subnet %s. Error: %v", subnetName, err)
	}

	time.Sleep(5 * time.Second)

	// Delete VPC network
	vpcCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"}}
	if err := shell.RunCommandE(t, vpcCmd); err != nil {
		t.Logf("Warning: failed to delete network %s. Error: %v", networkName, err)
	}
}

func createFirewallRuleForILBTraffic(t *testing.T, projectID, network, ruleName string, ports, tags []string, sourceCidr string) {
	t.Logf("Creating firewall rule '%s' to allow traffic from '%s'", ruleName, sourceCidr)
	rules := "tcp:" + strings.Join(ports, ",tcp:")
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", ruleName, "--project=" + projectID, "--network=" + network, "--action=ALLOW", "--rules=" + rules, "--source-ranges=" + sourceCidr, "--target-tags=" + strings.Join(tags, ",")}}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to create firewall rule for ILB traffic")
}

func createFirewallRuleForNLBHealthChecks(t *testing.T, projectID, network, ruleName string, tags []string) {
	t.Logf("Creating firewall rule '%s' for GCP health checkers", ruleName)
	// These are Google's public IP ranges for health checkers.
	hcRanges := "130.211.0.0/22,35.191.0.0/16,209.85.152.0/22,209.85.204.0/22"
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", ruleName, "--project=" + projectID, "--network=" + network, "--action=ALLOW", "--rules=tcp", "--source-ranges=" + hcRanges, "--target-tags=" + strings.Join(tags, ",")}}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to create firewall rule for health checks")
}

func createFirewallRuleForIAP(t *testing.T, projectID, network, ruleName string, tags []string) {
	t.Logf("Creating firewall rule '%s' for IAP SSH", ruleName)
	iapRange := "35.235.240.0/20"
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", ruleName, "--project=" + projectID, "--network=" + network, "--action=ALLOW", "--direction=INGRESS", "--rules=tcp:22", "--source-ranges=" + iapRange, "--target-tags=" + strings.Join(tags, ",")}}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to create firewall rule for IAP")
}

func deleteFirewallRule(t *testing.T, projectID, ruleName string) {
	t.Logf("Deleting firewall rule '%s'", ruleName)
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", ruleName, "--project=" + projectID, "--quiet"}}
	if err := shell.RunCommandE(t, cmd); err != nil {
		t.Logf("Warning: failed to delete firewall rule %s. This may be expected if the test failed before creation. Error: %v", ruleName, err)
	}
}

func createInstanceTemplate(t *testing.T, projectID, templateName, network, subnet, region string, tags []string) {
	t.Logf("Creating instance template '%s' with an IP echo server", templateName)

	// This startup script runs a simple Python web server that echoes the client's IP address.
	startupScript := `#!/bin/bash
apt-get update -y
apt-get install -y python3
cat <<EOF > /echo_server.py
import http.server
import socketserver

class MyHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        client_ip = self.client_address[0]
        self.wfile.write(bytes(client_ip, "utf8"))

with socketserver.TCPServer(("", 80), MyHandler) as httpd:
    print("serving at port 80")
    httpd.serve_forever()
EOF
nohup python3 /echo_server.py > /dev/null 2>&1 &`

	scriptFileName := "startup-script.sh"
	scriptPath := filepath.Join(ilbConfigFolderPath, scriptFileName)
	err := os.WriteFile(scriptPath, []byte(startupScript), 0755)
	assert.NoError(t, err, "Failed to write startup script to file")

	// Use --metadata-from-file to pass the script, which is more robust than passing a long string.
	metadataFlag := fmt.Sprintf("startup-script=%s", scriptPath)

	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instance-templates", "create", templateName, "--project=" + projectID, "--machine-type=e2-small", "--image-family=debian-11", "--image-project=debian-cloud", "--network=" + network, "--subnet=" + subnet, "--region=" + region, "--tags=" + strings.Join(tags, ","), "--metadata-from-file", metadataFlag}}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to create Instance Template")
}

func deleteInstanceTemplate(t *testing.T, projectID, templateName string) {
	t.Logf("Deleting instance template '%s'", templateName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "instance-templates", "delete", templateName, "--project=" + projectID, "--quiet"}})
}

func createManagedInstanceGroup(t *testing.T, projectID, region, zone, migName, templateName string, size int) {
	t.Logf("Creating MIG '%s'", migName)
	args := []string{"compute", "instance-groups", "managed", "create", migName, "--project=" + projectID, "--base-instance-name", migName, "--size", fmt.Sprintf("%d", size), "--template", templateName}
	if zone != "" {
		args = append(args, "--zone="+zone)
	} else {
		args = append(args, "--region="+region)
	}
	cmd := shell.Command{Command: "gcloud", Args: args}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to create MIG")
	t.Logf("Waiting 120s for MIG '%s' to stabilize...", migName)
	time.Sleep(120 * time.Second)
}

func deleteManagedInstanceGroup(t *testing.T, projectID, region, zone, migName string) {
	t.Logf("Deleting MIG '%s'", migName)
	args := []string{"compute", "instance-groups", "managed", "delete", migName, "--project=" + projectID, "--quiet"}
	if zone != "" {
		args = append(args, "--zone="+zone)
	} else {
		args = append(args, "--region="+region)
	}
	cmd := shell.Command{Command: "gcloud", Args: args}

	if err := shell.RunCommandE(t, cmd); err != nil {
		t.Logf("Warning: failed to delete MIG %s. Error: %v", migName, err)
	}
}

func setNamedPortsOnMIG(t *testing.T, projectID, region, zone, migName, portName, portNumber string) {
	t.Logf("Setting named port '%s:%s' on MIG '%s'", portName, portNumber, migName)
	args := []string{"compute", "instance-groups", "managed", "set-named-ports", migName, "--project=" + projectID, fmt.Sprintf("--named-ports=%s:%s", portName, portNumber)}
	if zone != "" {
		args = append(args, "--zone="+zone)
	} else {
		args = append(args, "--region="+region)
	}
	cmd := shell.Command{Command: "gcloud", Args: args}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to set named ports on MIG")
}

func createTestVM(t *testing.T, projectID, zone, vmName, network, subnet string) {
	t.Logf("Creating test VM '%s'", vmName)
	// Startup script installs necessary tools for connectivity checks.
	startupScript := "apt-get update -y && apt-get install -y curl dnsutils netcat-openbsd"
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "create", vmName, "--project=" + projectID, "--zone=" + zone, "--machine-type=e2-micro", "--image-family=debian-11", "--image-project=debian-cloud", "--network=" + network, "--subnet=" + subnet, "--tags=allow-iap-ssh", "--metadata=startup-script=" + startupScript}}
	_, err := retry.DoWithRetryE(t, "Create Test VM", 2, 10*time.Second, func() (string, error) {
		return shell.RunCommandAndGetOutputE(t, cmd)
	})
	assert.NoError(t, err, "Failed to create test VM")
	t.Logf("Waiting 60s for test VM to boot and run startup script...")
	time.Sleep(60 * time.Second)
}

func deleteTestVM(t *testing.T, projectID, zone, vmName string) {
	t.Logf("Deleting test VM '%s'", vmName)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "instances", "delete", vmName, "--project=" + projectID, "--zone=" + zone, "--quiet"},
	}

	if err := shell.RunCommandE(t, cmd); err != nil {
		t.Logf("Warning: failed to delete VM %s. Error: %v", vmName, err)
	}
}

func getVmInternalIp(t *testing.T, projectID, zone, vmName string) string {
	t.Logf("Getting internal IP for VM '%s'", vmName)
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "describe", vmName, "--project=" + projectID, "--zone=" + zone, "--format=get(networkInterfaces[0].networkIP)"}}
	ip, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NoError(t, err, "Failed to get internal IP for VM")
	t.Logf("Found internal IP for test VM %s: %s", vmName, ip)
	return ip
}

func getMigInstances(t *testing.T, projectID, region, migName string) map[string]string {
	t.Logf("Getting instance names and zones from MIG '%s'", migName)
	args := []string{
		"compute", "instance-groups", "managed", "list-instances", migName,
		"--project=" + projectID,
		"--region=" + region,
		"--format=json",
	}
	output, err := shell.RunCommandAndGetOutputE(t, shell.Command{Command: "gcloud", Args: args})
	assert.NoError(t, err, "Failed to list instances for MIG")

	instanceDetails := make(map[string]string)
	parsedOutput := gjson.Parse(output)
	parsedOutput.ForEach(func(key, value gjson.Result) bool {
		instanceURL := value.Get("instance").String()
		parts := strings.Split(instanceURL, "/")
		instanceName := parts[len(parts)-1]
		zone := parts[len(parts)-3]
		instanceDetails[instanceName] = zone
		return true
	})

	t.Logf("Found instances in MIG %s: %v", migName, instanceDetails)
	return instanceDetails
}

// getYamlFileForTest determines the source YAML filename based on the LB's name prefix.
func getYamlFileForTest(lbName string) string {
	if strings.Contains(lbName, "lite") {
		return minimalILBYamlFile
	}
	if strings.Contains(lbName, "expanded") {
		return maximalILBYamlFile
	}
	return ""
}
