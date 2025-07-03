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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "03-security/SecurityProfile")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/security/SecurityProfile/config")
	requiredGcpApis        = []string{
		"iam.googleapis.com",
		"compute.googleapis.com",
		"networksecurity.googleapis.com",
		"cloudresourcemanager.googleapis.com",
	}

	testSaProjectRoles = []string{
		"roles/compute.securityAdmin",
		"roles/compute.admin",
		"roles/serviceusage.serviceUsageConsumer",
		"roles/networksecurity.securityProfileAdmin",
		"roles/compute.networkAdmin",
	}
	testSaOrgRoles = []string{
		"roles/compute.orgFirewallPolicyAdmin",
		"roles/resourcemanager.organizationViewer",
		"roles/networksecurity.securityProfileAdmin",
	}
	resourceVpcSubnetRange = "10.10.10.0/24"
	sshFirewallRange       = "35.235.240.0/20"
)

func TestSecurityProfileIntegration(t *testing.T) {
	t.Parallel()
	projectID := os.Getenv("TF_VAR_project_id")
	orgID := os.Getenv("TF_VAR_organization_id")
	billingProjectID := os.Getenv("TF_VAR_billing_project_id")
	require.NotEmpty(t, projectID, "TF_VAR_project_id env var must be set")
	if orgID == "" {
		t.Skip("SKIPPING TEST: TF_VAR_organization_id environment variable is not set.")
	}
	if billingProjectID == "" {
		t.Skip("SKIPPING TEST: TF_VAR_billing_project_id environment variable is not set.")
	}
	err := setQuotaProjectE(t, billingProjectID)
	require.NoError(t, err)
	defer unsetQuotaProject(t)
	currentUser := getCurrentGcloudUser(t)
	instanceSuffix := strings.ToLower(random.UniqueId())
	serviceAccountName := fmt.Sprintf("sa-sp-test-%s", instanceSuffix)
	vpcName := fmt.Sprintf("vpc-sp-test-%s", instanceSuffix)
	firewallPolicyName := fmt.Sprintf("fwp-sp-test-%s", instanceSuffix)
	zone := "us-central1-a"
	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, Zone=%s, Suffix=%s", projectID, orgID, zone, instanceSuffix)
	enableGcpApis(t, projectID, requiredGcpApis)
	serviceAccountEmail, err := createServiceAccount(t, projectID, serviceAccountName, "Security Profile Test SA")
	assert.NoError(t, err)
	defer deleteServiceAccount(t, projectID, serviceAccountEmail)
	time.Sleep(15 * time.Second)
	addTokenCreatorRoleToPrincipal(t, projectID, serviceAccountEmail, "user:"+currentUser)
	defer removeTokenCreatorRoleFromPrincipal(t, projectID, serviceAccountEmail, "user:"+currentUser)
	addProjectIamBindings(t, projectID, serviceAccountEmail, testSaProjectRoles)
	defer removeProjectIamBindings(t, projectID, serviceAccountEmail, testSaProjectRoles)
	addOrgIamBindings(t, orgID, serviceAccountEmail, testSaOrgRoles)
	defer removeOrgIamBindings(t, orgID, serviceAccountEmail, testSaOrgRoles)
	t.Log("Waiting 60 seconds for IAM permissions to propagate...")
	time.Sleep(60 * time.Second)
	createVPC(t, projectID, vpcName, zone)
	defer deleteVPC(t, projectID, vpcName, zone)
	vmClientName := "vm-client-" + instanceSuffix
	vmServerName := "vm-server-" + instanceSuffix
	createVM(t, projectID, vmClientName, zone, vpcName)
	defer deleteVM(t, projectID, vmClientName, zone)
	createVM(t, projectID, vmServerName, zone, vpcName)
	defer deleteVM(t, projectID, vmServerName, zone)
	profileGroupName := "spg-integ-test-" + instanceSuffix
	createConfigYAML(t, orgID, "sp-integ-test-"+instanceSuffix, profileGroupName)
	createFirewallPolicy(t, orgID, firewallPolicyName)
	defer deleteFirewallPolicy(t, orgID, firewallPolicyName)
	tfVars := map[string]interface{}{
		"config_folder_path": configFolderPath,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		NoColor:      true,
		EnvVars: map[string]string{
			"GOOGLE_PROJECT":                     projectID,
			"GOOGLE_IMPERSONATE_SERVICE_ACCOUNT": serviceAccountEmail,
		},
	})
	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")
	addRuleAndAssociateFirewallPolicy(t, orgID, firewallPolicyName, vpcName, projectID, profileGroupName)
	defer deleteRuleAndFirewallPolicyAssociation(t, orgID, firewallPolicyName)
	t.Log("Validating that the security profile is blocking traffic...")
	err = verifyConnectivity(t, projectID, zone, vmClientName, vmServerName, false)
	require.NoError(t, err, "verifyConnectivity reported an unexpected error. It should have confirmed that the connection was blocked, but instead it saw a success or another error.")
	t.Log("Validation successful: Traffic was correctly blocked by the security profile.")
}

