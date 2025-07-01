/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * You may not use this file except in compliance with the License.
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
	"regexp"
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

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/UMIG")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/UMIG/config")
)

var (
	projectID = os.Getenv("TF_VAR_project_id")

	zoneA  = "us-central1-a"
	zoneB  = "us-central1-b"
	region = "us-central1"

	randSuffix = fmt.Sprintf("%d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000))

	vpcName                         = fmt.Sprintf("testing-net-umig-%s", randSuffix)
	subnetNameA                     = fmt.Sprintf("testing-subnet-umig-a-%s", randSuffix)
	subnetNameB                     = fmt.Sprintf("testing-subnet-umig-b-%s", randSuffix)
	allowTCPFirewallRuleName        = fmt.Sprintf("allow-tcp-lb-and-health-%s", randSuffix)
	ipv6HealthCheckFirewallRuleName = fmt.Sprintf("fw-allow-lb-access-ipv6-%s", randSuffix)
	allowSSHTestVMFirewallRuleName  = fmt.Sprintf("allow-ssh-testvm-%s", randSuffix)

	instanceNamesA = []string{
		fmt.Sprintf("vm-a1-%s", randSuffix),
		fmt.Sprintf("vm-a2-%s", randSuffix),
	}
	instanceNamesB = []string{
		fmt.Sprintf("vm-b1-%s", randSuffix),
		fmt.Sprintf("vm-b2-%s", randSuffix),
	}
	allInstanceNames = append(instanceNamesA, instanceNamesB...)

	testVmName = fmt.Sprintf("test-connectivity-vm-%s", randSuffix)

	instanceGroupNameA = fmt.Sprintf("instance-group-a-%s", randSuffix)
	instanceGroupNameB = fmt.Sprintf("instance-group-b-%s", randSuffix)

	lbExternalIPName   = fmt.Sprintf("test-lb-ip-%s", randSuffix)
	healthCheckName    = fmt.Sprintf("my-tcp-health-check-%s", randSuffix)
	backendServiceName = fmt.Sprintf("test-bs-%s", randSuffix)
	targetProxyName    = fmt.Sprintf("my-tcp-proxy-%s", randSuffix)
	forwardingRuleName = fmt.Sprintf("my-tcp-lb-forwarding-rule-%s", randSuffix)
)

// UMIGConfig struct for YAML
type UMIGConfig struct {
	ProjectID   string `yaml:"project_id"`
	Zone        string `yaml:"zone"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Network     string `yaml:"network"`
	Instances   []struct {
		Name string `yaml:"name"`
	} `yaml:"instances"`
	NamedPorts []struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"named_ports"`
}

// TestUMIGAndLoadBalancer tests the creation of UMIGs and a Global External TCP Proxy Load Balancer
func TestUMIGAndLoadBalancer(t *testing.T) {

	// Variable to store the actual LB IP, initialized here for clarity, will be assigned later.
	var lbExternalIPAddress string

	defer func() {
		t.Log("Starting deferred cleanup functions...")

		// Clean up test VM first if it exists, as it might be blocking subnet deletion
		t.Logf("Attempting to delete test VM %s...", testVmName)
		retry.DoWithRetry(t, "Delete test VM", 3, 10*time.Second, func() (string, error) {
			err := deleteTestVM(t, projectID, zoneA, testVmName)
			if err != nil {
				return "", err
			}
			return "Test VM deleted successfully", nil
		})

		time.Sleep(10 * time.Second)

		// Attempt to delete firewalls that might have been created
		deleteFirewallRule(t, projectID, allowTCPFirewallRuleName)
		deleteFirewallRule(t, projectID, ipv6HealthCheckFirewallRuleName)
		deleteFirewallRule(t, projectID, allowSSHTestVMFirewallRuleName)
		time.Sleep(5 * time.Second)

		// Now attempt to delete other instances.
		deleteInstances(t, projectID, zoneA, zoneB, allInstanceNames)
		time.Sleep(10 * time.Second)

		// Finally, delete VPC and subnets
		deleteVPC(t, projectID, vpcName, subnetNameA, subnetNameB, region)
		t.Log("Deferred cleanup functions completed.")
	}()

	err := createVPCNetwork(t, projectID, vpcName)
	assert.NoErrorf(t, err, "VPC '%s' should be created successfully.", vpcName)
	err = createSubnet(t, projectID, subnetNameA, vpcName, region, "IPV4_IPv6", "EXTERNAL", "10.1.0.0/24")
	assert.NoErrorf(t, err, "Subnet '%s' should be created successfully.", subnetNameA)
	err = createSubnet(t, projectID, subnetNameB, vpcName, region, "IPV4_IPv6", "EXTERNAL", "10.1.1.0/24")
	assert.NoErrorf(t, err, "Subnet '%s' should be created successfully.", subnetNameB)
	time.Sleep(5 * time.Second)

	t.Log("Step 2: Creating Firewall Rules.")
	err = createFirewallRule(t, projectID, allowTCPFirewallRuleName, vpcName, "35.191.0.0/16,130.211.0.0/22,35.235.240.0/20", "tcp:110", "Allow TCP LB, Health Checks, and IAP (IPv4)", "1000", "tcp-lb")
	assert.NoErrorf(t, err, "Firewall rule '%s' should be created successfully.", allowTCPFirewallRuleName)
	err = createFirewallRule(t, projectID, ipv6HealthCheckFirewallRuleName, vpcName, "2600:2d00:1:b029::/64,2600:2d00:1:1::/64", "all", "Allow IPv6 LB Health Checks", "1000", "tcp-lb")
	assert.NoErrorf(t, err, "Firewall rule '%s' should be created successfully.", ipv6HealthCheckFirewallRuleName)
	err = createFirewallRule(t, projectID, allowSSHTestVMFirewallRuleName, vpcName, "35.235.240.0/20", "tcp:22", "Allow SSH to Test VM via IAP", "1000", "test-vm-ssh")
	assert.NoErrorf(t, err, "Firewall rule '%s' should be created successfully.", allowSSHTestVMFirewallRuleName)
	time.Sleep(5 * time.Second)

	t.Log("Step 3: Creating Instances with Apache TCP setup.")
	err = createInstances(t, projectID, zoneA, subnetNameA, instanceNamesA, "tcp-lb")
	assert.NoErrorf(t, err, "Instances in zoneA should be created successfully with startup script.")
	err = createInstances(t, projectID, zoneB, subnetNameB, instanceNamesB, "tcp-lb")
	assert.NoErrorf(t, err, "Instances in zoneB should be created successfully with startup script.")
	t.Log("Waiting for instances to boot and run startup scripts (Apache + TCP setup)...")
	time.Sleep(120 * time.Second)
	t.Log("Step 3a: Creating Test Connectivity VM.")
	err = createTestVM(t, projectID, zoneA, subnetNameA, testVmName)
	assert.NoErrorf(t, err, "Test connectivity VM '%s' should be created successfully.", testVmName)
	t.Log("Waiting for test VM to boot and SSH daemon to be ready...")
	time.Sleep(60 * time.Second)

	t.Log("Step 4: Creating UMIGs via Terraform.")
	createUMIGsConfigYAML(t)

	tfVars := map[string]interface{}{
		"config_folder_path": configFolderPath,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars:         tfVars,
		TerraformDir: terraformDirectoryPath,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	var umigSelfLinks map[string]gjson.Result

	// Defer cleanup for Terraform managed resources specifically
	defer func() {
		t.Log("Starting deferred Terraform destroy...")
		var umigSelfLinksListForCleanup []string
		if umigSelfLinks != nil {
			for _, sl := range umigSelfLinks {
				umigSelfLinksListForCleanup = append(umigSelfLinksListForCleanup, sl.String())
			}
		} else {
			// Fallback if umigSelfLinks map was not populated (e.g., if apply failed early)
			umigSelfLinksListForCleanup = []string{
				fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroups/%s", projectID, zoneA, instanceGroupNameA),
				fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroups/%s", projectID, zoneB, instanceGroupNameB),
			}
		}
		// Pass the correctly populated lbExternalIPAddress to cleanup
		deleteLoadBalancerComponents(t, projectID, "", lbExternalIPName, healthCheckName, backendServiceName, targetProxyName, forwardingRuleName, umigSelfLinksListForCleanup)

		// Run Terraform destroy
		terraform.Destroy(t, terraformOptions)
		t.Log("Deferred Terraform destroy completed.")
	}()

	terraform.InitAndApply(t, terraformOptions)
	_, err = terraform.ApplyE(t, terraformOptions)
	assert.NoErrorf(t, err, "Terraform apply for UMIGs should succeed.")

	umigInstancesOutput := terraform.OutputJson(t, terraformOptions, "umig_instances")
	umigSelfLinksOutput := terraform.OutputJson(t, terraformOptions, "umig_self_links")

	t.Logf("UMIG Instances Output: %s", umigInstancesOutput)
	t.Logf("UMIG Self Links Output: %s", umigSelfLinksOutput)

	umigSelfLinks = gjson.Parse(umigSelfLinksOutput).Map()

	expectedUMIGs := map[string]struct {
		GroupName string
		Zone      string
		Instances []string
	}{
		"umig-0": {GroupName: instanceGroupNameA, Zone: zoneA, Instances: instanceNamesA},
		"umig-1": {GroupName: instanceGroupNameB, Zone: zoneB, Instances: instanceNamesB},
	}

	for umigKey, expected := range expectedUMIGs {
		t.Logf("Checking UMIG: %s (Expected: %s in %s)", umigKey, expected.GroupName, expected.Zone)
		// Describe the instance group
		describeCmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "instance-groups", "describe", expected.GroupName,
				"--zone", expected.Zone,
				"--project", projectID,
				"--format", "json",
			},
		}
		instanceGroupDescribeOutput := shell.RunCommandAndGetOutput(t, describeCmd)
		instanceGroupDescribeOutput = strings.TrimSpace(instanceGroupDescribeOutput)

		// Parse namedPorts as an array
		namedPorts := gjson.Get(instanceGroupDescribeOutput, "namedPorts").Array()
		assert.True(t, len(namedPorts) >= 2, "UMIG should have at least 2 named ports.")

		// Check for both named ports
		hasTcp110 := false
		hasHttp := false
		for _, np := range namedPorts {
			if np.Get("name").String() == "tcp110" && np.Get("port").Int() == 110 {
				hasTcp110 = true
			}
			if np.Get("name").String() == "http" && np.Get("port").Int() == 80 {
				hasHttp = true
			}
		}
		assert.True(t, hasTcp110, "Named port tcp110:110 should be present in UMIG %s", umigKey)
		assert.True(t, hasHttp, "Named port http:80 should be present in UMIG %s", umigKey)
	}
	t.Log("Step 5: Setting up Load Balancer components for Global External TCP Proxy LB...")

	t.Log("  Creating Global External IP Address (IPv4).")
	// Create the external IP address for the Load Balancer
	lbExternalIPAddress = createExternalIPAddress(t, projectID, lbExternalIPName, "IPV4")
	t.Logf("Load Balancer External IP Address: %s", lbExternalIPAddress)
	time.Sleep(5 * time.Second)

	t.Log("  Creating TCP Health Check.")
	err = createHealthCheck(t, projectID, healthCheckName, "TCP", 110)
	assert.NoErrorf(t, err, "Health Check '%s' should be created.", healthCheckName)
	hcDescribe := shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "health-checks", "describe", healthCheckName, "--project", projectID, "--format=json"},
	})
	hcDescribe = strings.TrimSpace(hcDescribe)
	assert.Equal(t, "TCP", gjson.Get(hcDescribe, "type").String(), "Health Check protocol should be TCP.")
	assert.Equal(t, 110, int(gjson.Get(hcDescribe, "tcpHealthCheck.port").Int()), "Health Check port should be 110.")
	t.Log("Assertion: Health check protocol and port are correct.")
	time.Sleep(5 * time.Second)

	t.Log("  Creating TCP Backend Service.")
	var umigSelfLinksList []string
	for _, sl := range umigSelfLinks {
		umigSelfLinksList = append(umigSelfLinksList, sl.String())
	}
	err = createBackendService(t, projectID, backendServiceName, healthCheckName, umigSelfLinksList, "TCP", "tcp110")
	assert.NoErrorf(t, err, "Backend Service '%s' should be created.", backendServiceName)
	bsDescribe := shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "backend-services", "describe", backendServiceName, "--project", projectID, "--global", "--format=json"},
	})
	bsDescribe = strings.TrimSpace(bsDescribe)
	assert.Equal(t, "TCP", gjson.Get(bsDescribe, "protocol").String(), "Backend Service protocol should be TCP.")
	assert.Equal(t, "tcp110", gjson.Get(bsDescribe, "portName").String(), "Backend Service port name should be 'tcp110'.")
	assert.True(t, len(gjson.Get(bsDescribe, "backends").Array()) == 2, "Backend Service should have 2 backends.")
	t.Log("Assertion: Backend service protocol, port name, and backend count are correct.")
	time.Sleep(5 * time.Second)

	t.Log("  Creating Target TCP Proxy.")
	createTargetTCPProxy(t, projectID, targetProxyName, backendServiceName)
	assert.NoErrorf(t, err, "Target TCP Proxy '%s' should be created.", targetProxyName)
	time.Sleep(5 * time.Second)

	t.Log("  Creating Global Forwarding Rule.")
	// Use lbExternalIPName here as it's the *name* of the address resource, not the IP itself.
	createForwardingRuleTCP(t, projectID, forwardingRuleName, lbExternalIPName, targetProxyName, 110)
	assert.NoErrorf(t, err, "Forwarding Rule '%s' should be created.", forwardingRuleName)
	t.Log("Waiting for Load Balancer to provision and health checks to pass...")
	time.Sleep(180 * time.Second)

	t.Logf("Step 6: Performing connectivity test to Load Balancer IP: %s on port 110 using curl", lbExternalIPAddress)

	// Curl directly from the test environment, removing the test VM as a dependency for this step.
	// The pattern expects "hello from vm-a1-xxxx" or "hello from vm-b1-xxxx"
	expectedContentPattern := `hello from vm-(a|b)[1-2]-\d+`
	verifyConnectivityToLBDirect(t, lbExternalIPAddress, "110", expectedContentPattern)

	t.Logf("Test Passed: Successfully connected to LB IP %s on port 110.", lbExternalIPAddress)
}

