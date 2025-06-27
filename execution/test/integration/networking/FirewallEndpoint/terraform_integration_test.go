// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "02-networking/FirewallEndpoint")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/networking/FirewallEndpoint/config")
	RequiredGcpApis        = []string{
		"iam.googleapis.com",
		"iamcredentials.googleapis.com",
		"compute.googleapis.com",
		"networksecurity.googleapis.com",
		"cloudresourcemanager.googleapis.com",
	}
	TestSaProjectRoles = []string{
		"roles/compute.admin",
		"roles/serviceusage.serviceUsageConsumer",
		"roles/networksecurity.firewallEndpointAdmin",
		"roles/compute.securityAdmin",
		"roles/compute.networkAdmin",
	}
	TestSaOrgRoles = []string{
		"roles/resourcemanager.organizationViewer",
		"roles/compute.orgFirewallPolicyAdmin",
		"roles/networksecurity.securityProfileAdmin",
		"roles/compute.networkAdmin",
	}
	inspectionVpcSubnetRange = "10.20.10.0/24"
	protectedVpcSubnetRange  = "10.30.10.0/24"
	sshFirewallRange         = "35.235.240.0/20"
	internalSrcRange         = "10.0.0.0/8"
)

func TestFirewallEndpointIntegration(t *testing.T) {
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
	serviceAccountName := fmt.Sprintf("sa-fe-test-%s", instanceSuffix)
	vpcInspectionName := fmt.Sprintf("vpc-inspection-fe-test-%s", instanceSuffix)
	vpcProtectedName := fmt.Sprintf("vpc-protected-fe-test-%s", instanceSuffix)
	zone := "us-central1-a"

	t.Logf("Test Run Config: ProjectID=%s, OrgID=%s, BillingProjectID=%s, Zone=%s, Suffix=%s", projectID, orgID, billingProjectID, zone, instanceSuffix)

	err = enableGcpApis(t, projectID, RequiredGcpApis)
	require.NoError(t, err)

	serviceAccountEmail, err := createServiceAccount(t, projectID, serviceAccountName, "Firewall Endpoint Test SA")
	require.NoError(t, err)
	defer deleteServiceAccount(t, projectID, serviceAccountEmail)
	time.Sleep(15 * time.Second)

	defer removeTokenCreatorRoleFromPrincipal(t, projectID, serviceAccountEmail, "user:"+currentUser)
	err = addTokenCreatorRoleToPrincipal(t, projectID, serviceAccountEmail, "user:"+currentUser)
	require.NoError(t, err)

	defer removeProjectIamBindings(t, projectID, serviceAccountEmail, TestSaProjectRoles)
	err = addProjectIamBindings(t, projectID, serviceAccountEmail, TestSaProjectRoles)
	require.NoError(t, err)

	defer removeOrgIamBindings(t, orgID, serviceAccountEmail, TestSaOrgRoles)
	err = addOrgIamBindings(t, orgID, serviceAccountEmail, TestSaOrgRoles)
	require.NoError(t, err)

	t.Log("Waiting for IAM permissions to propagate...")
	time.Sleep(60 * time.Second)

	endpointName := "fw-ep-integ-test-" + instanceSuffix
	assocName := "assoc-integ-test-" + instanceSuffix
	createConfigYAML(t, orgID, billingProjectID, projectID, vpcProtectedName, zone, endpointName, assocName)

	tfVars := map[string]interface{}{"config_folder_path": configFolderPath}
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
	err = createPeeredVPCs(t, projectID, vpcInspectionName, vpcProtectedName)
	require.NoError(t, err)
	defer deletePeeredVPCs(t, projectID, vpcInspectionName, vpcProtectedName)

	runConfigurationOnlyTest(t, terraformOptions)
}

func setQuotaProjectE(t *testing.T, projectID string) error {
	t.Logf("Setting gcloud billing/quota_project to: %s", projectID)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"config", "set", "billing/quota_project", projectID},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	return err
}