func setQuotaProjectE(t *testing.T, projectID string) error {
	t.Logf("Setting gcloud billing/quota_project to: %s", projectID)
	cmd := shell.Command{
		Command: "gcloud",
		// This array translates to "gcloud config set billing/quota_project PROJECT_ID"
		Args: []string{"config", "set", "billing/quota_project", projectID},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	return err
}

func unsetQuotaProject(t *testing.T) {
	t.Logf("Unsetting gcloud billing/quota_project.")
	cmd := shell.Command{
		Command: "gcloud",
		// This array translates to "gcloud config unset billing/quota_project"
		Args: []string{"config", "unset", "billing/quota_project"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Logf("WARN: Failed to unset quota project. Manual cleanup may be required. Error: %v", err)
	}
}

func createConfigYAML(t *testing.T, orgID, profileName, groupName string) {
	type securityProfile struct {
		Create                  bool                   `yaml:"create"`
		Name                    string                 `yaml:"name"`
		Type                    string                 `yaml:"type"`
		Description             string                 `yaml:"description"`
		ThreatPreventionProfile map[string]interface{} `yaml:"threat_prevention_profile"`
	}
	type securityProfileGroup struct {
		Create bool   `yaml:"create"`
		Name   string `yaml:"name"`
	}
	type testConfig struct {
		OrgID   string               `yaml:"organization_id"`
		Profile securityProfile      `yaml:"security_profile"`
		Group   securityProfileGroup `yaml:"security_profile_group"`
		Link    bool                 `yaml:"link_profile_to_group"`
	}
	config := testConfig{
		OrgID: orgID,
		Profile: securityProfile{
			Create:      true,
			Name:        profileName,
			Type:        "THREAT_PREVENTION",
			Description: "Deny INFORMATIONAL traffic for testing",
			ThreatPreventionProfile: map[string]interface{}{
				"severity_overrides": []map[string]string{
					{"severity": "INFORMATIONAL", "action": "DENY"},
				},
			},
		},
		Group: securityProfileGroup{
			Create: true,
			Name:   groupName,
		},
		Link: true,
	}
	yamlData, err := yaml.Marshal(&config)
	assert.NoError(t, err)

	err = os.MkdirAll(configFolderPath, 0755)
	assert.NoError(t, err)

	filePath := filepath.Join(configFolderPath, "instance.yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	assert.NoError(t, err)
	t.Logf("Created test YAML config file: %s", filePath)
}

func createFirewallPolicy(t *testing.T, orgID, policyName string) {
	t.Logf("Creating Firewall Policy '%s' in Org '%s'", policyName, orgID)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "create", "--short-name=" + policyName, "--organization=" + orgID, "--description=integ-test-policy"}})
}

func addRuleAndAssociateFirewallPolicy(t *testing.T, orgID, policyName, vpcName, projectID, profileGroupName string) {
	profileGroupPath := fmt.Sprintf("organizations/%s/locations/global/securityProfileGroups/%s", orgID, profileGroupName)
	vpcPath := fmt.Sprintf("projects/%s/global/networks/%s", projectID, vpcName)

	t.Logf("Adding rule to policy '%s' to apply security profile group '%s'", policyName, profileGroupPath)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "rules", "create", "1000", "--firewall-policy=" + policyName, "--organization=" + orgID, "--action=apply_security_profile_group", "--security-profile-group=" + profileGroupPath, "--src-ip-ranges=" + resourceVpcSubnetRange, "--layer4-configs=all", "--enable-logging", "--description=test-rule"}})

	t.Logf("Associating policy '%s' with VPC '%s'", policyName, vpcPath)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "associations", "create", "--firewall-policy=" + policyName, "--organization=" + orgID, fmt.Sprintf("--name=%s-association", policyName), "--replace-association-on-target"}})
}