// createUMIGsConfigYAML creates two separate UMIG YAML config files
func createUMIGsConfigYAML(t *testing.T) {
	writeUMIGConfig := func(t *testing.T, groupName, zone, subnetName string, instanceNames []string) {
		instancesForUMIG := make([]struct {
			Name string `yaml:"name"`
		}, len(instanceNames))

		for i, name := range instanceNames {
			instancesForUMIG[i] = struct {
				Name string `yaml:"name"`
			}{
				Name: name,
			}
		}

		umigConfig := UMIGConfig{
			ProjectID:   projectID,
			Zone:        zone,
			Name:        groupName,
			Description: fmt.Sprintf("Test UMIG for integration test (%s)", zone),
			Network:     vpcName,
			Instances:   instancesForUMIG,
			NamedPorts: []struct {
				Name string `yaml:"name"`
				Port int    `yaml:"port"`
			}{
				{Name: "tcp110", Port: 110},
				{Name: "http", Port: 80},
			},
		}

		yamlData, err := yaml.Marshal(&umigConfig)
		if err != nil {
			t.Fatalf("Error marshaling UMIG YAML for %s: %v", groupName, err)
		}

		configDir := configFolderPath
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		filePath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", groupName))
		if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
			t.Fatalf("Unable to write UMIG YAML for %s: %v", groupName, err)
		}
		t.Logf("Created UMIG YAML config at %s:\n%s", filePath, string(yamlData))
	}

	writeUMIGConfig(t, instanceGroupNameA, zoneA, subnetNameA, instanceNamesA)
	writeUMIGConfig(t, instanceGroupNameB, zoneB, subnetNameB, instanceNamesB)
}

