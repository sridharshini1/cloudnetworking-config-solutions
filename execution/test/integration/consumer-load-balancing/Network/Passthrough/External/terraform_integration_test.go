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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// Test configuration for Network Load Balancer
var (
	nlbProjectRoot, _         = filepath.Abs("../../../../../../../")
	nlbTerraformDirectoryPath = filepath.Join(nlbProjectRoot, "execution/07-consumer-load-balancing/Network/Passthrough/External")
	nlbConfigFolderPath       = filepath.Join(nlbProjectRoot, "execution/test/integration/consumer-load-balancing/Network/Passthrough/External/config")
)

var (
	// nlbProjectID is set by TF_VAR_project_id environment variable
	nlbProjectID         = os.Getenv("TF_VAR_project_id")
	nlbInstanceName      = fmt.Sprintf("nlb-%d", rand.Int())
	nlbRegion            = "us-central1" // Or make this configurable
	nlbZone              = nlbRegion + "-a"
	nlbNetworkName       = fmt.Sprintf("vpc-%s-test", nlbInstanceName)
	nlbSubnetName        = fmt.Sprintf("%s-subnet", nlbNetworkName)
	nlbMigName           = fmt.Sprintf("mig-%s-nlb", nlbInstanceName)
	nlbTemplateName      = fmt.Sprintf("%s-it-nlb", nlbInstanceName)
	nlbFwHcRuleName      = fmt.Sprintf("%s-fw-hc", nlbNetworkName)
	nlbFwTrafficRuleName = fmt.Sprintf("%s-fw-traffic", nlbNetworkName)
	nlbTestVmName        = fmt.Sprintf("test-vm-%s-nlb", nlbInstanceName)
	nlbInstanceTag       = "nlb-backend-instance"
)

const (
	minimalNLBYamlFile = "nlb-lite.yaml"
	maximalNLBYamlFile = "nlb-expanded.yaml"
	apachePort         = "80" // Port Apache listens on in the MIG instances
)

// NetworkLoadBalancerConfig struct for YAML parsing
type NetworkLoadBalancerConfig struct {
	Name            string                          `yaml:"name"`
	ProjectID       string                          `yaml:"project_id"`
	Region          string                          `yaml:"region"`
	Description     string                          `yaml:"description,omitempty"`
	Backends        []BackendItemConfig             `yaml:"backends"` // Slice for backend groups
	HealthCheck     *NetworkHealthCheckConfig       `yaml:"health_check,omitempty"`
	ForwardingRules map[string]ForwardingRuleConfig `yaml:"forwarding_rules,omitempty"`
}

// For individual backend group items
type BackendItemConfig struct {
	GroupName                    string                    `yaml:"group_name"`
	GroupRegion                  string                    `yaml:"group_region,omitempty"`
	Failover                     *bool                     `yaml:"failover,omitempty"`
	Description                  string                    `yaml:"description,omitempty"`
	Protocol                     string                    `yaml:"protocol,omitempty"` // TCP, UDP
	PortName                     string                    `yaml:"port_name,omitempty"`
	TimeoutSec                   *int                      `yaml:"timeout_sec,omitempty"`
	ConnectionDrainingTimeoutSec *int                      `yaml:"connection_draining_timeout_sec,omitempty"`
	LogSampleRate                *float64                  `yaml:"log_sample_rate,omitempty"`
	LocalityLbPolicy             string                    `yaml:"locality_lb_policy,omitempty"`
	SessionAffinity              string                    `yaml:"session_affinity,omitempty"`
	ConnectionTracking           *ConnectionTrackingConfig `yaml:"connection_tracking,omitempty"`
	FailoverConfig               *BackendFailoverConfig    `yaml:"failover_config,omitempty"`
}

type ConnectionTrackingConfig struct {
	IdleTimeoutSec         *int   `yaml:"idle_timeout_sec,omitempty"`
	PersistConnOnUnhealthy string `yaml:"persist_conn_on_unhealthy,omitempty"`
	TrackPerSession        *bool  `yaml:"track_per_session,omitempty"`
}

type BackendFailoverConfig struct {
	DisableConnDrain       *bool    `yaml:"disable_conn_drain,omitempty"`
	DropTrafficIfUnhealthy *bool    `yaml:"drop_traffic_if_unhealthy,omitempty"`
	Ratio                  *float64 `yaml:"ratio,omitempty"`
}

type NetworkHealthCheckConfig struct {
	Name               string           `yaml:"name,omitempty"`
	CheckIntervalSec   *int             `yaml:"check_interval_sec,omitempty"`
	TimeoutSec         *int             `yaml:"timeout_sec,omitempty"`
	HealthyThreshold   *int             `yaml:"healthy_threshold,omitempty"`
	UnhealthyThreshold *int             `yaml:"unhealthy_threshold,omitempty"`
	EnableLogging      *bool            `yaml:"enable_logging,omitempty"`
	Description        string           `yaml:"description,omitempty"`
	TCP                *TCPHealthCheck  `yaml:"tcp,omitempty"`
	HTTP               *HTTPHealthCheck `yaml:"http,omitempty"`
	HTTPS              *HTTPHealthCheck `yaml:"https,omitempty"`
	HTTP2              *HTTPHealthCheck `yaml:"http2,omitempty"`
	GRPC               *GRPCHealthCheck `yaml:"grpc,omitempty"`
	SSL                *SSLHealthCheck  `yaml:"ssl,omitempty"`
}

type TCPHealthCheck struct {
	Port              *int   `yaml:"port,omitempty"`
	PortSpecification string `yaml:"port_specification,omitempty"`
	Request           string `yaml:"request,omitempty"`
	Response          string `yaml:"response,omitempty"`
	ProxyHeader       string `yaml:"proxy_header,omitempty"`
}

type HTTPHealthCheck struct {
	Port              *int   `yaml:"port,omitempty"`
	PortSpecification string `yaml:"port_specification,omitempty"`
	RequestPath       string `yaml:"request_path,omitempty"`
	ProxyHeader       string `yaml:"proxy_header,omitempty"`
	Response          string `yaml:"response,omitempty"`
}

type SSLHealthCheck struct {
	Port              *int   `yaml:"port,omitempty"`
	PortSpecification string `yaml:"port_specification,omitempty"`
	Request           string `yaml:"request,omitempty"`
	Response          string `yaml:"response,omitempty"`
	ProxyHeader       string `yaml:"proxy_header,omitempty"`
}

type GRPCHealthCheck struct {
	Port              *int   `yaml:"port,omitempty"`
	PortSpecification string `yaml:"port_specification,omitempty"`
	ServiceName       string `yaml:"service_name,omitempty"`
}