func unsetQuotaProject(t *testing.T) {
	t.Logf("Unsetting gcloud billing/quota_project.")
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"config", "unset", "billing/quota_project"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("WARN: Failed to unset quota project. Manual cleanup may be required. Error: %v", err)
	}
}

func runConfigurationOnlyTest(t *testing.T, terraformOptions *terraform.Options) {
	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")

	t.Log("Validating that Terraform has outputted a firewall endpoint ID...")
	outputJson := terraform.OutputJson(t, terraformOptions, "firewall_endpoints")
	require.True(t, gjson.Valid(outputJson), "Terraform output 'firewall_endpoints' is not valid JSON")

	var endpointId string
	results := gjson.Parse(outputJson)
	results.ForEach(func(key, value gjson.Result) bool {
		endpointId = value.Get("id").String()
		return false
	})
	assert.NotEmpty(t, endpointId, "Could not find 'id' in the 'firewall_endpoints' Terraform output")
	t.Logf("Validation successful: Found endpoint ID: %s", endpointId)

	t.Log("Validating control plane configuration...")
	err := verifyControlPlaneConfiguration(t, terraformOptions)
	assert.NoError(t, err, "Control plane configuration validation failed.")
	if err == nil {
		t.Log("Validation successful: Terraform resources were created.")
	}
}

func createConfigYAML(t *testing.T, orgID, billingProjectID, assocProjectID, vpcName, location, endpointName, assocName string) {
	type firewallEndpoint struct {
		Create           bool   `yaml:"create"`
		Name             string `yaml:"name"`
		OrganizationID   string `yaml:"organization_id"`
		BillingProjectID string `yaml:"billing_project_id"`
	}
	type firewallEndpointAssociation struct {
		Create               bool   `yaml:"create"`
		Name                 string `yaml:"name"`
		AssociationProjectID string `yaml:"association_project_id"`
		NetworkSelfLink      string `yaml:"vpc_id"`
	}
	type testConfig struct {
		Location string                      `yaml:"location"`
		Endpoint firewallEndpoint            `yaml:"firewall_endpoint"`
		Assoc    firewallEndpointAssociation `yaml:"firewall_endpoint_association"`
	}
	networkSelfLink := fmt.Sprintf("projects/%s/global/networks/%s", assocProjectID, vpcName)
	config := testConfig{
		Location: location,
		Endpoint: firewallEndpoint{Create: true, Name: endpointName, OrganizationID: orgID, BillingProjectID: billingProjectID},
		Assoc:    firewallEndpointAssociation{Create: true, Name: assocName, AssociationProjectID: assocProjectID, NetworkSelfLink: networkSelfLink},
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

func createPeeredVPCs(t *testing.T, projectID, inspectionVPC, protectedVPC string) error {
	t.Logf("Creating Inspection VPC '%s' and Protected VPC '%s'", inspectionVPC, protectedVPC)
	inspectionURI := fmt.Sprintf("projects/%s/global/networks/%s", projectID, inspectionVPC)
	protectedURI := fmt.Sprintf("projects/%s/global/networks/%s", projectID, protectedVPC)

	region := "us-central1"
	commands := []shell.Command{
		{Command: "gcloud", Args: []string{"compute", "networks", "create", inspectionVPC, "--project=" + projectID, "--subnet-mode=custom"}},
		{Command: "gcloud", Args: []string{"compute", "networks", "create", protectedVPC, "--project=" + projectID, "--subnet-mode=custom"}},
		{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", fmt.Sprintf("%s-subnet", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--range=" + inspectionVpcSubnetRange, "--region=" + region}},
		{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", fmt.Sprintf("%s-subnet", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--range=" + protectedVpcSubnetRange, "--region=" + region}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-all", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--allow=all", "--source-ranges=" + internalSrcRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-all", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--allow=all", "--source-ranges=1" + internalSrcRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-ssh", inspectionVPC), "--project=" + projectID, "--network=" + inspectionVPC, "--allow=tcp:22", "--source-ranges=" + sshFirewallRange}},
		{Command: "gcloud", Args: []string{"compute", "firewall-rules", "create", fmt.Sprintf("fw-%s-allow-ssh", protectedVPC), "--project=" + projectID, "--network=" + protectedVPC, "--allow=tcp:22", "--source-ranges=" + sshFirewallRange}},
		{Command: "gcloud", Args: []string{"compute", "networks", "peerings", "create", fmt.Sprintf("peering-to-%s", protectedVPC), "--network=" + inspectionVPC, "--peer-network=" + protectedURI, "--project=" + projectID, "--export-custom-routes", "--import-custom-routes"}},
		{Command: "gcloud", Args: []string{"compute", "networks", "peerings", "create", fmt.Sprintf("peering-to-%s", inspectionVPC), "--network=" + protectedVPC, "--peer-network=" + inspectionURI, "--project=" + projectID, "--export-custom-routes", "--import-custom-routes"}},
	}

	for _, cmd := range commands {
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			t.Errorf("failed to run gcloud command %s", err)
		}
	}
	return nil
}