// createVPCNetwork creates a custom mode VPC network and returns the error string if any
func createVPCNetwork(t *testing.T, projectID, vpcName string) error {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"compute", "networks", "create", vpcName, "--project=" + projectID, "--subnet-mode=custom"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating VPC network '%s': %v\nOutput: %s", vpcName, err, output)
	}
	t.Logf("VPC network '%s' created.", vpcName)
	return nil
}

func createSubnet(t *testing.T, projectID, subnetName, networkName, region, stackType, ipv6AccessType, ipRange string) error {
	args := []string{
		"compute", "networks", "subnets", "create", subnetName,
		"--project=" + projectID,
		"--network=" + networkName,
		"--region=" + region,
		"--range=" + ipRange,
		"--stack-type=" + stackType,
	}
	if stackType == "IPV4_IPv6" && ipv6AccessType != "" {
		args = append(args, "--ipv6-access-type="+ipv6AccessType)
	}

	cmd := shell.Command{
		Command: "gcloud",
		Args:    args,
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating subnet '%s': %v", subnetName, err)
	}
	t.Logf("Subnet '%s' created.", subnetName)
	return nil
}

func createFirewallRule(t *testing.T, projectID, ruleName, network, sourceRanges, allow, description, priority, targetTags string) error {
	args := []string{
		"compute", "firewall-rules", "create", ruleName,
		"--project=" + projectID,
		"--network=" + network,
		"--allow=" + allow,
		"--source-ranges=" + sourceRanges,
		"--description=" + description,
		"--priority=" + priority,
		"--quiet",
	}
	if targetTags != "" {
		args = append(args, "--target-tags="+targetTags)
	}
	if strings.Contains(description, "Health Checks") {
		args = append(args, "--enable-logging")
	}

	cmd := shell.Command{
		Command: "gcloud",
		Args:    args,
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			t.Logf("Firewall rule '%s' already exists. Skipping creation.", ruleName)
			return nil
		} else if strings.Contains(err.Error(), "Quota 'FIREWALLS' exceeded") {
			return fmt.Errorf("CRITICAL ERROR: Quota 'FIREWALLS' exceeded when creating rule '%s'. Please increase your project quota or delete unused firewall rules. Error: %v", ruleName, err)
		} else {
			return fmt.Errorf("Error creating firewall rule '%s': %v", ruleName, err)
		}
	}
	t.Logf("Firewall rule '%s' created", ruleName)
	return nil
}