type ForwardingRuleConfig struct {
	Address     string   `yaml:"address,omitempty"`
	Description string   `yaml:"description,omitempty"`
	IPv6        *bool    `yaml:"ipv6,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Protocol    string   `yaml:"protocol,omitempty"`
	Subnetwork  string   `yaml:"subnetwork,omitempty"`
}

// TestCreateNetworkLoadBalancer tests the creation and verification of Network Load Balancers
func TestCreateNetworkLoadBalancer(t *testing.T) {
	t.Parallel() // Enable parallel execution if desired, ensure unique names

	if nlbProjectID == "" {
		t.Fatal("TF_VAR_project_id must be set as an environment variable.")
	}

	createNetworkLoadBalancerYAML(t) // Create YAML configurations

	tfVars := map[string]interface{}{
		"config_folder_path": nlbConfigFolderPath,
		// Add any other tfvars overrides needed for the NLB module specifically for testing
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:         nlbTerraformDirectoryPath,
		Vars:                 tfVars,
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})

	nlbFwIapRuleName := fmt.Sprintf("%s-fw-iap-ssh", nlbNetworkName) // IAP firewall rule name

	// Setup common infrastructure
	createVPC(t, nlbProjectID, nlbNetworkName)
	time.Sleep(60 * time.Second) // Allow VPC to be ready

	// Defer cleanup of all resources - LIFO order
	// Note: terraform.Destroy should handle most resources defined in TF.
	// Manual gcloud deletes are for resources created outside TF or as a fallback.
	// 1. (Runs Last) Delete VPC
	defer deleteVPC(t, nlbProjectID, nlbNetworkName)

	// 2. Delete manually created Firewall Rules (after NLBs/MIGs using tags are gone)
	defer deleteFirewallRule(t, nlbProjectID, nlbFwIapRuleName)
	defer deleteFirewallRule(t, nlbProjectID, nlbFwTrafficRuleName)
	defer deleteFirewallRule(t, nlbProjectID, nlbFwHcRuleName)

	// 3. Delete Test VM
	defer deleteTestVM(t, nlbProjectID, nlbZone, nlbTestVmName)

	// 4. (Runs First) Delete Instance Template (Needed by MIG)
	defer deleteInstanceTemplateNLB(t)

	// 5. Delete MIG (Prerequisite for Terraform apply, needed by BS)
	defer deleteManagedInstanceGroupNLB(t)

	// 6. Terraform Destroy (Handles NLB, Backend Service, Health Check)
	defer terraform.Destroy(t, terraformOptions)

	// Create resources needed for NLB (firewalls, instance template, MIG)
	// These are prerequisites for the Terraform module to apply successfully or for testing.
	createFirewallRuleForNLBHealthChecks(t, nlbProjectID, nlbNetworkName, nlbFwHcRuleName, []string{nlbInstanceTag})
	createFirewallRuleForNLBTraffic(t, nlbProjectID, nlbNetworkName, nlbFwTrafficRuleName, []string{apachePort, "9000"}, []string{nlbInstanceTag})
	createFirewallRuleForIAP(t, nlbProjectID, nlbNetworkName, nlbFwIapRuleName, []string{"allow-iap-ssh"})

	createInstanceTemplate(t, nlbTemplateName, nlbProjectID, nlbNetworkName, nlbSubnetName, nlbRegion, []string{nlbInstanceTag})
	createManagedInstanceGroupNLB(t) // Creates 'nlbMigName'
	setNamedPortsOnMIG(t, nlbProjectID, nlbRegion, nlbMigName, "http", apachePort)

	// Init and Apply Terraform configuration for the NLB
	if _, err := terraform.InitAndApplyE(t, terraformOptions); err != nil {
		t.Fatalf("Failed to apply Terraform configuration for NLB: %v", err)
	}

	// Fetch Terraform outputs
	nlbForwardingRuleAddresses := terraform.OutputJson(t, terraformOptions, "nlb_forwarding_rule_addresses")
	if !gjson.Valid(nlbForwardingRuleAddresses) {
		t.Fatalf("Output 'nlb_forwarding_rule_addresses' is not valid JSON: %s", nlbForwardingRuleAddresses)
	}
	nlbForwardingRulesOutput := terraform.OutputJson(t, terraformOptions, "nlb_forwarding_rules") // For self_links
	nlbBackendServicesOutput := terraform.OutputJson(t, terraformOptions, "nlb_backend_services") // For self_links

	// Prepare for verification
	nlbNameToYaml := map[string]string{
		fmt.Sprintf("lite-%s", nlbInstanceName):     minimalNLBYamlFile,
		fmt.Sprintf("expanded-%s", nlbInstanceName): maximalNLBYamlFile,
	}

	loadBalancersToTest := gjson.Parse(nlbForwardingRuleAddresses).Map()
	if len(loadBalancersToTest) == 0 {
		t.Logf("No load balancers found in the output 'nlb_forwarding_rule_addresses'. Raw output: %s", nlbForwardingRuleAddresses)
	}

	// Create Test VM for connectivity checks (after LBs are provisioned)
	createTestVM(t, nlbProjectID, nlbZone, nlbTestVmName, nlbNetworkName, nlbSubnetName)

	// Iterate through created LBs for verification and connectivity tests
	for lbNameFromOutput := range loadBalancersToTest {
		t.Logf("Processing Load Balancer from output: %s", lbNameFromOutput)

		yamlFileName, ok := nlbNameToYaml[lbNameFromOutput]
		if !ok {
			t.Errorf("No YAML mapping found for NLB instance name from output: %s. Available mappings: %v", lbNameFromOutput, nlbNameToYaml)
			continue
		}

		// Perform configuration verification using gcloud describe and YAML comparison
		verifyNetworkLoadBalancerConfiguration(t, lbNameFromOutput, yamlFileName, terraformOptions)

		// Connectivity Check
		lbFwdRuleIPs := gjson.Parse(nlbForwardingRuleAddresses).Get(lbNameFromOutput) // Map of rule_key -> IP_address
		if !lbFwdRuleIPs.Exists() {
			t.Errorf("Forwarding rule IP addresses not found for NLB %s in output 'nlb_forwarding_rule_addresses'.", lbNameFromOutput)
			continue
		}

		if strings.HasPrefix(lbNameFromOutput, "lite-") { // lite NLB
			ipAddress := lbFwdRuleIPs.Get("").String() // Default rule key for lite is often ""
			if ipAddress == "" {
				t.Errorf("IP address for lite NLB %s (rule key '') is empty in 'nlb_forwarding_rule_addresses'. Raw IPs for this LB: %s", lbNameFromOutput, lbFwdRuleIPs.Raw)
			} else {
				verifyConnectivityToNLB(t, nlbProjectID, nlbZone, nlbTestVmName, ipAddress, apachePort, "Instance.*responding")
			}
		} else if strings.HasPrefix(lbNameFromOutput, "expanded-") { // Expanded NLB
			t.Logf("--- Starting Debug Block for Expanded NLB Port 80 & 9000 Connectivity (%s) ---", lbNameFromOutput)
			t.Logf("DEBUG: Expanded NLB (%s) raw forwarding rule IP addresses from 'nlb_forwarding_rule_addresses': %s", lbNameFromOutput, lbFwdRuleIPs.Raw)

			// Debug: Describe associated resources for Expanded NLB
			actualFwdRuleMapForMaxNLB := gjson.Parse(nlbForwardingRulesOutput).Get(lbNameFromOutput)
			actualBackendServiceJSONForMaxNLB := gjson.Parse(nlbBackendServicesOutput).Get(lbNameFromOutput)

			if actualFwdRuleMapForMaxNLB.Exists() {
				ruleCustomPortFromTFOutput := actualFwdRuleMapForMaxNLB.Get("rule-custom-port") // Key from your YAML
				if ruleCustomPortFromTFOutput.Exists() && ruleCustomPortFromTFOutput.Get("self_link").Exists() {
					frSelfLink := ruleCustomPortFromTFOutput.Get("self_link").String()
					frPathParts := strings.Split(frSelfLink, "/")
					frName := frPathParts[len(frPathParts)-1]
					frRegion := frPathParts[len(frPathParts)-3]
					t.Logf("DEBUG: Describing Forwarding Rule '%s' (region %s) for Expanded NLB 'rule-custom-port':", frName, frRegion)
					cmdFrDescribe := shell.Command{
						Command: "gcloud",
						Args:    []string{"compute", "forwarding-rules", "describe", frName, "--region", frRegion, "--project", nlbProjectID, "--format=json"},
					}
					frDetailsJsonString, errFrDescribe := shell.RunCommandAndGetOutputE(t, cmdFrDescribe)
					if errFrDescribe != nil {
						t.Logf("DEBUG: ERROR describing Forwarding Rule '%s': %v. Output: %s", frName, errFrDescribe, frDetailsJsonString)
					} else {
						t.Logf("DEBUG: Forwarding Rule '%s' details: %s", frName, frDetailsJsonString)
					}
				} else {
					t.Logf("DEBUG: 'rule-custom-port' not found or no self_link in 'nlb_forwarding_rules' output for %s. Raw map: %s", lbNameFromOutput, actualFwdRuleMapForMaxNLB.Raw)
				}
			}

			if actualBackendServiceJSONForMaxNLB.Exists() && actualBackendServiceJSONForMaxNLB.Get("self_link").Exists() {
				bsSelfLink := actualBackendServiceJSONForMaxNLB.Get("self_link").String()
				bsPathParts := strings.Split(bsSelfLink, "/")
				bsName := bsPathParts[len(bsPathParts)-1]
				bsRegion := bsPathParts[len(bsPathParts)-3]
				t.Logf("DEBUG: Describing Backend Service '%s' (region %s) for Expanded NLB:", bsName, bsRegion)
				cmdBsDescribe := shell.Command{
					Command: "gcloud",
					Args:    []string{"compute", "backend-services", "describe", bsName, "--region", bsRegion, "--project", nlbProjectID, "--format=json"},
				}
				bsDetailsJson, errBsDescribe := shell.RunCommandAndGetOutputE(t, cmdBsDescribe)
				if errBsDescribe != nil {
					t.Logf("DEBUG: ERROR describing Backend Service '%s': %v. Output: %s", bsName, errBsDescribe, bsDetailsJson)
				} else {
					t.Logf("DEBUG: Backend Service '%s' details: %s", bsName, bsDetailsJson)
				}

				t.Logf("DEBUG: Getting health for Backend Service '%s' (region %s) of Expanded NLB:", bsName, bsRegion)
				cmdBsGetHealth := shell.Command{
					Command: "gcloud",
					Args:    []string{"compute", "backend-services", "get-health", bsName, "--region", bsRegion, "--project", nlbProjectID, "--format=json"},
				}
				bsHealthJson, errBsHealth := shell.RunCommandAndGetOutputE(t, cmdBsGetHealth)
				if errBsHealth != nil {
					t.Logf("DEBUG: ERROR getting health for Backend Service '%s': %v. Output: %s", bsName, errBsHealth, bsHealthJson)
				} else {
					t.Logf("DEBUG: Backend Service '%s' health: %s", bsName, bsHealthJson)
				}
			}

			t.Logf("DEBUG: Describing MIG '%s' (region %s) associated with the NLBs:", nlbMigName, nlbRegion)
			cmdMigDescribe := shell.Command{
				Command: "gcloud",
				Args:    []string{"compute", "instance-groups", "managed", "describe", nlbMigName, "--region", nlbRegion, "--project", nlbProjectID, "--format=json"},
			}
			migDetailsJson, errMigDescribe := shell.RunCommandAndGetOutputE(t, cmdMigDescribe)
			if errMigDescribe != nil {
				t.Logf("DEBUG: ERROR describing MIG '%s': %v. Output: %s", nlbMigName, errMigDescribe, migDetailsJson)
			} else {
				t.Logf("DEBUG: MIG '%s' details: %s", nlbMigName, migDetailsJson)
				t.Logf("DEBUG: Listing instances in MIG '%s':", nlbMigName)
				cmdMigListInstances := shell.Command{
					Command: "gcloud",
					Args:    []string{"compute", "instance-groups", "managed", "list-instances", nlbMigName, "--region", nlbRegion, "--project", nlbProjectID, "--format=json"},
				}
				migInstancesJson, errMigList := shell.RunCommandAndGetOutputE(t, cmdMigListInstances)
				if errMigList != nil {
					t.Logf("DEBUG: ERROR listing instances in MIG '%s': %v. Output: %s", nlbMigName, errMigList, migInstancesJson)
				} else {
					t.Logf("DEBUG: MIG '%s' instances: %s", nlbMigName, migInstancesJson)
				}
			}
			t.Logf("--- End Debug Block for Expanded NLB (%s) ---", lbNameFromOutput)

			// Connectivity check for rule-http (port 80)
			ipAddressRuleHttp := lbFwdRuleIPs.Get("rule-http").String()
			if ipAddressRuleHttp == "" {
				t.Errorf("IP address for Expanded NLB %s (rule 'rule-http') is empty in 'nlb_forwarding_rule_addresses'.", lbNameFromOutput)
			} else {
				verifyConnectivityToNLB(t, nlbProjectID, nlbZone, nlbTestVmName, ipAddressRuleHttp, apachePort, "Instance.*responding")
			}

			// Connectivity check for rule-custom-port (port 9000)
			ipAddressRuleCustom := lbFwdRuleIPs.Get("rule-custom-port").String()
			if ipAddressRuleCustom == "" {
				t.Errorf("IP address for Expanded NLB %s (rule 'rule-custom-port') is empty in 'nlb_forwarding_rule_addresses'.", lbNameFromOutput)
			} else {
				verifyConnectivityToNLB(t, nlbProjectID, nlbZone, nlbTestVmName, ipAddressRuleCustom, "9000", "Instance.*responding")
			}
		}
	}
}

func createNetworkLoadBalancerYAML(t *testing.T) {
	t.Log("========= Generating YAML Files for Network Load Balancers =========")
	if err := os.MkdirAll(nlbConfigFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create NLB config directory %s: %v", nlbConfigFolderPath, err)
	}

	// Lite NLB Configuration - Relying on defaults for BackendService
	minNLBName := fmt.Sprintf("lite-%s", nlbInstanceName)
	minimalNLBCfg := NetworkLoadBalancerConfig{
		Name:      minNLBName,
		ProjectID: nlbProjectID,
		Region:    nlbRegion,
		Backends: []BackendItemConfig{
			{
				GroupName: nlbMigName,
			},
		},
		// Health check and Forwarding rule will also use module defaults if not specified
	}
	yamlMinimalData, err := yaml.Marshal(&minimalNLBCfg)
	if err != nil {
		t.Fatalf("Error marshaling lite NLB config: %v", err)
	}
	minimalFilePath := filepath.Join(nlbConfigFolderPath, minimalNLBYamlFile)
	if err := os.WriteFile(minimalFilePath, yamlMinimalData, 0644); err != nil {
		t.Fatalf("Unable to write lite NLB config to %s: %v", minimalFilePath, err)
	}
	t.Logf("Created Lite NLB YAML config at %s:\n%s", minimalFilePath, string(yamlMinimalData))

	// Expanded NLB Configuration
	maxNLBName := fmt.Sprintf("expanded-%s", nlbInstanceName)
	hcTimeout := 5
	hcCheckInterval := 10
	hcHealthyThreshold := 2
	hcUnhealthyThreshold := 3
	hcEnableLogging := true
	hcTCPPort := 80 // Health check will target port 80 on instances

	maximalNLBCfg := NetworkLoadBalancerConfig{
		Name:        maxNLBName,
		ProjectID:   nlbProjectID,
		Region:      nlbRegion,
		Description: "Main production backend group for expanded NLB",
		Backends: []BackendItemConfig{
			{
				GroupName:   nlbMigName,
				GroupRegion: nlbRegion,
				Description: "Main production backend group for expanded NLB",
			},
		},
		HealthCheck: &NetworkHealthCheckConfig{
			Description:        "Custom TCP Health Check for NLB",
			CheckIntervalSec:   &hcCheckInterval,
			TimeoutSec:         &hcTimeout,
			HealthyThreshold:   &hcHealthyThreshold,
			UnhealthyThreshold: &hcUnhealthyThreshold,
			EnableLogging:      &hcEnableLogging,
			TCP: &TCPHealthCheck{
				Port:              &hcTCPPort,
				PortSpecification: "USE_FIXED_PORT",
			},
		},
		ForwardingRules: map[string]ForwardingRuleConfig{
			"rule-http": {
				Protocol:    "TCP",
				Ports:       []string{apachePort},
				Description: "Forwarding rule for HTTP traffic",
			},
			"rule-custom-port": {
				Protocol:    "TCP",
				Ports:       []string{"9000"},
				Description: "Forwarding rule for custom port 9000 traffic",
			},
		},
	}
	yamlMaximalData, err := yaml.Marshal(&maximalNLBCfg)
	if err != nil {
		t.Fatalf("Error marshaling expanded NLB config: %v", err)
	}
	maximalFilePath := filepath.Join(nlbConfigFolderPath, maximalNLBYamlFile)
	if err := os.WriteFile(maximalFilePath, yamlMaximalData, 0644); err != nil {
		t.Fatalf("Unable to write expanded NLB config to %s: %v", maximalFilePath, err)
	}
	t.Logf("Created Expanded NLB YAML config at %s:\n%s", maximalFilePath, string(yamlMaximalData))
}

func verifyNetworkLoadBalancerConfiguration(t *testing.T, lbNameFromOutput string, yamlFileName string, terraformOptions *terraform.Options) {
	t.Logf("Verifying NLB configuration for: %s using YAML: %s", lbNameFromOutput, yamlFileName)

	yamlFilePath := filepath.Join(nlbConfigFolderPath, yamlFileName)
	yamlFileContent, err := os.ReadFile(yamlFilePath)
	if err != nil {
		t.Errorf("Error reading YAML file %s for NLB %s: %v", yamlFilePath, lbNameFromOutput, err)
		return // Cannot proceed without the expected config
	}
	var expectedConfig NetworkLoadBalancerConfig
	if err := yaml.Unmarshal(yamlFileContent, &expectedConfig); err != nil {
		t.Errorf("Error unmarshaling YAML for NLB %s from %s: %v", lbNameFromOutput, yamlFileName, err)
		return // Cannot proceed without parsed expected config
	}

	// Fetch current state from Terraform outputs
	nlbBackendServicesOutput := terraform.OutputJson(t, terraformOptions, "nlb_backend_services")
	nlbHealthChecksOutput := terraform.OutputJson(t, terraformOptions, "nlb_health_checks")
	nlbForwardingRulesOutput := terraform.OutputJson(t, terraformOptions, "nlb_forwarding_rules")

	// --- 1. Verify Backend Service ---
	actualBackendServiceJSON := gjson.Parse(nlbBackendServicesOutput).Get(lbNameFromOutput)
	if !actualBackendServiceJSON.Exists() {
		t.Errorf("Backend service for NLB %s not found in Terraform output 'nlb_backend_services'. Raw Output: %s", lbNameFromOutput, nlbBackendServicesOutput)
		// Attempt to continue verifying other parts if possible
	} else {
		bsSelfLink := actualBackendServiceJSON.Get("self_link").String()
		if bsSelfLink == "" {
			t.Errorf("Backend service self_link is empty for NLB %s in Terraform output.", lbNameFromOutput)
		} else {
			t.Logf("Verifying Backend Service: %s for NLB %s", bsSelfLink, lbNameFromOutput)
			bsPathParts := strings.Split(bsSelfLink, "/")
			bsName := bsPathParts[len(bsPathParts)-1]
			bsRegion := bsPathParts[len(bsPathParts)-3] // Assuming regional backend service

			cmdBsDescribe := shell.Command{
				Command: "gcloud",
				Args:    []string{"compute", "backend-services", "describe", bsName, "--region", bsRegion, "--project", nlbProjectID, "--format=json"},
			}
			bsDetailsJsonString, errBs := shell.RunCommandAndGetOutputE(t, cmdBsDescribe)
			if errBs != nil {
				t.Errorf("Failed to describe backend service %s in region %s: %v. Output: %s", bsName, bsRegion, errBs, bsDetailsJsonString)
			} else {
				bsDetailsJson := gjson.Parse(bsDetailsJsonString)

				if len(expectedConfig.Backends) == 0 {
					t.Errorf("NLB %s: No backends defined in expectedConfig (YAML) to verify against.", lbNameFromOutput)
				} else {
					expectedBackend := expectedConfig.Backends[0]

					// Compare attributes defined in the first expectedConfig.Backends item
					if expectedBackend.Protocol != "" && bsDetailsJson.Get("protocol").String() != expectedBackend.Protocol {
						t.Errorf("NLB %s: Backend Service protocol mismatch. YAML Expected: %s, Actual: %s", lbNameFromOutput, expectedBackend.Protocol, bsDetailsJson.Get("protocol").String())
					}
					if expectedBackend.PortName != "" && bsDetailsJson.Get("portName").String() != expectedBackend.PortName {
						t.Errorf("NLB %s: Backend Service portName mismatch. YAML Expected: %s, Actual: %s", lbNameFromOutput, expectedBackend.PortName, bsDetailsJson.Get("portName").String())
					}
					if expectedBackend.TimeoutSec != nil && bsDetailsJson.Get("timeoutSec").Int() != int64(*expectedBackend.TimeoutSec) {
						t.Errorf("NLB %s: Backend Service timeoutSec mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedBackend.TimeoutSec, bsDetailsJson.Get("timeoutSec").Int())
					}
					if expectedBackend.ConnectionDrainingTimeoutSec != nil && bsDetailsJson.Get("connectionDraining.drainingTimeoutSec").Int() != int64(*expectedBackend.ConnectionDrainingTimeoutSec) {
						t.Errorf("NLB %s: Backend Service connectionDrainingTimeoutSec mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedBackend.ConnectionDrainingTimeoutSec, bsDetailsJson.Get("connectionDraining.drainingTimeoutSec").Int())
					}
					if expectedBackend.SessionAffinity != "" && bsDetailsJson.Get("sessionAffinity").String() != expectedBackend.SessionAffinity {
						t.Errorf("NLB %s: Backend Service sessionAffinity mismatch. YAML Expected: %s, Actual: %s", lbNameFromOutput, expectedBackend.SessionAffinity, bsDetailsJson.Get("sessionAffinity").String())
					}
					// Compare logConfig only if it's defined in YAML
					if expectedBackend.LogSampleRate != nil {
						logConfig := bsDetailsJson.Get("logConfig")
						if !logConfig.Exists() || !logConfig.Get("enable").Bool() {
							t.Errorf("NLB %s: Backend Service logConfig expected to be enabled (due to sample rate) but is not. Actual: %s", lbNameFromOutput, logConfig.Raw)
						} else if logConfig.Get("sampleRate").Float() != *expectedBackend.LogSampleRate {
							t.Errorf("NLB %s: Backend Service logConfig sampleRate mismatch. YAML Expected: %f, Actual: %f", lbNameFromOutput, *expectedBackend.LogSampleRate, logConfig.Get("sampleRate").Float())
						}
					}
					if expectedBackend.Description != "" && bsDetailsJson.Get("description").String() != expectedBackend.Description {
						t.Errorf("NLB %s: Backend Service description mismatch (comparing backend item desc). YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, expectedBackend.Description, bsDetailsJson.Get("description").String())
					} else if expectedConfig.Description != "" && bsDetailsJson.Get("description").String() != expectedConfig.Description && expectedBackend.Description == "" {
						t.Logf("NLB %s: Comparing overall LB description from YAML with Backend Service description. YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, expectedConfig.Description, bsDetailsJson.Get("description").String())
					}

					// Verify the correct MIG is attached
					migFoundInBackend := false
					expectedMigRegion := nlbRegion // Default to LB region
					if expectedBackend.GroupRegion != "" {
						expectedMigRegion = expectedBackend.GroupRegion
					}
					// Use GroupName from the specific backend item in YAML
					expectedMigLink := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/instanceGroups/%s", nlbProjectID, expectedMigRegion, expectedBackend.GroupName)

					for _, backendEntry := range bsDetailsJson.Get("backends").Array() {
						if backendEntry.Get("group").String() == expectedMigLink {
							migFoundInBackend = true
							break
						}
					}
					if !migFoundInBackend {
						t.Errorf("NLB %s: Expected MIG %s (from YAML backend.group_name '%s') not found in backend service %s backends. Found: %s", lbNameFromOutput, expectedMigLink, expectedBackend.GroupName, bsName, bsDetailsJson.Get("backends").Raw)
					}
				}
			}
		}

		// --- 2. Verify Health Check ---
		if expectedConfig.HealthCheck != nil { // Only verify if HC is defined in the YAML
			hcFromOutput := gjson.Parse(nlbHealthChecksOutput).Get(lbNameFromOutput) // Assumes HC name in output matches LB name (module convention)
			hcSelfLink := ""
			hcNameInYaml := expectedConfig.HealthCheck.Name // Check if an existing HC was specified by name

			if hcNameInYaml != "" {
				t.Logf("NLB %s: Verifying association with pre-existing health check defined in YAML: %s.", lbNameFromOutput, hcNameInYaml)
				// Find the self-link of the named HC from the Backend Service details
				hcAssociated := false
				if actualBackendServiceJSON.Exists() { // Check if we successfully got BS details earlier
					bsDetailsJsonString, _ := shell.RunCommandAndGetOutputE(t, shell.Command{Command: "gcloud", Args: []string{"compute", "backend-services", "describe", gjson.Parse(actualBackendServiceJSON.Get("self_link").String()).Get("name").String(), "--region", gjson.Parse(actualBackendServiceJSON.Get("self_link").String()).Get("region").String(), "--project", nlbProjectID, "--format=json"}}) // Re-fetch if needed, handle error
					bsHealthChecks := gjson.Parse(bsDetailsJsonString).Get("healthChecks").Array()
					for _, bsHcLink := range bsHealthChecks {
						if strings.HasSuffix(bsHcLink.String(), "/"+hcNameInYaml) {
							hcAssociated = true
							hcSelfLink = bsHcLink.String() // Found the full self-link
							break
						}
					}
				}
				if !hcAssociated {
					t.Errorf("NLB %s: Expected existing health check %s (from YAML) not found associated with the backend service.", lbNameFromOutput, hcNameInYaml)
					// Cannot verify HC details if we couldn't confirm association/find link
				}
			} else if hcFromOutput.Exists() { // Auto-created HC (name not specified in YAML), get link from TF output
				hcSelfLink = hcFromOutput.Get("self_link").String()
				if hcSelfLink == "" {
					t.Errorf("NLB %s: Health Check self_link is empty in Terraform output 'nlb_health_checks'.", lbNameFromOutput)
				}
			} else {
				// Fallback: No HC name in YAML, not found in TF output by LB name key.
				// Check if the BS has *any* HC associated (module might create one with generated name)
				if actualBackendServiceJSON.Exists() {
					bsDetailsJsonString, _ := shell.RunCommandAndGetOutputE(t, shell.Command{Command: "gcloud", Args: []string{"compute", "backend-services", "describe", gjson.Parse(actualBackendServiceJSON.Get("self_link").String()).Get("name").String(), "--region", gjson.Parse(actualBackendServiceJSON.Get("self_link").String()).Get("region").String(), "--project", nlbProjectID, "--format=json"}}) // Re-fetch if needed, handle error
					bsHealthChecks := gjson.Parse(bsDetailsJsonString).Get("healthChecks").Array()
					if len(bsHealthChecks) > 0 {
						hcSelfLink = bsHealthChecks[0].String()
						t.Logf("NLB %s: Health Check name not in YAML/output key, verifying HC found on Backend Service: %s", lbNameFromOutput, hcSelfLink)
					} else {
						t.Errorf("NLB %s: No health check specified in YAML, none found in Terraform output by key '%s', and none associated with the backend service.", lbNameFromOutput, lbNameFromOutput)
					}
				} else {
					t.Errorf("NLB %s: Cannot determine health check - YAML is missing HC block, and Backend Service details were not available from TF output.", lbNameFromOutput)
				}
			}

			// Describe and compare the HC if we have a self-link
			if hcSelfLink != "" {
				t.Logf("Verifying Health Check details for: %s (associated with NLB %s)", hcSelfLink, lbNameFromOutput)
				hcPathParts := strings.Split(hcSelfLink, "/")
				hcName := hcPathParts[len(hcPathParts)-1]

				// Determine command args based on scope from self-link
				hcDescribeArgs := []string{"compute"}
				var hcRegion, hcScope string
				isRegional := false
				isGlobal := false

				// Find scope and region/global identifier
				for i, part := range hcPathParts {
					if (part == "regions" || part == "global") && i+1 < len(hcPathParts) {
						hcScope = part
						if hcScope == "regions" {
							hcRegion = hcPathParts[i+1]
							isRegional = true
						} else {
							isGlobal = true
						}
						break
					}
				}

				if isRegional {
					hcDescribeArgs = append(hcDescribeArgs, "health-checks", "describe", hcName, "--region", hcRegion)
				} else if isGlobal {
					hcDescribeArgs = append(hcDescribeArgs, "health-checks", "describe", hcName, "--global")
				} else {
					t.Errorf("NLB %s: Could not determine scope (regional/global) for health check from self_link: %s", lbNameFromOutput, hcSelfLink)
					hcDescribeArgs = nil // Prevent command execution
				}

				if hcDescribeArgs != nil {
					hcDescribeArgs = append(hcDescribeArgs, "--project", nlbProjectID, "--format=json")
					cmdHcDescribe := shell.Command{
						Command: "gcloud",
						Args:    hcDescribeArgs,
					}
					hcDetailsJsonString, errHc := shell.RunCommandAndGetOutputE(t, cmdHcDescribe)
					if errHc != nil {
						t.Errorf("Failed to describe health check %s (Args: %v): %v. Output: %s", hcName, cmdHcDescribe.Args, errHc, hcDetailsJsonString)
					} else {
						hcDetailsJson := gjson.Parse(hcDetailsJsonString)

						// --- Compare Health Check Attributes ---
						// Compare Type (only if specified in YAML, otherwise module default)
						expectedType := ""
						if expectedConfig.HealthCheck.TCP != nil {
							expectedType = "TCP"
						}
						if expectedConfig.HealthCheck.SSL != nil {
							expectedType = "SSL"
						}
						if expectedConfig.HealthCheck.HTTP != nil {
							expectedType = "HTTP"
						}
						if expectedConfig.HealthCheck.HTTPS != nil {
							expectedType = "HTTPS"
						}
						if expectedConfig.HealthCheck.HTTP2 != nil {
							expectedType = "HTTP2"
						}
						if expectedConfig.HealthCheck.GRPC != nil {
							expectedType = "GRPC"
						}

						if expectedType != "" && hcDetailsJson.Get("type").String() != expectedType {
							t.Errorf("NLB %s: HC type mismatch. YAML implies %s, Actual: %s", lbNameFromOutput, expectedType, hcDetailsJson.Get("type").String())
						}

						// Compare common attributes (only if set in YAML expectedConfig)
						if expectedConfig.HealthCheck.CheckIntervalSec != nil && hcDetailsJson.Get("checkIntervalSec").Int() != int64(*expectedConfig.HealthCheck.CheckIntervalSec) {
							t.Errorf("NLB %s: HC checkIntervalSec mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedConfig.HealthCheck.CheckIntervalSec, hcDetailsJson.Get("checkIntervalSec").Int())
						}
						if expectedConfig.HealthCheck.TimeoutSec != nil && hcDetailsJson.Get("timeoutSec").Int() != int64(*expectedConfig.HealthCheck.TimeoutSec) {
							t.Errorf("NLB %s: HC timeoutSec mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedConfig.HealthCheck.TimeoutSec, hcDetailsJson.Get("timeoutSec").Int())
						}
						if expectedConfig.HealthCheck.HealthyThreshold != nil && hcDetailsJson.Get("healthyThreshold").Int() != int64(*expectedConfig.HealthCheck.HealthyThreshold) {
							t.Errorf("NLB %s: HC healthyThreshold mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedConfig.HealthCheck.HealthyThreshold, hcDetailsJson.Get("healthyThreshold").Int())
						}
						if expectedConfig.HealthCheck.UnhealthyThreshold != nil && hcDetailsJson.Get("unhealthyThreshold").Int() != int64(*expectedConfig.HealthCheck.UnhealthyThreshold) {
							t.Errorf("NLB %s: HC unhealthyThreshold mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedConfig.HealthCheck.UnhealthyThreshold, hcDetailsJson.Get("unhealthyThreshold").Int())
						}
						if expectedConfig.HealthCheck.Description != "" && hcDetailsJson.Get("description").String() != expectedConfig.HealthCheck.Description {
							t.Errorf("NLB %s: HC description mismatch. YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, expectedConfig.HealthCheck.Description, hcDetailsJson.Get("description").String())
						}
						// Compare logging (handle case where logConfig block might not exist if disabled)
						if expectedConfig.HealthCheck.EnableLogging != nil {
							actualLoggingEnabled := hcDetailsJson.Get("logConfig.enable").Bool() // Defaults to false if path doesn't exist
							if actualLoggingEnabled != *expectedConfig.HealthCheck.EnableLogging {
								t.Errorf("NLB %s: HC logConfig.enable mismatch. YAML Expected: %t, Actual: %t", lbNameFromOutput, *expectedConfig.HealthCheck.EnableLogging, actualLoggingEnabled)
							}
						}

						// Compare type-specific details (e.g., TCP)
						if expectedConfig.HealthCheck.TCP != nil {
							tcpHcPath := "tcpHealthCheck" // Check gcloud output format if this changes
							if !hcDetailsJson.Get(tcpHcPath).Exists() {
								t.Errorf("NLB %s: HC type is TCP but details block '%s' is missing. Actual Details: %s", lbNameFromOutput, tcpHcPath, hcDetailsJsonString)
							} else {
								if expectedConfig.HealthCheck.TCP.Port != nil && hcDetailsJson.Get(tcpHcPath+".port").Int() != int64(*expectedConfig.HealthCheck.TCP.Port) {
									t.Errorf("NLB %s: HC TCP port mismatch. YAML Expected: %d, Actual: %d", lbNameFromOutput, *expectedConfig.HealthCheck.TCP.Port, hcDetailsJson.Get(tcpHcPath+".port").Int())
								}
								if expectedConfig.HealthCheck.TCP.PortSpecification != "" && hcDetailsJson.Get(tcpHcPath+".portSpecification").String() != expectedConfig.HealthCheck.TCP.PortSpecification {
									t.Errorf("NLB %s: HC TCP PortSpecification mismatch. YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, expectedConfig.HealthCheck.TCP.PortSpecification, hcDetailsJson.Get(tcpHcPath+".portSpecification").String())
								}
								// Add checks for Request, Response, ProxyHeader if defined in YAML TCP config
							}
						}
						// Add similar blocks for SSL, HTTP, etc., if needed
					}
				}
			}
		} // End Health Check Verification

		// --- 3. Verify Forwarding Rules ---
		actualFwdRulesMapJSON := gjson.Parse(nlbForwardingRulesOutput).Get(lbNameFromOutput) // Map of rule_key -> rule_details from TF output
		if !actualFwdRulesMapJSON.Exists() {
			t.Errorf("Forwarding rules map for NLB %s not found in Terraform output 'nlb_forwarding_rules'. Raw Output: %s", lbNameFromOutput, nlbForwardingRulesOutput)
		} else {
			expectedFwdRules := expectedConfig.ForwardingRules // Rules defined in the YAML

			// Handle lite config case (no forwarding_rules block in YAML -> module creates one default rule)
			if len(expectedFwdRules) == 0 {
				outputRuleKeys := []string{}
				actualFwdRulesMapJSON.ForEach(func(key, value gjson.Result) bool {
					outputRuleKeys = append(outputRuleKeys, key.String())
					return true
				})

				if len(outputRuleKeys) == 1 {
					t.Logf("NLB %s: Lite config (no rules in YAML), verifying single output rule with key '%s'", lbNameFromOutput, outputRuleKeys[0])
					// Synthesize an expected rule entry using the output key, mainly to check its existence via gcloud
					expectedFwdRules = map[string]ForwardingRuleConfig{
						outputRuleKeys[0]: {}, // We'll verify basic properties of this rule below
					}
				} else if len(outputRuleKeys) > 1 {
					t.Errorf("NLB %s: Lite config expected (no rules in YAML), but multiple rules found in TF output: %v", lbNameFromOutput, outputRuleKeys)
				} else {
					// This case should ideally not happen if TF apply succeeded and created an NLB
					t.Errorf("NLB %s: Lite config expected, but NO forwarding rules found in TF output.", lbNameFromOutput)
				}
			}

			// Iterate through the rules EXPECTED based on YAML (or the synthesized single rule for lite)
			for ruleKeyFromYaml, expectedRuleConfig := range expectedFwdRules {
				actualRuleDetailsFromTF := actualFwdRulesMapJSON.Get(ruleKeyFromYaml)

				if !actualRuleDetailsFromTF.Exists() {
					t.Errorf("NLB %s: Forwarding rule with key '%s' (from YAML/default) not found in TF output map. Available output keys: %s", lbNameFromOutput, ruleKeyFromYaml, actualFwdRulesMapJSON.Raw)
					continue // Skip verification for this rule
				}

				frSelfLink := actualRuleDetailsFromTF.Get("self_link").String()
				if frSelfLink == "" {
					t.Errorf("NLB %s, Rule Key '%s': self_link is empty in Terraform output.", lbNameFromOutput, ruleKeyFromYaml)
					continue
				}
				t.Logf("Verifying Forwarding Rule details (key %s): %s for NLB %s", ruleKeyFromYaml, frSelfLink, lbNameFromOutput)
				frPathParts := strings.Split(frSelfLink, "/")
				frName := frPathParts[len(frPathParts)-1]
				frRegion := frPathParts[len(frPathParts)-3] // Assuming regional forwarding rule

				cmdFrDescribe := shell.Command{
					Command: "gcloud",
					Args:    []string{"compute", "forwarding-rules", "describe", frName, "--region", frRegion, "--project", nlbProjectID, "--format=json"},
				}
				frDetailsJsonString, errFr := shell.RunCommandAndGetOutputE(t, cmdFrDescribe)
				if errFr != nil {
					t.Errorf("Failed to describe forwarding rule %s in region %s: %v. Output: %s", frName, frRegion, errFr, frDetailsJsonString)
					continue
				}
				frDetailsJson := gjson.Parse(frDetailsJsonString)

				// --- Compare Forwarding Rule Attributes ---
				if frDetailsJson.Get("IPAddress").String() == "" {
					t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Actual IPAddress from gcloud is empty.", lbNameFromOutput, ruleKeyFromYaml, frName)
				}
				if expectedRuleConfig.Protocol != "" && frDetailsJson.Get("IPProtocol").String() != expectedRuleConfig.Protocol {
					t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Protocol mismatch. YAML Expected: %s, Actual: %s", lbNameFromOutput, ruleKeyFromYaml, frName, expectedRuleConfig.Protocol, frDetailsJson.Get("IPProtocol").String())
				}

				// Compare Ports (handling 'ports' array, 'portRange', and 'allPorts')
				yamlPorts := expectedRuleConfig.Ports
				gcloudPortsJson := frDetailsJson.Get("ports")              // Array like ["80", "443"]
				gcloudPortRange := frDetailsJson.Get("portRange").String() // String like "8000-8008" or "80"
				gcloudAllPorts := frDetailsJson.Get("allPorts").Bool()     // Boolean

				if len(yamlPorts) > 0 { // YAML specified specific ports
					if gcloudAllPorts {
						t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port mismatch. YAML Expected specific ports %v, but Actual FR has 'allPorts' enabled.", lbNameFromOutput, ruleKeyFromYaml, frName, yamlPorts)
					} else if gcloudPortsJson.Exists() && gcloudPortsJson.IsArray() {
						// Compare YAML ports set with gcloud ports set
						actualPortsSet := make(map[string]struct{})
						for _, p := range gcloudPortsJson.Array() {
							actualPortsSet[p.String()] = struct{}{}
						}
						yamlPortsSet := make(map[string]struct{})
						for _, yp := range yamlPorts {
							yamlPortsSet[yp] = struct{}{}
						}

						if len(actualPortsSet) != len(yamlPortsSet) {
							t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port count mismatch. YAML Expected: %d (%v), Actual: %d (%v)", lbNameFromOutput, ruleKeyFromYaml, frName, len(yamlPortsSet), yamlPorts, len(actualPortsSet), gcloudPortsJson.Array())
						} else {
							for yp := range yamlPortsSet {
								if _, ok := actualPortsSet[yp]; !ok {
									t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port mismatch. YAML Expected port '%s' not found in Actual ports %v", lbNameFromOutput, ruleKeyFromYaml, frName, yp, gcloudPortsJson.Array())
								}
							}
						}
					} else if gcloudPortRange != "" {
						// Handle portRange case - less common for NLB, usually for single ports or ranges
						if len(yamlPorts) == 1 && yamlPorts[0] == gcloudPortRange {
							// Match
						} else {
							t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port mismatch. YAML Expected ports %v, but Actual FR uses 'portRange': '%s'", lbNameFromOutput, ruleKeyFromYaml, frName, yamlPorts, gcloudPortRange)
						}
					} else {
						// No 'allPorts', no 'ports' array, no 'portRange' - Actual FR has no port specification
						t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port mismatch. YAML Expected specific ports %v, but Actual FR has no port specification.", lbNameFromOutput, ruleKeyFromYaml, frName, yamlPorts)
					}
				} else { // YAML did not specify ports (e.g., lite config implies all ports)
					if !gcloudAllPorts {
						// If YAML implies all ports, 'allPorts' should be true on the actual FR
						t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Port mismatch. YAML expected all ports (no specific ports defined), but Actual FR has 'allPorts' set to false. Ports: '%s', Range: '%s'", lbNameFromOutput, ruleKeyFromYaml, frName, gcloudPortsJson.Raw, gcloudPortRange)
					}
				}

				// Compare other FR attributes if specified in YAML
				if expectedRuleConfig.Description != "" && frDetailsJson.Get("description").String() != expectedRuleConfig.Description {
					t.Errorf("NLB %s, Rule Key '%s' (Name: %s): Description mismatch. YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, ruleKeyFromYaml, frName, expectedRuleConfig.Description, frDetailsJson.Get("description").String())
				}
				if expectedRuleConfig.Address != "" && frDetailsJson.Get("IPAddress").String() != expectedRuleConfig.Address {
					t.Errorf("NLB %s, Rule Key '%s' (Name: %s): IPAddress mismatch. YAML Expected: '%s', Actual: '%s'", lbNameFromOutput, ruleKeyFromYaml, frName, expectedRuleConfig.Address, frDetailsJson.Get("IPAddress").String())
				}
				// Add checks for Subnetwork, IP version etc. if needed
			}
		} // End Forwarding Rule Verification

		t.Logf("Finished verifying NLB configuration for: %s", lbNameFromOutput)
	}
}

// Helper to create an instance template
func createInstanceTemplate(t *testing.T, templateName, projectID, networkName, subnetName, region string, tags []string) {
	fullStartupScript := `#!/bin/bash
	# Install necessary tools
	apt-get update -y
	apt-get install -y apache2 netcat-openbsd
	
	# Configure Apache for Port 80
	echo "Instance $(hostname) responding on port 80" > /var/www/html/index.html
	systemctl restart apache2
	
	# Configure a simple listener for Port 9000 (using netcat)
	nohup bash -c '
	while true; do
	  CURRENT_HOSTNAME=$(hostname)
	  MESSAGE_BODY="Instance $CURRENT_HOSTNAME responding on port 9000"
	  CONTENT_LENGTH=$(echo -n "$MESSAGE_BODY" | wc -c)
	  # Use printf for more reliable header construction with CRLF and ensure variables are evaluated in this subshell
	  printf "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %s\r\nConnection: close\r\n\r\n%s" "$CONTENT_LENGTH" "$MESSAGE_BODY" | nc -lp 9000
	done
	' >/dev/null 2>&1 &
	
	echo "Startup script finished."
	` // Use direct assignment with backticks for the multi-line script

	args := []string{
		"compute", "instance-templates", "create", templateName,
		"--project=" + projectID,
		"--machine-type=e2-small",
		"--image-family=debian-11",
		"--image-project=debian-cloud",
		"--network=" + networkName,
		"--subnet=" + subnetName,
		"--region=" + region,
		"--tags=" + strings.Join(tags, ","),
		"--metadata=startup-script=" + fullStartupScript, // Use the enhanced script
	}

	cmd := shell.Command{Command: "gcloud", Args: args}
	commandString := fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " "))

	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Failed to create Instance Template %s: %v. Command: [%s]", templateName, err, commandString)
	} else {
		t.Logf("Successfully created Instance Template: %s", templateName)
	}
}

func setNamedPortsOnMIG(t *testing.T, projectID, region, migName, portName, portNumber string) {
	t.Logf("Setting named port %s:%s on MIG %s in region %s", portName, portNumber, migName, region)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instance-groups", "managed", "set-named-ports", migName,
			"--project=" + projectID,
			"--region=" + region,
			fmt.Sprintf("--named-ports=%s:%s", portName, portNumber),
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		// Use t.Fatalf if setting named ports is critical
		t.Fatalf("Failed to set named ports on MIG %s: %v", migName, err)
	}
	t.Logf("Successfully set named port %s:%s on MIG %s", portName, portNumber, migName)
}

// Helper to create MIG for NLB
func createManagedInstanceGroupNLB(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instance-groups", "managed", "create", nlbMigName,
			"--project=" + nlbProjectID,
			"--base-instance-name", fmt.Sprintf("%s-instance", strings.TrimSuffix(nlbMigName, "-nlb")),
			"--size", "2", // Start with 2 instances
			"--template", nlbTemplateName,
			"--region", nlbRegion,
			// Add health check if needed (though module can create backend service based HC)
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Failed to create Managed Instance Group %s: %v", nlbMigName, err)
	} else {
		t.Logf("Successfully created Managed Instance Group: %s", nlbMigName)
		// Wait for MIG to stabilize
		time.Sleep(120 * time.Second)
	}
}

// Helper to delete MIG for NLB
func deleteManagedInstanceGroupNLB(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instance-groups", "managed", "delete", nlbMigName,
			"--project=" + nlbProjectID,
			"--region=" + nlbRegion,
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		// Don't fail the test, just log, as other cleanup might still be needed
		t.Logf("Failed to delete Managed Instance Group %s: %v. This might be okay if it was already deleted or never created.", nlbMigName, err)
	} else {
		t.Logf("Successfully deleted Managed Instance Group: %s", nlbMigName)
	}
}

// Helper to delete instance template for NLB
func deleteInstanceTemplateNLB(t *testing.T) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instance-templates", "delete", nlbTemplateName,
			"--project=" + nlbProjectID,
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Failed to delete Instance Template %s: %v.", nlbTemplateName, err)
	} else {
		t.Logf("Successfully deleted Instance Template: %s", nlbTemplateName)
	}
}

// Firewall rule for NLB Health Checkers (Google's known IP ranges)
func createFirewallRuleForNLBHealthChecks(t *testing.T, projectID, networkName, ruleName string, targetTags []string) {
	// Google Cloud health checker IP ranges
	healthCheckIPRanges := []string{"130.211.0.0/22", "35.191.0.0/16", "209.85.152.0/22", "209.85.204.0/22"}
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "create", ruleName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--action=ALLOW",
			"--rules=tcp,udp", // Allow TCP and UDP for various health check types
			"--source-ranges=" + strings.Join(healthCheckIPRanges, ","),
			"--target-tags=" + strings.Join(targetTags, ","),
			"--description=Allow traffic from GCP health checkers for NLB",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Failed to create firewall rule %s for NLB health checks: %v", ruleName, err)
	} else {
		t.Logf("Successfully created firewall rule %s for NLB health checks.", ruleName)
	}
}

// Firewall rule for NLB Traffic (from anywhere to specific ports)
func createFirewallRuleForNLBTraffic(t *testing.T, projectID, networkName, ruleName string, ports []string, targetTags []string) {
	allowRules := []string{}
	for _, port := range ports {
		allowRules = append(allowRules, "tcp:"+port, "udp:"+port) // Assuming TCP and UDP for given ports
	}

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "create", ruleName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--action=ALLOW",
			"--rules=" + strings.Join(allowRules, ","),
			"--source-ranges=0.0.0.0/0", // Allow traffic from anywhere
			"--target-tags=" + strings.Join(targetTags, ","),
			"--description=Allow external traffic to NLB instances",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Failed to create firewall rule %s for NLB traffic: %v", ruleName, err)
	} else {
		t.Logf("Successfully created firewall rule %s for NLB traffic.", ruleName)
	}
}

// createTestVM: Creates a VM for connectivity testing
func createTestVM(t *testing.T, projectID, zone, vmName, networkName, subnetName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "create", vmName,
			"--project=" + projectID,
			"--zone=" + zone,
			"--machine-type=e2-micro",
			"--image-family=debian-11",
			"--image-project=debian-cloud",
			"--network=" + networkName,
			"--subnet=" + subnetName,
			"--scopes=cloud-platform", // For gcloud commands from within if needed
			"--tags=allow-iap-ssh",
			"--metadata=startup-script=apt-get update -y && apt-get install -y curl dnsutils netcat-openbsd",
		},
	}
	_, err := retry.DoWithRetryE(t, "Create Test VM", 2, 10*time.Second, func() (string, error) {
		return shell.RunCommandAndGetOutputE(t, cmd)
	})
	if err != nil {
		t.Fatalf("Failed to create test VM %s after retries: %v", vmName, err)
	}
	time.Sleep(60 * time.Second) // Give VM time to boot and run startup
}

// deleteTestVM: Deletes the test VM
func deleteTestVM(t *testing.T, projectID, zone, vmName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "delete", vmName,
			"--project=" + projectID,
			"--zone=" + zone,
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Failed to delete test VM %s: %v. This might be okay if it was already deleted.", vmName, err)
	} else {
		t.Logf("Successfully deleted test VM %s.", vmName)
	}
}

// verifyConnectivityToNLB: Uses gcloud to SSH into test VM and curl the LB IP
func verifyConnectivityToNLB(t *testing.T, projectID, zone, testVmName, lbIpAddress, port, expectedTextPattern string) {
	t.Logf("Verifying connectivity from VM %s to NLB IP %s on port %s, expecting to match: '%s'", testVmName, lbIpAddress, port, expectedTextPattern)

	maxRetries := 5
	sleepBetweenRetries := 20 * time.Second
	var sshCommand string

	// Compile the regex pattern once before the retry loop
	expectedRegex, regexCompileErr := regexp.Compile(expectedTextPattern)
	if regexCompileErr != nil {
		// If the pattern is invalid, fail fast, no point in retrying
		t.Fatalf("Invalid regex pattern '%s' provided for matching: %v", expectedTextPattern, regexCompileErr)
		return
	}

	// Determine the command based on whether a specific text pattern is expected (implies curl)
	// or just a successful connection (can use netcat).
	if expectedTextPattern == "" || expectedTextPattern == "Netcat success" {
		// Use netcat for simple TCP connect check
		sshCommand = fmt.Sprintf("nc -z -w 5 %s %s && echo 'Netcat success' || echo 'Netcat failed'", lbIpAddress, port)
		if expectedTextPattern == "" {
			expectedTextPattern = "Netcat success" // Explicitly expect this if none provided
		}
		t.Logf("Using netcat for connectivity check to %s:%s", lbIpAddress, port)
		// Recompile regex if needed for the netcat case (should already be compiled above)
		expectedRegex, _ = regexp.Compile(expectedTextPattern) // Assume pattern is valid
	} else {
		// Use curl with --fail for HTTP status check via exit code, plus -v for debug, -L follow redirects, -m timeout
		sshCommand = fmt.Sprintf("curl -v -L --fail -m 15 http://%s:%s", lbIpAddress, port)
		t.Logf("Using curl --fail for connectivity check to %s:%s", lbIpAddress, port)
	}

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("SSH and check NLB %s:%s", lbIpAddress, port), maxRetries, sleepBetweenRetries, func() (string, error) {
		gcloudArgs := []string{
			"compute", "ssh", testVmName,
			"--project=" + projectID,
			"--zone=" + zone,
			"--command=" + sshCommand,
			"--tunnel-through-iap", // Assume IAP is needed based on common setup
			"--quiet",              // Suppress gcloud informational messages
		}
		cmd := shell.Command{Command: "gcloud", Args: gcloudArgs}
		// Execute the command. runErr will be non-nil if gcloud ssh fails OR if the command inside (curl --fail / nc) fails.
		output, runErr := shell.RunCommandAndGetOutputE(t, cmd)

		// Log the raw output regardless of error for better debugging context on retries
		t.Logf("DEBUG: Attempting SSH command (%s) to %s:%s:\n%s", cmd.Command, lbIpAddress, port, output)

		// --- Primary Failure Check ---
		if runErr != nil {
			// Check for specific connection errors *within* the output for better context, even though runErr is set.
			// These might indicate network path issues before curl --fail could even evaluate HTTP status.
			connectionRefused := strings.Contains(output, "Connection refused") || strings.Contains(output, "connect to .* port .* failed: Connection refused")
			timeoutOccurred := strings.Contains(output, "Connection timed out") || strings.Contains(output, "Operation timed out")
			netcatFailed := strings.Contains(output, "Netcat failed") // Check if using nc

			failureReason := "gcloud ssh or internal command failed (check runErr)" // Default reason if runErr != nil
			if connectionRefused {
				failureReason = "connection refused"
			} else if timeoutOccurred {
				failureReason = "connection timed out"
			} else if netcatFailed {
				failureReason = "netcat failed"
			}
			// For curl --fail, runErr indicates either SSH issue or HTTP status >= 400

			// Return the specific reason and the original error from RunCommandAndGetOutputE
			return "", fmt.Errorf("%s checking NLB %s:%s via %s. Error: %v. Output logged above.", failureReason, lbIpAddress, port, testVmName, runErr)
		}

		// --- Success Check (if runErr is nil) ---
		// If we are here, gcloud ssh succeeded AND the internal command (curl --fail or nc) also had exit code 0.
		// For curl, this means HTTP status < 400. For nc, it means connection was successful.

		// Now, check if the output *content* matches the expected pattern.
		if !expectedRegex.MatchString(output) {
			// Command execution succeeded, but the response content is wrong.
			return "", fmt.Errorf("command succeeded but response from NLB %s:%s via %s did not match expected pattern '%s'. Output logged above.", lbIpAddress, port, testVmName, expectedTextPattern)
		}

		// Both command execution and content matching succeeded
		t.Logf("Successfully connected to NLB %s:%s from %s. Output matches expected pattern.", lbIpAddress, port, testVmName)
		return output, nil // Success
	})

	// Handle final error after retries
	if err != nil {
		// If all retries failed, log the final error and fail the test
		t.Errorf("Failed to verify connectivity to NLB %s:%s after %d attempts: %v", lbIpAddress, port, maxRetries, err)
		// Consider t.Fatalf if connectivity is absolutely critical for the test pass/fail
		t.Fail() // Mark the test as failed but allow other parts of the test function to potentially run cleanup
	}
}

// createVPC creates a Virtual Private Cloud (VPC) network and a subnet in Google Cloud
func createVPC(t *testing.T, projectID string, networkName string) {
	// Check if VPC already exists
	t.Logf("Attempting to describe VPC %s to check if it already exists in project %s...", networkName, projectID) // Added log
	cmdCheckVPC := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "describe", networkName, "--project=" + projectID, "--format=value(name)"},
	}
	vpcExistsOutput, errVPC := shell.RunCommandAndGetOutputE(t, cmdCheckVPC)
	// It's good practice to log the error from describe, even if we proceed, to see why it might have failed if not "not found"
	if errVPC != nil {
		t.Logf("Describing VPC %s failed (this is often expected if it doesn't exist): %v. Output: %s", networkName, errVPC, vpcExistsOutput)
	}

	if strings.TrimSpace(vpcExistsOutput) == networkName {
		t.Logf("VPC %s already exists, skipping creation.", networkName)
	} else {
		t.Logf("VPC %s not found or describe output mismatch, proceeding with creation.", networkName) // Added log
		cmdCreateVPC := shell.Command{
			Command: "gcloud",
			Args: []string{"compute", "networks", "create", networkName,
				"--project=" + projectID,
				"--format=json",
				"--bgp-routing-mode=global",
				"--subnet-mode=custom"},
		}
		if _, err := shell.RunCommandAndGetOutputE(t, cmdCreateVPC); err != nil {
			t.Fatalf("Error creating VPC %s: %v", networkName, err)
		}
		t.Logf("Successfully created VPC: %s", networkName)
	}

	time.Sleep(10 * time.Second) // allow network to be ready

	// Check if Subnet already exists
	currentSubnetName := fmt.Sprintf("%s-subnet", networkName)
	t.Logf("Attempting to describe Subnet %s in region %s (VPC %s) to check if it already exists...", currentSubnetName, nlbRegion, networkName) // Added log
	cmdCheckSubnet := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "subnets", "describe", currentSubnetName, "--project=" + projectID, "--region=" + nlbRegion, "--format=value(name)"},
	}
	subnetExistsOutput, errSubnet := shell.RunCommandAndGetOutputE(t, cmdCheckSubnet)
	if errSubnet != nil {
		t.Logf("Describing Subnet %s failed (this is often expected if it doesn't exist): %v. Output: %s", currentSubnetName, errSubnet, subnetExistsOutput)
	}

	if strings.TrimSpace(subnetExistsOutput) == currentSubnetName {
		t.Logf("Subnet %s in VPC %s already exists, skipping creation.", currentSubnetName, networkName)
	} else {
		t.Logf("Subnet %s not found or describe output mismatch, proceeding with creation.", currentSubnetName) // Added log
		cmdCreateSubnet := shell.Command{
			Command: "gcloud",
			Args: []string{"compute", "networks", "subnets", "create", currentSubnetName,
				"--project=" + projectID,
				"--network=" + networkName,
				"--region=" + nlbRegion,
				"--range=10.10.0.0/24"},
		}
		if _, err := shell.RunCommandAndGetOutputE(t, cmdCreateSubnet); err != nil {
			t.Fatalf("Error creating subnet %s in VPC %s: %v", currentSubnetName, networkName, err)
		}
		t.Logf("Successfully created Subnet: %s in VPC: %s", currentSubnetName, networkName)
	}
}

// deleteVPC deletes a Virtual Private Cloud (VPC) network and its associated subnet
func deleteVPC(t *testing.T, projectID string, networkName string) {
	// It's important to delete resources that depend on the VPC first, like MIGs, LBs, firewall rules.
	// Terraform destroy should handle most of this. This is a fallback.
	time.Sleep(30 * time.Second) // Wait for dependent resources to be potentially deleted by TF

	currentSubnetName := fmt.Sprintf("%s-subnet", networkName)
	cmdDeleteSubnet := shell.Command{
		Command: "gcloud",
		Args: []string{"compute", "networks", "subnets", "delete", currentSubnetName,
			"--project=" + projectID,
			"--region=" + nlbRegion, // Use nlbRegion
			"--quiet"},
	}
	// Log error but don't fail test, as it might have been cleaned up or not existed
	if _, err := shell.RunCommandAndGetOutputE(t, cmdDeleteSubnet); err != nil {
		t.Logf("Error deleting subnet %s: %v. This might be okay.", currentSubnetName, err)
	} else {
		t.Logf("Successfully deleted subnet %s.", currentSubnetName)
	}

	time.Sleep(60 * time.Second) // Increased delay before deleting VPC

	cmdDeleteVPC := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmdDeleteVPC); err != nil {
		t.Logf("Error deleting VPC %s: %v. This might be okay.", networkName, err)
	} else {
		t.Logf("Successfully deleted VPC %s.", networkName)
	}
}

// deleteFirewallRule generic delete firewall rule
func deleteFirewallRule(t *testing.T, projectID string, ruleName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "delete", ruleName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Failed to delete firewall rule %s: %v. This might be okay.", ruleName, err)
	} else {
		t.Logf("Successfully deleted firewall rule %s.", ruleName)
	}
}

func createFirewallRuleForIAP(t *testing.T, projectID, networkName, ruleName string, targetTags []string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "create", ruleName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--action=ALLOW",
			"--direction=INGRESS",
			"--rules=tcp:22",
			"--source-ranges=35.235.240.0/20", // Google's IAP IP range
			"--target-tags=" + strings.Join(targetTags, ","),
			"--description=Allow SSH via IAP",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		// It might be acceptable for this to fail if running tests in parallel and another created it.
		// However, for a single test run, this should ideally succeed.
		t.Logf("Warning: Failed to create firewall rule %s for IAP: %v. This might be an issue.", ruleName, err)
	} else {
		t.Logf("Successfully created firewall rule %s for IAP.", ruleName)
	}
}