func deletePeeredVPCs(t *testing.T, projectID, inspectionVPC, protectedVPC string) {
	if inspectionVPC == "" || protectedVPC == "" {
		return
	}
	t.Logf("--- Deleting Peered VPCs and their dependent resources: %s, %s ---", inspectionVPC, protectedVPC)
	rulesToDelete := []string{fmt.Sprintf("fw-%s-allow-all", inspectionVPC), fmt.Sprintf("fw-%s-allow-ssh", inspectionVPC), fmt.Sprintf("fw-%s-allow-all", protectedVPC), fmt.Sprintf("fw-%s-allow-ssh", protectedVPC)}
	subnetsToDelete := []string{fmt.Sprintf("%s-subnet", inspectionVPC), fmt.Sprintf("%s-subnet", protectedVPC)}
	for _, ruleName := range rulesToDelete {
		cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "firewall-rules", "delete", ruleName, "--project=" + projectID, "--quiet"}}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			t.Errorf("WARN: Failed to delete firewall rule %s. Manual cleanup may be required. Error: %v", ruleName, err)
		}
	}
	for _, subnetName := range subnetsToDelete {
		region := "us-central1"
		cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", subnetName, "--project=" + projectID, "--region=" + region, "--quiet"}}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			t.Errorf("WARN: Failed to delete subnet %s. Manual cleanup may be required. Error: %v", subnetName, err)
		}
	}
	cmdinspection := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", inspectionVPC, "--project=" + projectID, "--quiet"}}
	if _, err := shell.RunCommandAndGetOutputE(t, cmdinspection); err != nil {
		t.Errorf("WARN: Failed to delete VPC %s. Manual cleanup may be required. Error: %v", inspectionVPC, err)
	}
	cmdProtected := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", protectedVPC, "--project=" + projectID, "--quiet"}}
	if _, err := shell.RunCommandAndGetOutputE(t, cmdProtected); err != nil {
		t.Errorf("WARN: Failed to delete VPC %s. Manual cleanup may be required. Error: %v", protectedVPC, err)
	}
}