func createInstances(t *testing.T, projectID, zone, subnetName string, names []string, tags string) error {
	for _, name := range names {
		startupScriptContent := fmt.Sprintf(`#! /bin/bash
sudo apt-get update
sudo apt-get install apache2 -y
# Stop any service using port 110 (like dovecot or pop3d)
sudo systemctl stop dovecot || true
sudo systemctl disable dovecot || true
sudo fuser -k 110/tcp || true
# Change Apache to listen on port 110
echo 'Listen 110' | sudo tee /etc/apache2/ports.conf
sudo sed -i 's/<VirtualHost \*:80>/<VirtualHost \*:110>/' /etc/apache2/sites-available/000-default.conf
sudo sed -i '/<Directory \/var\/www\/>/a\        AllowOverride All' /etc/apache2/apache2.conf
sudo systemctl restart apache2
echo "hello from %s" | sudo tee /var/www/html/index.html
`, name)

		tempScriptFile, err := os.CreateTemp("", fmt.Sprintf("startup-script-%s-*.sh", name))
		if err != nil {
			return fmt.Errorf("Failed to create temp startup script file: %v", err)
		}
		defer func(filePath string) {
			if err := os.Remove(filePath); err != nil {
				t.Logf("Failed to delete temp script file %s: %v", filePath, err)
			}
		}(tempScriptFile.Name())

		_, err = tempScriptFile.WriteString(startupScriptContent)
		if err != nil {
			tempScriptFile.Close()
			return fmt.Errorf("Failed to write to temp startup script file: %v", err)
		}
		tempScriptFile.Close()

		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "instances", "create", name,
				"--project=" + projectID,
				"--zone=" + zone,
				"--subnet=" + subnetName,
				"--machine-type=e2-micro",
				"--image-family=ubuntu-2204-lts",
				"--image-project=ubuntu-os-cloud",
				"--stack-type=IPV4_IPV6",
				"--no-address",
				"--tags=" + tags,
				"--metadata-from-file", fmt.Sprintf("startup-script=%s", tempScriptFile.Name()),
			},
		}
		_, err = shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("Error creating instance %s: %v", name, err)
		}
		t.Logf("Instance %s created.", name)
	}
	return nil
}