func deleteRuleAndFirewallPolicyAssociation(t *testing.T, orgID, policyName string) {
	if policyName == "" {
		return
	}
	t.Logf("--- Deleting Firewall Policy Association: %s-association ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "associations", "delete", fmt.Sprintf("%s-association", policyName), "--firewall-policy=" + policyName, "--organization=" + orgID}})

	t.Logf("--- Deleting Firewall Policy Rule '1000' from policy '%s' ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "rules", "delete", "1000", "--firewall-policy=" + policyName, "--organization=" + orgID}})
}

func deleteFirewallPolicy(t *testing.T, orgID, policyName string) {
	if policyName == "" {
		return
	}
	t.Logf("--- Deleting Firewall Policy: %s ---", policyName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-policies", "delete", policyName, "--organization=" + orgID, "--quiet"}})
}

func verifyConnectivity(t *testing.T, projectID, zone, clientVM, serverVM string, expectSuccess bool) error {
	getIpCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "describe", serverVM, "--project=" + projectID, "--zone=" + zone, "--format=get(networkInterfaces[0].networkIP)"}}
	serverIP, err := shell.RunCommandAndGetOutputE(t, getIpCmd)
	assert.NoError(t, err)
	serverIP = strings.TrimSpace(serverIP)

	command := fmt.Sprintf(`curl -s -o /dev/null -w "%%{http_code}" --connect-timeout 5 http://%s`, serverIP)
	runCmd := shell.Command{Command: "gcloud", Args: []string{"compute", "ssh", clientVM, "--project=" + projectID, "--zone=" + zone, "--command=" + command}}

	var lastErr error
	for i := 0; i < 5; i++ {
		t.Logf("Attempting to run on %s: curl to %s (Attempt %d/5)", clientVM, serverIP, i+1)

		httpCode, cmdErr := shell.RunCommandAndGetOutputE(t, runCmd)

		if expectSuccess {
			if cmdErr == nil && strings.TrimSpace(httpCode) == "200" {
				t.Log("Connectivity successful as expected.")
				return nil
			}
			lastErr = fmt.Errorf("expected successful connection (200 OK), but got code '%s' and error: %w", httpCode, cmdErr)
		} else {
			if cmdErr != nil || strings.TrimSpace(httpCode) != "200" {
				t.Logf("Connectivity failed as expected. Curl output: %s, Error: %v", httpCode, cmdErr)
				return nil
			}
			lastErr = fmt.Errorf("expected connection to fail, but it succeeded with code 200")
		}

		t.Logf("Verification failed on this attempt, retrying in 20 seconds...")
		time.Sleep(20 * time.Second)
	}
	return lastErr
}

func getCurrentGcloudUser(t *testing.T) string {
	cmd := shell.Command{Command: "gcloud", Args: []string{"auth", "list", "--filter=status:ACTIVE", "--format=value(account)"}}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	require.NoError(t, err, "Failed to get current gcloud user. Ensure gcloud is authenticated.")
	currentUser := strings.TrimSpace(output)
	require.NotEmpty(t, currentUser, "gcloud config get-value account returned empty string.")
	t.Logf("Current gcloud principal identified as: %s", currentUser)
	return currentUser
}

func addTokenCreatorRoleToPrincipal(t *testing.T, projectID, serviceAccountEmail, principal string) {
	t.Logf("Adding roles/iam.serviceAccountTokenCreator for principal '%s' on service account '%s'", principal, serviceAccountEmail)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "add-iam-policy-binding", serviceAccountEmail, "--project=" + projectID, "--member=" + principal, "--role=roles/iam.serviceAccountTokenCreator", "--format=none"},
	}
	shell.RunCommand(t, cmd)
}

func removeTokenCreatorRoleFromPrincipal(t *testing.T, projectID, serviceAccountEmail, principal string) {
	t.Logf("--- Removing roles/iam.serviceAccountTokenCreator for principal '%s' on service account '%s' ---", principal, serviceAccountEmail)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "remove-iam-policy-binding", serviceAccountEmail, "--project=" + projectID, "--member=" + principal, "--role=roles/iam.serviceAccountTokenCreator", "--format=none"},
	}
	shell.RunCommand(t, cmd)
}

func enableGcpApis(t *testing.T, projectID string, apis []string) {
	t.Logf("Enabling %d GCP APIs for project '%s'...", len(apis), projectID)
	for _, api := range apis {
		shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"services", "enable", api, "--project=" + projectID}})
	}
}

func addProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("Adding project role %s to %s", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"projects", "add-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--format=none"}}
		shell.RunCommand(t, cmd)
	}
}

func removeProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("--- Removing project role %s from %s ---", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"projects", "remove-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--format=none"}}
		shell.RunCommand(t, cmd)
	}
}