func verifyControlPlaneConfiguration(t *testing.T, terraformOptions *terraform.Options) error {
	outputJson := terraform.OutputJson(t, terraformOptions, "firewall_endpoints")
	require.True(t, gjson.Valid(outputJson), "Terraform output 'firewall_endpoints' is not valid JSON")

	projectID := os.Getenv("TF_VAR_project_id")
	cmd := shell.Command{Command: "gcloud", Args: []string{"compute", "routes", "list", "--project=" + projectID, "--format=json"}}

	t.Logf("Verifying that an auto-generated peering route exists...")

	var lastErr error
	const maxRetries = 10
	const sleepBetweenRetries = 30 * time.Second

	for i := 0; i < maxRetries; i++ {
		routesJson, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			lastErr = fmt.Errorf("failed to list routes with gcloud: %w", err)
			t.Logf("Attempt %d/%d: Failed to list routes, retrying in %v...", i+1, maxRetries, sleepBetweenRetries)
			time.Sleep(sleepBetweenRetries)
			continue
		}
		parsedRoutes := gjson.Parse(routesJson)
		routeFound := false
		for _, route := range parsedRoutes.Array() {
			description := route.Get("description").String()
			if strings.Contains(description, "Auto generated route via peering") {
				t.Logf("Validation successful on attempt %d/%d. Found peering route '%s'.", i+1, maxRetries, route.Get("name").String())
				routeFound = true
				break
			}
		}
		if routeFound {
			return nil
		}

		lastErr = fmt.Errorf("could not find a route with a description indicating it was auto-generated by peering")
		t.Logf("Attempt %d/%d: Required peering route not found, retrying in %v...", i+1, maxRetries, sleepBetweenRetries)
		time.Sleep(sleepBetweenRetries)
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

func addTokenCreatorRoleToPrincipal(t *testing.T, projectID, serviceAccountEmail, principal string) error {
	t.Logf("Adding roles/iam.serviceAccountTokenCreator for principal '%s' on service account '%s'", principal, serviceAccountEmail)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "add-iam-policy-binding", serviceAccountEmail, "--project=" + projectID, "--member=" + principal, "--role=roles/iam.serviceAccountTokenCreator", "--format=none"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	return err
}

func removeTokenCreatorRoleFromPrincipal(t *testing.T, projectID, serviceAccountEmail, principal string) {
	t.Logf("--- Removing roles/iam.serviceAccountTokenCreator for principal '%s' on service account '%s' ---", principal, serviceAccountEmail)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "remove-iam-policy-binding", serviceAccountEmail, "--project=" + projectID, "--member=" + principal, "--role=roles/iam.serviceAccountTokenCreator", "--format=none"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("WARN: Failed to remove Token Creator role. Manual cleanup may be required. Error: %v", err)
	}
}

func enableGcpApis(t *testing.T, projectID string, apis []string) error {
	t.Logf("Enabling %d GCP APIs for project '%s'...", len(apis), projectID)
	for _, api := range apis {
		cmd := shell.Command{Command: "gcloud", Args: []string{"services", "enable", api, "--project=" + projectID}}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			return fmt.Errorf("failed to enable api %s: %w", api, err)
		}
	}
	return nil
}

func addProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) error {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("Adding project role %s to %s", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"projects", "add-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--format=none"}}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			return fmt.Errorf("failed to add project role %s: %w", role, err)
		}
	}
	return nil
}

func removeProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("--- Removing project role %s from %s ---", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"projects", "remove-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--format=none"}}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Errorf("WARN: Failed to remove project role %s. Manual cleanup may be required. Error: %v", role, err)
		}
	}
}

func addOrgIamBindings(t *testing.T, orgID string, serviceAccountEmail string, roles []string) error {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("Adding organization role %s to %s", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"organizations", "add-iam-policy-binding", orgID, "--member=" + member, "--role=" + role, "--format=none"}}
		if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
			return fmt.Errorf("failed to add org role %s: %w", role, err)
		}
	}
	return nil
}

func removeOrgIamBindings(t *testing.T, orgID string, serviceAccountEmail string, roles []string) {
	member := "serviceAccount:" + serviceAccountEmail
	for _, role := range roles {
		t.Logf("--- Removing organization role %s from %s ---", role, member)
		cmd := shell.Command{Command: "gcloud", Args: []string{"organizations", "remove-iam-policy-binding", orgID, "--member=" + member, "--role=" + role, "--format=none"}}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Errorf("WARN: Failed to remove organization role %s. Manual cleanup may be required. Error: %v", role, err)
		}
	}
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
	cmd := shell.Command{Command: "gcloud", Args: []string{"iam", "service-accounts", "delete", saEmail, "--project=" + projectID, "--quiet"}}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("WARN: Failed to delete service account %s. Manual cleanup may be required. Error: %v", saEmail, err)
	}
}