func createTestVM(t *testing.T, projectID, zone, subnetName, vmName string) error {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "instances", "create", vmName,
			"--project=" + projectID,
			"--zone=" + zone,
			"--subnet=" + subnetName,
			"--machine-type=e2-micro",
			"--image-family=ubuntu-2204-lts",
			"--image-project=ubuntu-os-cloud",
			"--no-address",
			"--tags=test-vm-ssh", // Tag for firewall rule
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating test VM '%s': %v", vmName, err)
	}
	t.Logf("Test VM '%s' created.", vmName)
	return nil
}

func createHealthCheck(t *testing.T, projectID, hcName, protocol string, port int) error {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "health-checks", "create", strings.ToLower(protocol), hcName,
			"--project=" + projectID,
			fmt.Sprintf("--port=%d", port),
			"--check-interval=5s",
			"--timeout=5s",
			"--unhealthy-threshold=2",
			"--healthy-threshold=2",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating Health Check '%s': %v", hcName, err)
	}
	t.Logf("Health Check '%s' created.", hcName)
	return nil
}

// createExternalIPAddress creates a global external IP address
func createExternalIPAddress(t *testing.T, projectID, ipName, ipVersion string) string {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "addresses", "create", ipName,
			"--project=" + projectID,
			"--global",
			"--ip-version=" + ipVersion,
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Error creating external IP address '%s': %v", ipName, err)
	}
	t.Logf("External IP address '%s' created.", ipName)

	getIPCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "addresses", "describe", ipName,
			"--project=" + projectID,
			"--global",
			"--format=json",
		},
	}
	output := shell.RunCommandAndGetOutput(t, getIPCmd)
	output = strings.TrimSpace(output)

	return gjson.Get(output, "address").String()
}