func addOrgIamBindings(t *testing.T, orgID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("Adding organization role %s to %s", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"organizations", "add-iam-policy-binding", orgID, "--member=" + member, "--role=" + role, "--format=none"}}
		shell.RunCommand(t, cmd)
	}
}

func removeOrgIamBindings(t *testing.T, orgID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("--- Removing organization role %s from %s ---", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"organizations", "remove-iam-policy-binding", orgID, "--member=" + member, "--role=" + role, "--format=none"}}
		shell.RunCommand(t, cmd)
	}
}

func getRegionFromZone(t *testing.T, zone string) string {
	lastHyphen := strings.LastIndex(zone, "-")
	if lastHyphen == -1 {
		t.Fatalf("Invalid zone format: %s. Expected format like 'us-central1-a'", zone)
	}
	return zone[:lastHyphen]
}

func createServiceAccount(t *testing.T, projectID, saName, displayName string) (string, error) {
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, projectID)
	cmd := shell.Command{Command: "gcloud", Args: []string{"iam", "service-accounts", "create", saName, "--project=" + projectID, "--display-name=" + displayName}}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return "", err
	}
	t.Logf("Service account %s created/verified.", saName)
	return saEmail, nil
}

func deleteServiceAccount(t *testing.T, projectID, saEmail string) {
	if saEmail == "" {
		return
	}
	t.Logf("--- Deleting service account: %s ---", saEmail)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"iam", "service-accounts", "delete", saEmail, "--project=" + projectID, "--quiet"}})
}

func createVPC(t *testing.T, projectID, networkName, zone string) {
	region := getRegionFromZone(t, zone)
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "create", networkName, "--project=" + projectID, "--subnet-mode=custom"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", subnetName, "--project=" + projectID, "--network=" + networkName, "--range=" + resourceVpcSubnetRange, "--region=" + region}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-allow-ssh-%s", networkName), "--project=" + projectID, "--network=" + networkName, "--allow=tcp:22", "--source-ranges=" + sshFirewallRange}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-allow-http-internal-%s", networkName), "--project=" + projectID, "--network=" + networkName, "--allow=tcp:80", "--source-ranges=" + resourceVpcSubnetRange}})
}

func deleteVPC(t *testing.T, projectID, networkName, zone string) {
	if networkName == "" {
		return
	}
	t.Logf("--- Deleting VPC: %s ---", networkName)
	region := getRegionFromZone(t, zone)
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", fmt.Sprintf("fw-allow-ssh-%s", networkName), "--project=" + projectID, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", fmt.Sprintf("fw-allow-http-internal-%s", networkName), "--project=" + projectID, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", fmt.Sprintf("%s-subnet", networkName), "--project=" + projectID, "--region=" + region, "--quiet"}})
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"}})
}

func createVM(t *testing.T, projectID, vmName, zone, networkName string) {
	t.Logf("Creating VM: %s in zone %s", vmName, zone)
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	startupScript := ""
	if strings.Contains(vmName, "server") {
		startupScript = "#!/bin/bash\nsudo apt-get update\nsudo apt-get install -y nginx\nsudo systemctl start nginx"
	} else if strings.Contains(vmName, "client") {
		startupScript = "#!/bin/bash\nsudo apt-get update\nsudo apt-get install -y curl"
	}
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "create", vmName,
		"--project=" + projectID,
		"--zone=" + zone,
		"--machine-type=e2-micro",
		"--subnet=" + subnetName,
		"--no-address",
		"--image-family=ubuntu-2204-lts", "--image-project=ubuntu-os-cloud",
		fmt.Sprintf("--metadata-from-file=startup-script=%s", createStartupScriptFile(t, startupScript)),
	}}
	shell.RunCommand(t, cmd)
}

func createStartupScriptFile(t *testing.T, scriptContent string) string {
	if scriptContent == "" {
		scriptContent = "#!/bin/bash\n# No startup script"
	}
	file, err := os.CreateTemp("", "startup-script-*.sh")
	assert.NoError(t, err)
	_, err = file.WriteString(scriptContent)
	assert.NoError(t, err)
	file.Close()
	t.Cleanup(func() { os.Remove(file.Name()) })
	return file.Name()
}

func deleteVM(t *testing.T, projectID, vmName, zone string) {
	if vmName == "" {
		return
	}
	shell.RunCommand(t, shell.Command{Command: "gcloud", Args: []string{"compute", "instances", "delete", vmName, "--project=" + projectID, "--zone=" + zone, "--quiet"}})
}