// createBackendService creates a global backend service and adds multiple instance groups
func createBackendService(t *testing.T, projectID, bsName, hcName string, umigSelfLinks []string, protocol, portName string) error {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "backend-services", "create", bsName,
			"--project=" + projectID,
			"--protocol=" + protocol,
			"--health-checks=" + hcName,
			"--global",
			"--port-name=" + portName,
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating Backend Service '%s': %v", bsName, err)
	}
	t.Logf("Backend Service '%s' created.", bsName)

	for _, umigSelfLink := range umigSelfLinks {
		parts := strings.Split(umigSelfLink, "/")
		umigZone := parts[len(parts)-2]

		addBackendCmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "backend-services", "add-backend", bsName,
				"--project=" + projectID,
				"--instance-group=" + umigSelfLink,
				"--instance-group-zone=" + umigZone,
				"--global",
			},
		}
		_, err = shell.RunCommandAndGetOutputE(t, addBackendCmd)
		if err != nil && !strings.Contains(err.Error(), "already member") && !strings.Contains(err.Error(), "was not found") {
			return fmt.Errorf("Error adding backend %s to service '%s': %v", umigSelfLink, bsName, err)
		}
		t.Logf("UMIG %s added to Backend Service '%s'.", umigSelfLink, bsName)
		time.Sleep(2 * time.Second)
	}
	return nil
}

// createTargetTCPProxy creates a global target TCP proxy
func createTargetTCPProxy(t *testing.T, projectID, tpName, bsName string) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "target-tcp-proxies", "create", tpName,
			"--project=" + projectID,
			"--backend-service=" + bsName,
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Error creating Target TCP Proxy '%s': %v", tpName, err)
	}
	t.Logf("Target TCP Proxy '%s' created.", tpName)
}

// createForwardingRuleTCP creates a global forwarding rule for TCP proxy LB
func createForwardingRuleTCP(t *testing.T, projectID, frName, ipName, tpName string, port int) {
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "forwarding-rules", "create", frName,
			"--project=" + projectID,
			"--address=" + ipName, // This is the *name* of the IP address resource, not the actual IP.
			"--global",
			"--target-tcp-proxy=" + tpName,
			fmt.Sprintf("--ports=%d", port),
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Error creating Forwarding Rule '%s': %v", frName, err)
	}
	t.Logf("Forwarding Rule '%s' created.", frName)
}

// verifyConnectivityToLBDirect: Uses curl directly from the test runner to the LB IP.
func verifyConnectivityToLBDirect(t *testing.T, lbIpAddress, port, expectedTextPattern string) {
	t.Logf("Verifying connectivity to LB IP %s on port %s using curl directly, expecting to match: '%s'", lbIpAddress, port, expectedTextPattern)

	maxRetries := 5
	sleepBetweenRetries := 20 * time.Second
	// Note: This is a direct curl command to the LB IP address on port 110.
	// For HTTP-like content on a custom TCP port, curl often works.
	curlCommand := "curl"
	curlArgs := []string{"--retry", "5", "--retry-delay", "5", "--connect-timeout", "10", fmt.Sprintf("http://%s:%s", lbIpAddress, port)}
	// Note: We use http:// here because Apache is serving HTTP content over TCP port 110.

	expectedRegex, regexCompileErr := regexp.Compile(expectedTextPattern)
	if regexCompileErr != nil {
		t.Fatalf("Invalid regex pattern '%s' provided for matching: %v", expectedTextPattern, regexCompileErr)
	}

	curlOutput, err := retry.DoWithRetryE(t, fmt.Sprintf("Curl LB %s:%s", lbIpAddress, port), maxRetries, sleepBetweenRetries, func() (string, error) {
		cmd := shell.Command{Command: curlCommand, Args: curlArgs}
		output, runErr := shell.RunCommandAndGetOutputE(t, cmd)

		t.Logf("DEBUG: Attempting curl command (%s %s). Output:\n%s", curlCommand, strings.Join(curlArgs, " "), output)

		if runErr != nil {
			return "", fmt.Errorf("curl command failed. Error: %v. Full Output:\n%s", runErr, output)
		}

		if !expectedRegex.MatchString(output) {
			return "", fmt.Errorf("curl succeeded but response from LB %s:%s did not match expected pattern '%s'. Full Output:\n%s", lbIpAddress, port, expectedTextPattern, output)
		}

		return output, nil
	})

	if err != nil {
		t.Errorf("Failed to verify connectivity to LB %s:%s after %d attempts (total duration: %s): %v",
			lbIpAddress, port, maxRetries, time.Duration(maxRetries)*sleepBetweenRetries, err)
		t.Fail()
	} else {
		t.Logf("Final curl output: %s", curlOutput)
	}
}

// deleteInstances deletes the test VM instances
func deleteInstances(t *testing.T, projectID, zoneA, zoneB string, instanceNames []string) {
	t.Log("Cleaning up instances...")
	for _, name := range instanceNames {
		currentZone := zoneA
		if strings.HasPrefix(name, "vm-b") {
			currentZone = zoneB
		}

		retry.DoWithRetry(t, fmt.Sprintf("Delete instance %s", name), 3, 10*time.Second, func() (string, error) {
			cmd := shell.Command{
				Command: "gcloud",
				Args: []string{
					"compute", "instances", "delete", name,
					"--project=" + projectID,
					"--zone=" + currentZone,
					"--quiet",
				},
			}
			_, err := shell.RunCommandAndGetOutputE(t, cmd)
			if err != nil && !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Resource not found") {
				return "", fmt.Errorf("Error deleting instance %s in zone %s: %v", name, currentZone, err)
			}
			t.Logf("Instance %s in zone %s deleted", name, currentZone)
			return "Instance deleted", nil
		})
	}
}

// deleteTestVM deletes the dedicated test VM
func deleteTestVM(t *testing.T, projectID, zone, vmName string) error {
	t.Logf("Cleaning up test VM '%s'...", vmName)
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
	if err != nil && !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Resource not found") {
		return fmt.Errorf("Error deleting test VM %s: %v", vmName, err)
	} else if err == nil {
		t.Logf("Test VM %s deleted.", vmName)
	}
	return nil
}

// deleteFirewallRule deletes a specific firewall rule
func deleteFirewallRule(t *testing.T, projectID, ruleName string) {
	t.Logf("Cleaning up firewall rule: %s...", ruleName)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "delete", ruleName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Resource not found") {
		t.Logf("Error deleting firewall rule %s: %v", ruleName, err)
	} else if err == nil {
		t.Logf("Firewall rule %s deleted", ruleName)
	}
	time.Sleep(2 * time.Second)
}

// deleteLoadBalancerComponents deletes all load balancer related resources for a TCP Proxy LB
func deleteLoadBalancerComponents(t *testing.T, projectID, sslCertName, ipName, hcName, bsName, tpName, frName string, umigSelfLinks []string) {
	t.Log("Cleaning up Load Balancer components...")

	// Remove backends from backend service first
	for _, umigSelfLink := range umigSelfLinks {
		parts := strings.Split(umigSelfLink, "/")
		umigZone := parts[len(parts)-2]

		removeBackendCmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "backend-services", "remove-backend", bsName,
				"--project=" + projectID,
				"--instance-group=" + umigSelfLink,
				"--instance-group-zone=" + umigZone,
				"--global",
				"--quiet",
			},
		}
		if _, err := shell.RunCommandAndGetOutputE(t, removeBackendCmd); err != nil {
			if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") || strings.Contains(err.Error(), "backend not found") {
				t.Logf("Backend %s was not found in service %s or already removed. Skipping.", umigSelfLink, bsName)
			} else {
				t.Logf("Error removing backend %s from service %s: %v", umigSelfLink, bsName, err)
			}
		} else {
			t.Logf("Backend %s removed from service %s.", umigSelfLink, bsName)
		}
		time.Sleep(5 * time.Second)
	}

	// Delete forwarding rule
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "forwarding-rules", "delete", frName,
			"--project=" + projectID,
			"--global",
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
			t.Logf("Forwarding rule %s was not found. Skipping deletion.", frName)
		} else {
			t.Logf("Error deleting forwarding rule %s: %v", frName, err)
		}
	} else {
		t.Logf("Forwarding rule %s deleted.", frName)
	}
	time.Sleep(5 * time.Second)

	// Delete target TCP proxy
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "target-tcp-proxies", "delete", tpName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
			t.Logf("Target TCP proxy %s was not found. Skipping deletion.", tpName)
		} else {
			t.Logf("Error deleting target TCP proxy %s: %v", tpName, err)
		}
	} else {
		t.Logf("Target TCP proxy %s deleted.", tpName)
	}
	time.Sleep(5 * time.Second)

	// Delete backend service
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "backend-services", "delete", bsName,
			"--project=" + projectID,
			"--global",
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
			t.Logf("Backend service %s was not found. Skipping deletion.", bsName)
			// Check specifically if the error is due to backends still attached
			if strings.Contains(err.Error(), "The backendService resource") && strings.Contains(err.Error(), "is not empty") {
				t.Logf("WARNING: Backend service %s not empty, retrying backend removal in case of race condition.", bsName)
				// Re-attempt backend removal (though ideally handled by earlier loop)
				for _, umigSelfLink := range umigSelfLinks {
					parts := strings.Split(umigSelfLink, "/")
					umigZone := parts[len(parts)-2]
					removeBackendCmd := shell.Command{
						Command: "gcloud",
						Args: []string{
							"compute", "backend-services", "remove-backend", bsName,
							"--project=" + projectID,
							"--instance-group=" + umigSelfLink,
							"--instance-group-zone=" + umigZone,
							"--global",
							"--quiet",
						},
					}
					shell.RunCommandAndGetOutputE(t, removeBackendCmd)
				}
				time.Sleep(10 * time.Second)
				// Re-attempt deleting backend service
				shell.RunCommandAndGetOutputE(t, cmd)
			}
		} else {
			t.Logf("Error deleting backend service %s: %v", bsName, err)
		}
	} else {
		t.Logf("Backend service %s deleted.", bsName)
	}
	time.Sleep(5 * time.Second)

	// Delete health check
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "health-checks", "delete", hcName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
			t.Logf("Health check %s was not found. Skipping deletion.", hcName)
		} else {
			t.Logf("Error deleting health check %s: %v", hcName, err)
		}
	} else {
		t.Logf("Health check %s deleted.", hcName)
	}
	time.Sleep(5 * time.Second)

	// Delete external IP address
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "addresses", "delete", ipName,
			"--project=" + projectID,
			"--global",
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
			t.Logf("External IP address %s was not found. Skipping deletion.", ipName)
		} else {
			t.Logf("Error deleting external IP address %s: %v", ipName, err)
		}
	} else {
		t.Logf("External IP address %s deleted.", ipName)
	}
	time.Sleep(5 * time.Second)

	// SSL certificate is not used for TCP proxy
	if sslCertName != "" { // Only attempt if it's explicitly set
		cmd = shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "ssl-certificates", "delete", sslCertName,
				"--project=" + projectID,
				"--quiet",
			},
		}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") {
				t.Logf("SSL certificate %s was not found. Skipping deletion.", sslCertName)
			} else {
				t.Logf("Error deleting SSL certificate %s: %v", sslCertName, err)
			}
		} else {
			t.Logf("SSL certificate %s deleted.", sslCertName)
		}
		time.Sleep(5 * time.Second)
	}
}

// deleteVPC deletes the VPC and subnets
func deleteVPC(t *testing.T, projectID, vpcName, subnetNameA, subnetNameB, region string) {
	t.Log("Cleaning up VPC and subnets...")

	// Give some time after main resources are destroyed
	time.Sleep(30 * time.Second)

	subnetsToDelete := []string{subnetNameA, subnetNameB}
	for _, sn := range subnetsToDelete {
		// Use a retry mechanism for subnet deletion as well, as they can sometimes be in use momentarily.
		retry.DoWithRetry(t, fmt.Sprintf("Delete subnet %s", sn), 3, 10*time.Second, func() (string, error) {
			cmd := shell.Command{
				Command: "gcloud",
				Args: []string{
					"compute", "networks", "subnets", "delete", sn,
					"--project=" + projectID,
					"--region=" + region,
					"--quiet",
				},
			}
			_, err := shell.RunCommandAndGetOutputE(t, cmd)
			if err != nil && !strings.Contains(err.Error(), "was not found") && !strings.Contains(err.Error(), "Resource not found") && !strings.Contains(err.Error(), "is already being used by") {
				return "", fmt.Errorf("Error deleting subnet %s: %v", sn, err)
			} else if err != nil && strings.Contains(err.Error(), "is already being used by") {
				t.Logf("Subnet %s is still in use, retrying...", sn)
				return "", fmt.Errorf("subnet still in use")
			}
			t.Logf("Subnet %s deleted", sn)
			return "Subnet deleted", nil
		})
	}
	// Allow some time for subnets to be fully deleted before deleting the VPC
	time.Sleep(10 * time.Second)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "networks", "delete", vpcName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "Resource not found") || strings.Contains(err.Error(), "is already being used by") {
			t.Logf("VPC %s was not found or is still in use. Skipping deletion: %v", vpcName, err)
		} else {
			t.Logf("Error deleting VPC %s: %v", vpcName, err)
		}
	} else {
		t.Logf("VPC %s deleted.", vpcName)
	}
}
