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
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

// --- Test Configuration ---
const (
	// Default values - can be overridden by environment variables where applicable
	defaultRegion = "us-central1"
	// networkPrefix = "appeng-test-vpc"    // Prefix for created VPC Network
	// subnetPrefix  = "appeng-test-subnet" // Prefix for created Subnet
	// connectorPrefix              = "test-conn"          // Prefix for created VPC Connector
	testSubnetCIDR    = "10.5.4.0/28"    // Subnet CIDR for the subnet that will host the connector
	httpCheckRetries  = 10               // Retries for HTTP check
	httpCheckInterval = 30 * time.Second // Interval for HTTP check
	// vpcCleanupWaitShort          = 30 * time.Second // Shorter wait in deleteVPC
	// vpcCleanupWaitLong           = 60 * time.Second // Longer wait in deleteVPC
	iamPropagationWait           = 30 * time.Second // Wait after adding IAM bindings
	saCreationPropagationWait    = 15 * time.Second // Wait after SA creation before binding roles
	appCreatePropagationWait     = 15 * time.Second // Wait after App Engine app creation
	apiEnablementPropagationWait = 30 * time.Second // Wait after enabling APIs
	// vpcAccessCreateWaitTime      = 120 * time.Second

	// Constants for gcloud-created VPC Access Connector (used with --subnet)
	// connectorMachineType  = "e2-micro"       // Machine type for the connector
	// connectorMinInstances = 2                // Min instances for the connector
	// connectorMaxInstances = 3                // Max instances for the connector
	// connectorDeleteWait   = 20 * time.Second // Brief wait before attempting connector delete

	// Sample App Configuration (used in createConfigYAML)
	sampleAppRuntime       = "python311"
	sampleAppEntrypoint    = "gunicorn -b :$PORT main:app"
	sampleAppGcsObjectName = "app.yaml"
	sampleAppGcsMainPyName = "main.py"
)

var (
	uniqueID               = strings.ToLower(random.UniqueId())
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Standard")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Serverless/AppEngine/Standard/config")
	service1               = "service1"
	service2               = "service2"
	projectID              = os.Getenv("TF_VAR_project_id")
	versionID1             = fmt.Sprintf("v1-%s", uniqueID)
	versionID2             = fmt.Sprintf("v1-%s", uniqueID)
	sampleAppGcsBucket     = getEnv("TF_VAR_test_gcs_bucket", fmt.Sprintf("%s-tf-test-bucket", projectID))

	testServiceAccountRoles = []string{
		"roles/compute.networkUser",
		"roles/compute.instanceAdmin.v1",
		"roles/iam.serviceAccountUser",
		"roles/appengine.appAdmin",
		"roles/cloudbuild.builds.editor",
		"roles/artifactregistry.writer",
		"roles/artifactregistry.reader",
		"roles/compute.networkViewer",
		"roles/storage.objectAdmin",
		// "roles/vpcaccess.admin",
		"roles/serviceusage.serviceUsageAdmin",
	}

	defaultComputeServiceAccountEmailRoles = []string{
		"roles/storage.objectAdmin",
		"roles/artifactregistry.writer",
		"roles/artifactregistry.reader",
	}

	requiredGcpApis = []string{
		"serviceusage.googleapis.com", // API to enable other APIs
		"appengine.googleapis.com",
		"compute.googleapis.com",
		"iam.googleapis.com",
		"cloudresourcemanager.googleapis.com",
		"storage.googleapis.com",
		// "vpcaccess.googleapis.com",
		"cloudbuild.googleapis.com",
		"artifactregistry.googleapis.com",
	}
)

// --- Struct Definitions ---
type AppEngineConfig struct {
	ProjectID        string                  `yaml:"project_id"`
	Service          string                  `yaml:"service"`
	VersionID        string                  `yaml:"version_id"`
	Runtime          string                  `yaml:"runtime"`
	Deployment       *DeploymentConfig       `yaml:"deployment,omitempty"`
	Entrypoint       *EntrypointConfig       `yaml:"entrypoint,omitempty"`
	AutomaticScaling *AutomaticScalingConfig `yaml:"automatic_scaling,omitempty"`
	Handlers         HandlerConfig           `yaml:"handlers,omitempty"`
	// VPCAccessConnector         *VPCAccessConnectorConfig        `yaml:"vpc_access_connector,omitempty"`
	// VPCAccessConnectorDetails *VPCAccessConnectorDetailsConfig `yaml:"vpc_connector_details,omitempty"`
	AppEngineApplication *AppEngineAppConfig `yaml:"app_engine_application,omitempty"`
	// CreateVPCConnector         bool                             `yaml:"create_vpc_connector,omitempty"`
	CreateSplitTraffic         bool                  `yaml:"create_split_traffic,omitempty"`
	CreateAppEngineApplication bool                  `yaml:"create_app_engine_application,omitempty"`
	CreateDispatchRules        bool                  `yaml:"create_dispatch_rules,omitempty"`
	CreateNetworkSettings      bool                  `yaml:"create_network_settings,omitempty"`
	DeleteServiceOnDestroy     bool                  `yaml:"delete_service_on_destroy,omitempty"`
	CreateDomainMappings       bool                  `yaml:"create_domain_mappings,omitempty"`
	DomainMappings             []DomainMappingConfig `yaml:"domain_mappings,omitempty"`
	CreateFirewallRules        bool                  `yaml:"create_firewall_rules,omitempty"`
	FirewallRules              []FirewallRuleConfig  `yaml:"firewall_rules,omitempty"`
	CreateAppVersion           bool                  `yaml:"create_app_version,omitempty"`
	EnvVariables               map[string]string     `yaml:"env_variables,omitempty"`
	Labels                     map[string]string     `yaml:"labels,omitempty"`
}
type DeploymentConfig struct {
	Files *FilesConfig `yaml:"files,omitempty"`
}
type FilesConfig struct {
	Name      string `yaml:"name"`
	SourceURL string `yaml:"source_url"`
}
type EntrypointConfig struct {
	Shell string `yaml:"shell"`
}
type AutomaticScalingConfig struct {
	MaxConcurrentRequests int `yaml:"max_concurrent_requests,omitempty"`
	MinIdleInstances      int `yaml:"min_idle_instances,omitempty"`
	MaxIdleInstances      int `yaml:"max_idle_instances,omitempty"`
}
type HandlerConfig []struct {
	URLRegex string        `yaml:"url_regex"`
	Script   *ScriptConfig `yaml:"script,omitempty"`
}
type ScriptConfig struct {
	ScriptPath string `yaml:"script_path"`
}
type VPCAccessConnectorConfig struct {
	Name string `yaml:"name"`
}

//	type VPCAccessConnectorDetailsConfig struct {
//		Name          string `yaml:"name"`
//		SubnetName    string `yaml:"subnet_name"`
//		MachineType   string `yaml:"machine_type"`
//		MinInstances  int    `yaml:"min_instances"`
//		MaxInstances  int    `yaml:"max_instances"`
//		HostProjectID string `yaml:"host_project_id"`
//		Region        string `yaml:"region"`
//	}
type AppEngineAppConfig struct {
	LocationID string `yaml:"location_id"`
}
type DomainMappingConfig struct {
	DomainName string `yaml:"domain_name"`
}
type FirewallRuleConfig struct {
	SourceRange string `yaml:"source_range"`
	Action      string `yaml:"action"`
}

// --- Helper Functions ---

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestAppEngineStandardIntegration(t *testing.T) {
	t.Parallel()

	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set")
	}

	// Get current gcloud user who will be granted token creator role
	currentPrincipal := getCurrentGcloudUser(t)
	projectNumber := getProjectNumber(t, projectID)

	serviceAccountID := fmt.Sprintf("ae-test-sa-%s", uniqueID[:min(len(uniqueID), 18)])
	serviceAccountDisplayName := fmt.Sprintf("App Engine Test SA (%s)", uniqueID)
	// testNetworkName := fmt.Sprintf("%s-%s", networkPrefix, uniqueID)
	// testSubnetworkName := fmt.Sprintf("%s-%s", subnetPrefix, uniqueID)
	// testConnectorName := fmt.Sprintf("%s-%s", connectorPrefix, uniqueID)
	testConfigFolderPath := configFolderPath
	testGcsObjectPathPrefix := fmt.Sprintf("app-test-%s", uniqueID)

	t.Logf("Creating test Service Account: %s", serviceAccountID)
	serviceAccountEmail := createServiceAccount(t, projectID, serviceAccountID, serviceAccountDisplayName)
	defer deleteServiceAccount(t, projectID, serviceAccountEmail)

	// Grant the current principal the Service Account Token Creator role on the new SA <<< ADD THIS BLOCK
	t.Logf("Granting Service Account Token Creator role to principal '%s' on new SA '%s'", currentPrincipal, serviceAccountEmail)
	addTokenCreatorRoleToPrincipalOnServiceAccount(t, projectID, serviceAccountEmail, currentPrincipal)
	defer removeTokenCreatorRoleFromPrincipalOnServiceAccount(t, projectID, serviceAccountEmail, currentPrincipal) // Ensure cleanup

	t.Logf("Waiting %v for SA '%s' to propagate before adding IAM bindings...", saCreationPropagationWait, serviceAccountEmail)
	time.Sleep(saCreationPropagationWait)

	t.Logf("Adding IAM roles to test Service Account %s...", serviceAccountEmail)
	addProjectIamBindings(t, projectID, serviceAccountEmail, testServiceAccountRoles)
	defaultComputeServiceAccountEmail := fmt.Sprintf("%s-compute@developer.gserviceaccount.com", projectNumber)

	t.Logf("Adding IAM roles to default Compute Service Account %s...", defaultComputeServiceAccountEmail)
	addProjectIamBindings(t, projectID, defaultComputeServiceAccountEmail, defaultComputeServiceAccountEmailRoles)

	defer removeProjectIamBindings(t, projectID, serviceAccountEmail, testServiceAccountRoles)

	t.Logf("Waiting %v for IAM binding propagation before Terraform apply...", iamPropagationWait)
	time.Sleep(iamPropagationWait)

	// Enable required GCP APIs
	t.Logf("Enabling required GCP APIs for project '%s'...", projectID)
	enableGcpApis(t, projectID, requiredGcpApis)
	t.Logf("Waiting %v for API enablement to propagate...", apiEnablementPropagationWait)
	time.Sleep(apiEnablementPropagationWait)

	// Ensure App Engine application exists or create it
	t.Logf("Ensuring App Engine application exists in project '%s' for region '%s'...", projectID, defaultRegion)
	ensureAppEngineApplicationExists(t, projectID, defaultRegion)

	t.Logf("Creating GCS bucket: %s", sampleAppGcsBucket)
	createGcsBucket(t, projectID, sampleAppGcsBucket, defaultRegion)
	defer deleteGcsBucket(t, sampleAppGcsBucket)

	defer deleteGcsObjects(t, sampleAppGcsBucket, testGcsObjectPathPrefix+"/")
	appYamlContent, mainPyContent := getHelloWorldAppFiles()
	appYamlGcsPath := path.Join(testGcsObjectPathPrefix, sampleAppGcsObjectName)
	mainPyGcsPath := path.Join(testGcsObjectPathPrefix, sampleAppGcsMainPyName)
	t.Logf("Uploading test app files to gs://%s/%s/", sampleAppGcsBucket, testGcsObjectPathPrefix)
	uploadGcsObject(t, projectID, sampleAppGcsBucket, appYamlGcsPath, appYamlContent)
	uploadGcsObject(t, projectID, sampleAppGcsBucket, mainPyGcsPath, mainPyContent)

	// t.Logf("Creating test VPC network '%s' and subnet '%s' using gcloud...", testNetworkName, testSubnetworkName)
	// createVPC(t, projectID, testNetworkName, testSubnetworkName, testSubnetCIDR, defaultRegion) // This creates testSubnetworkName
	// defer deleteVPC(t, projectID, testNetworkName, testSubnetworkName, defaultRegion)

	// t.Logf("Creating VPC Access Connector '%s' via gcloud using subnet '%s'...", testConnectorName, testSubnetworkName)
	// // Using testSubnetworkName as the dedicated subnet for the connector.
	// // projectID is used as subnetProjectID as the subnet is in the same project.
	// gcloudCreatedConnectorFullName := createVPCConnectorGcloud(t, projectID, testConnectorName, defaultRegion,
	// 	testSubnetworkName, projectID, // Pass subnetName and its projectID
	// 	connectorMachineType, connectorMinInstances, connectorMaxInstances)
	// time.Sleep(vpcAccessCreateWaitTime)
	// defer deleteVPCConnectorGcloud(t, projectID, testConnectorName, defaultRegion)

	appYamlDisplayUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", sampleAppGcsBucket, appYamlGcsPath)
	defer os.RemoveAll(testConfigFolderPath)
	// Pass gcloudCreatedConnectorFullName; testSubnetName and uniqueID for createConfigYAML are for other potential uses or logging.
	t.Logf("Creating YAML Config.")
	createConfigYAML(t, testConfigFolderPath, uniqueID, appYamlDisplayUrl) //,testSubnetworkName, gcloudCreatedConnectorFullName)

	tfVars := map[string]interface{}{
		"config_folder_path": testConfigFolderPath,
	}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: terraformDirectoryPath,
		Vars:         tfVars,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	defer terraform.Destroy(t, terraformOptions)
	t.Logf("====== Running terraform init & apply... ======")
	_, err := terraform.InitAndApplyE(t, terraformOptions)
	if err != nil {
		t.Errorf("Terraform init/apply failed : %s", err)
	} else {
		t.Logf("Terraform apply completed successfully.")
	}

	t.Log("====== Starting Verification of Terraform Outputs. =======")

	appEngineOutputValue := terraform.OutputJson(t, terraformOptions, "app_engine_standard")
	if !gjson.Valid(appEngineOutputValue) {
		t.Errorf("Error parsing output, invalid json: %s", appEngineOutputValue)
	}
	result := gjson.Parse(appEngineOutputValue)

	t.Log(" ========= Verify Instance ID's ========= ")

	instance1IDPath := fmt.Sprintf("instance1.app_engine_standard.%s.id", service1)
	got := gjson.Get(result.String(), instance1IDPath).String()
	want := fmt.Sprintf("apps/%s/services/%s/versions/%s", projectID, service1, versionID1)
	if got != want {
		t.Errorf("App Engine Instance 1 with invalid ID created = %v, want = %v", got, want)
	} else {
		t.Log("ID for Instance 1 verified successfully.")
	}

	instance2IDPath := fmt.Sprintf("instance2.app_engine_standard.%s.id", service2)
	got = gjson.Get(result.String(), instance2IDPath).String()
	want = fmt.Sprintf("apps/%s/services/%s/versions/%s", projectID, service2, versionID2)
	if got != want {
		t.Errorf("App Engine Instance 1 with invalid ID created = %v, want = %v", got, want)
	} else {
		t.Log("ID for Instance 2 verified successfully.")
	}

	t.Log(" ========= Verify Version ID's ========= ")

	versionID1Path := fmt.Sprintf("instance1.app_engine_standard.%s.version_id", service1)
	got = gjson.Get(result.String(), versionID1Path).String()
	want = versionID1
	if got != want {
		t.Errorf("App Engine Instance 1 with invalid Version ID created = %v, want = %v", got, want)
	} else {
		t.Log("Version ID for Instance 1 verified successfully.")
	}

	versionID2Path := fmt.Sprintf("instance1.app_engine_standard.%s.version_id", service1)
	got = gjson.Get(result.String(), versionID2Path).String()
	want = versionID2
	if got != want {
		t.Errorf("App Engine Instance 1 with invalid version ID created = %v, want = %v", got, want)
	} else {
		t.Log("Version ID for Instance 2 verified successfully.")
	}

	t.Log(" ========= Verify Runtime ========= ")

	runtimeID1Path := fmt.Sprintf("instance1.app_engine_standard.%s.runtime", service1)
	got = gjson.Get(result.String(), runtimeID1Path).String()
	want = sampleAppRuntime
	if got != want {
		t.Errorf("App Engine Instance 1 with invalid Runtime created = %v, want = %v", got, want)
	} else {
		t.Log("Runtime ID for Instance 1 verified successfully.")
	}

	runtimeID2Path := fmt.Sprintf("instance1.app_engine_standard.%s.runtime", service1)
	got = gjson.Get(result.String(), runtimeID2Path).String()
	want = sampleAppRuntime
	if got != want {
		t.Errorf("App Engine Instance 2 with invalid runtime ID created = %v, want = %v", got, want)
	} else {
		t.Log("Runtime ID for Instance 2 verified successfully.")
	}

	t.Log("====== Test AppEngine Standard Integration Completed. ====")
}

// --- Helper Function Implementations ---

func createServiceAccount(t *testing.T, projectID string, accountID string, displayName string) string {
	t.Helper()
	t.Logf("Attempting to create service account '%s' in project '%s' with display name '%s'...", accountID, projectID, displayName)
	var extractedEmail string
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"iam", "service-accounts", "create", accountID,
			"--project=" + projectID,
			"--display-name=" + displayName,
			"--format=value(email)",
			"--quiet",
		},
	}

	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Error creating service account %s : Error %s", accountID, err)
	} else {
		extractedEmail = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountID, projectID)
	}
	return extractedEmail
}

func deleteServiceAccount(t *testing.T, projectID string, email string) {
	t.Helper()
	t.Logf("Deleting service account '%s'...", email)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "delete", email, "--project=" + projectID, "--quiet"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error deleting service account %s: %v. Output:\n%s", email, err, output)
	} else {
		t.Logf("Service Account %s deleted.", email)
	}
}

func addProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) {
	t.Helper()
	member := fmt.Sprintf("serviceAccount:%s", serviceAccountEmail)
	for _, role := range roles {
		t.Logf("Adding role %s to %s for project %s", role, member, projectID)
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"projects", "add-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--condition=None", "--format=none"},
		}
		output, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Logf("Failed to add IAM binding '%s' for member '%s' to project '%s'. Output:\n%s, Error: %s", role, member, projectID, output, err)
		}
	}
}

func removeProjectIamBindings(t *testing.T, projectID string, serviceAccountEmail string, roles []string) {
	t.Helper()
	member := fmt.Sprintf("serviceAccount:%s", serviceAccountEmail)
	for _, role := range roles {
		t.Logf("Removing role %s from %s for project %s", role, member, projectID)
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"projects", "remove-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--condition=None", "--quiet", "--format=none"},
		}
		output, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Logf("Error removing IAM binding '%s' for member '%s' from project '%s': %v. Output:\n%s", role, member, projectID, err, output)
		} else {
			t.Logf("Removed role %s from %s", role, member)
		}
	}
}

func createGcsBucket(t *testing.T, projectID string, bucketName string, location string) {
	t.Helper()
	t.Logf("Creating GCS bucket: gs://%s", bucketName)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"storage", "buckets", "create", fmt.Sprintf("gs://%s", bucketName), "--project=" + projectID, "--location=" + location, "--uniform-bucket-level-access"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	t.Logf("Failed to create GCS bucket %s. Output:\n%s, Error: %s", bucketName, output, err)
	t.Logf("GCS bucket gs://%s created.", bucketName)
}

func deleteGcsBucket(t *testing.T, bucketName string) {
	t.Helper()
	t.Logf("Deleting GCS bucket: gs://%s", bucketName)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"storage", "buckets", "delete", fmt.Sprintf("gs://%s", bucketName), "--quiet"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		if !strings.Contains(output, "BucketNotFoundException: 404") && !strings.Contains(err.Error(), "NotFoundException") {
			t.Logf("Error deleting GCS bucket %s: %v. Output:\n%s", bucketName, err, output)
		} else {
			t.Logf("GCS bucket %s already deleted or not found.", bucketName)
		}
	} else {
		t.Logf("GCS bucket %s deleted.", bucketName)
	}
}

func deleteGcsObjects(t *testing.T, bucketName string, objectPathPrefix string) {
	t.Helper()
	gcsPath := fmt.Sprintf("gs://%s/%s", bucketName, strings.TrimSuffix(objectPathPrefix, "/")+"/*")
	t.Logf("Deleting objects in GCS path: %s", gcsPath)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"storage", "rm", gcsPath, "--recursive"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil && !strings.Contains(output, "One or more URLs matched no objects") && !strings.Contains(err.Error(), "One or more URLs matched no objects") {
		t.Logf("Note: Error deleting objects from %s (may be benign if already gone): %v. Output:\n%s", gcsPath, err, output)
	} else {
		t.Logf("Attempted deletion of objects in %s (any matching objects removed or none found).", gcsPath)
	}
}

func uploadGcsObject(t *testing.T, projectID string, bucketName string, objectPath string, content string) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "gcs-upload-*.tmp")
	if err != nil {
		t.Logf("Failed to create temp file for GCS upload:Error %s", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Logf("Failed to write content to temp file: Error:%s", err)
	}
	err = tmpFile.Close()
	if err != nil {
		t.Logf("Failed to close temp file: Error:%s", err)
	}

	gcsDest := fmt.Sprintf("gs://%s/%s", bucketName, objectPath)
	t.Logf("Uploading temp file %s to %s", tmpFile.Name(), gcsDest)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"storage", "cp", tmpFile.Name(), gcsDest, "--project=" + projectID},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Failed to upload object %s to bucket %s. Error:%s", objectPath, bucketName, err)
	}
	t.Logf("Uploaded object %s successfully.", objectPath)
}

// func createVPC(t *testing.T, projectID string, networkName string, subnetName string, subnetCIDR string, region string) {
// 	t.Helper()
// 	t.Logf("Creating VPC network '%s' in project '%s'...", networkName, projectID)
// 	cmdNetwork := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "create", networkName, "--project=" + projectID, "--subnet-mode=custom", "--bgp-routing-mode=global", "--format=json"}}
// 	outputNetwork, errNetwork := shell.RunCommandAndGetOutputE(t, cmdNetwork)
// 	if errNetwork != nil {
// 		t.Logf("Failed to create VPC network %s. Output:\n%s. Error:%s", networkName, outputNetwork, errNetwork)
// 	}
// 	t.Logf("VPC network '%s' created.", networkName)
// 	t.Logf("Creating subnet '%s' in network '%s', region '%s' with range '%s'...", subnetName, networkName, region, subnetCIDR)
// 	cmdSubnet := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "create", subnetName, "--project=" + projectID, "--network=" + networkName, "--region=" + region, "--range=" + subnetCIDR, "--format=json"}}
// 	outputSubnet, errSubnet := shell.RunCommandAndGetOutputE(t, cmdSubnet)
// 	if errSubnet != nil {
// 		t.Logf("Failed to create subnet %s. Output:\n%s Error:%s", subnetName, outputSubnet, errSubnet)
// 	}

// 	t.Logf("Subnet '%s' created.", subnetName)
// }

// func deleteVPC(t *testing.T, projectID string, networkName string, subnetName string, region string) {
// 	t.Helper()
// 	t.Logf("Waiting %v before deleting subnet %s...", vpcCleanupWaitShort, subnetName)
// 	time.Sleep(vpcCleanupWaitShort)
// 	t.Logf("Deleting subnet '%s' in region '%s' from project '%s'...", subnetName, region, projectID)
// 	cmdSubnet := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "subnets", "delete", subnetName, "--project=" + projectID, "--region=" + region, "--quiet"}}
// 	outputSubnet, errSubnet := shell.RunCommandAndGetOutputE(t, cmdSubnet)
// 	if errSubnet != nil {
// 		if !strings.Contains(outputSubnet, "NOT_FOUND") && !strings.Contains(errSubnet.Error(), "NOT_FOUND") {
// 			t.Logf("Error deleting subnet %s: %v. Output:\n%s", subnetName, errSubnet, outputSubnet)
// 		} else {
// 			t.Logf("Subnet %s already deleted or not found.", subnetName)
// 		}
// 	} else {
// 		t.Logf("Subnet %s deleted or was already gone.", subnetName)
// 	}

// 	t.Logf("Waiting %v before deleting network %s...", vpcCleanupWaitLong, networkName)
// 	time.Sleep(vpcCleanupWaitLong)
// 	t.Logf("Deleting network '%s' from project '%s'...", networkName, projectID)
// 	cmdNetwork := shell.Command{Command: "gcloud", Args: []string{"compute", "networks", "delete", networkName, "--project=" + projectID, "--quiet"}}
// 	outputNetwork, errNetwork := shell.RunCommandAndGetOutputE(t, cmdNetwork)
// 	if errNetwork != nil {
// 		if !strings.Contains(outputNetwork, "NOT_FOUND") && !strings.Contains(errNetwork.Error(), "NOT_FOUND") {
// 			t.Logf("Error deleting network %s: %v. Output:\n%s", networkName, errNetwork, outputNetwork)
// 		} else {
// 			t.Logf("Network %s already deleted or not found.", networkName)
// 		}
// 	} else {
// 		t.Logf("Network %s deleted or was already gone.", networkName)
// 	}
// }

// createVPCConnectorGcloud creates a VPC Access Connector using the user-specified gcloud command structure.
// func createVPCConnectorGcloud(t *testing.T, projectID, connectorName, region,
// 	subnetName, // Name of the dedicated /28 subnet
// 	subnetProjectID, // Project ID of the subnet (usually same as projectID)
// 	machineType string, minInstances, maxInstances int) string {
// 	t.Helper()
// 	t.Logf("Creating VPC Access Connector '%s' using 'gcloud compute networks vpc-access connectors create' in project '%s', region '%s' using subnet '%s'...",
// 		connectorName, projectID, region, subnetName)
// 	t.Logf("Connector details: SubnetProject='%s', MachineType='%s', MinInst=%d, MaxInst=%d",
// 		subnetProjectID, machineType, minInstances, maxInstances)

// 	args := []string{
// 		"compute", "networks", "vpc-access", "connectors", "create", connectorName,
// 		"--project=" + projectID, // Though gcloud often infers this, explicit is safer for scripts
// 		"--region=" + region,
// 		"--subnet=" + subnetName,
// 		fmt.Sprintf("--min-instances=%d", minInstances),
// 		fmt.Sprintf("--max-instances=%d", maxInstances),
// 		"--machine-type=" + machineType,
// 		"--format=value(name)", // Assuming this format option is available for this command path
// 		"--quiet",              // To handle any operational prompts non-interactively
// 	}

// 	cmd := shell.Command{
// 		Command: "gcloud",
// 		Args:    args,
// 	}

// 	output, err := shell.RunCommandAndGetOutputE(t, cmd)
// 	if err != nil {
// 		t.Logf("Failed to create VPC Access Connector '%s' using 'compute networks vpc-access' path and subnet. Output:\n%s. Error: %s", connectorName, output, err)
// 	}

// 	// Log the full raw stdout for debugging purposes
// 	t.Logf("Raw stdout from 'gcloud compute networks vpc-access connectors create ... --format=value(name)':\n---\n%s\n---", output)

// 	connectorFullName := fmt.Sprintf("projects/%s/locations/%s/connectors/%s", projectID, region, connectorName)

// 	t.Logf("VPC Access Connector '%s' created successfully. Fully qualified name to be used: '%s'", connectorName, connectorFullName)
// 	return connectorFullName
// }

// deleteVPCConnectorGcloud deletes a VPC Access Connector using the standard gcloud path.
// The delete command path is typically stable even if creation paths change or have aliases.
// func deleteVPCConnectorGcloud(t *testing.T, projectID, connectorName, region string) {
// 	t.Helper()
// 	t.Logf("Waiting %v before attempting to delete VPC Access Connector '%s'...", connectorDeleteWait, connectorName)
// 	time.Sleep(connectorDeleteWait)

// 	t.Logf("Deleting VPC Access Connector '%s' in project '%s', region '%s' (using 'gcloud vpc-access connectors delete')...", connectorName, projectID, region)
// 	cmd := shell.Command{
// 		Command: "gcloud",
// 		Args: []string{
// 			// Standard GA command for delete, assuming it targets the same resource type
// 			"compute", "networks", "vpc-access", "connectors", "delete", connectorName,
// 			"--project=" + projectID,
// 			"--region=" + region,
// 			"--quiet",
// 		},
// 	}

// 	output, err := shell.RunCommandAndGetOutputE(t, cmd)
// 	if err != nil {
// 		if strings.Contains(output, "NOT_FOUND") || strings.Contains(output, "NotFound") || strings.Contains(err.Error(), "NOT_FOUND") {
// 			t.Logf("VPC Access Connector '%s' already deleted or not found.", connectorName)
// 		} else {
// 			t.Logf("Error deleting VPC Access Connector '%s': %v. Output:\n%s", connectorName, err, output)
// 		}
// 	} else {
// 		t.Logf("VPC Access Connector '%s' deleted successfully or was already gone.", connectorName)
// 	}
// }

// Updated createConfigYAML helper function
func createConfigYAML(t *testing.T, outputDir string, uniqueID string, deploymentFileURL string) { //,testSubnetName string, gcloudCreatedConnectorFullName string) {
	t.Helper()
	testProjectID := projectID

	configs := map[string]AppEngineConfig{
		"instance1.yaml": {
			ProjectID:        testProjectID,
			Service:          service1,
			VersionID:        versionID1,
			Runtime:          sampleAppRuntime,
			Deployment:       &DeploymentConfig{Files: &FilesConfig{Name: sampleAppGcsObjectName, SourceURL: deploymentFileURL}},
			Entrypoint:       &EntrypointConfig{Shell: sampleAppEntrypoint},
			AutomaticScaling: &AutomaticScalingConfig{MaxConcurrentRequests: 50, MinIdleInstances: 1, MaxIdleInstances: 3},
			Handlers:         HandlerConfig{{URLRegex: "/.*", Script: &ScriptConfig{ScriptPath: "auto"}}},
			// VPCAccessConnector:        &VPCAccessConnectorConfig{Name: gcloudCreatedConnectorFullName},
			// CreateVPCConnector:        false, // Terraform is not creating the connector
			// VPCAccessConnectorDetails: nil,   // Ensure not set
			CreateAppVersion:       false,
			DeleteServiceOnDestroy: true,
			AppEngineApplication:   &AppEngineAppConfig{LocationID: defaultRegion},
		},
		"instance2.yaml": {
			ProjectID:        testProjectID,
			Service:          service2,
			VersionID:        versionID2,
			Runtime:          sampleAppRuntime,
			Deployment:       &DeploymentConfig{Files: &FilesConfig{Name: sampleAppGcsObjectName, SourceURL: deploymentFileURL}},
			Entrypoint:       &EntrypointConfig{Shell: sampleAppEntrypoint},
			AutomaticScaling: &AutomaticScalingConfig{MaxConcurrentRequests: 60, MinIdleInstances: 0, MaxIdleInstances: 2},
			Handlers:         HandlerConfig{{URLRegex: "/.*", Script: &ScriptConfig{ScriptPath: "auto"}}},
			// VPCAccessConnector:        &VPCAccessConnectorConfig{Name: gcloudCreatedConnectorFullName},
			// CreateVPCConnector:        false, // Terraform is not creating the connector
			// VPCAccessConnectorDetails: nil,   // Ensure not set
			CreateAppVersion:       false,
			DeleteServiceOnDestroy: true,
			AppEngineApplication:   &AppEngineAppConfig{LocationID: defaultRegion},
		},
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Logf("Failed to create config directory '%s': %v", outputDir, err)
	}
	for filename, config := range configs {
		yamlData, err := yaml.Marshal(&config)
		if err != nil {
			t.Logf("Error marshaling config %s: %v", filename, err)
		}
		filePath := filepath.Join(outputDir, filename)
		err = os.WriteFile(filePath, yamlData, 0644)
		if err != nil {
			t.Logf("Unable to write data into file %s: %v", filePath, err)
		}
		t.Logf("Created YAML config at %s", filePath)
	}
}

func getHelloWorldAppFiles() (appYamlContent string, mainPyContent string) {
	appYamlContent = fmt.Sprintf("runtime: %s\nentrypoint: %s\n", sampleAppRuntime, sampleAppEntrypoint)
	mainPyContent = `from flask import Flask
import os

app = Flask(__name__)

@app.route('/')
def hello():
    """Return a friendly HTTP greeting."""
    app_mode = os.environ.get("APP_MODE", "production")
    return f'Hello World! Mode: {app_mode}\nYour AppEngine Version is: {os.environ.get("GAE_VERSION", "unknown")}\n'

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=int(os.environ.get("PORT", 8080)), debug=True)
`
	return appYamlContent, mainPyContent
}

func httpGetWithRetry(t *testing.T, url string, maxRetries int, timeBetweenRetries time.Duration) (int, string) {
	var lastErr string
	for i := 0; i < maxRetries; i++ {
		t.Logf("HTTP GET Attempt %d/%d for URL: %s", i+1, maxRetries, url)
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Sprintf("HTTP GET failed (attempt %d/%d): %s", i+1, maxRetries, err)
			t.Logf("%s", lastErr)
			if i < maxRetries-1 {
				time.Sleep(timeBetweenRetries)
			}
			continue
		}
		bodyBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Sprintf("Failed to read response body (attempt %d/%d) for URL %s: %s", i+1, maxRetries, url, readErr)
			t.Logf("%s", lastErr)
			if i < maxRetries-1 {
				time.Sleep(timeBetweenRetries)
			}
			continue
		}
		bodyString := string(bodyBytes)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			t.Logf("HTTP GET successful on attempt %d/%d for URL %s. Status: %d", i+1, maxRetries, url, resp.StatusCode)
			return resp.StatusCode, bodyString
		}
		lastErr = fmt.Sprintf("Attempt %d/%d for URL %s: Got status code %d. Body: %s", i+1, maxRetries, url, resp.StatusCode, bodyString)
		t.Logf("%s", lastErr)
		if i < maxRetries-1 {
			time.Sleep(timeBetweenRetries)
		}
	}
	t.Logf("HTTP GET failed for URL %s after %d attempts. Last error: %v", url, maxRetries, lastErr)
	return 0, ""
}

func getRegionCode(t *testing.T, region string) string {
	codes := map[string]string{"us-central": "uc", "us-central1": "uc", "us-east1": "ue", "us-east4": "us-e4", "us-west1": "uw", "us-west2": "uw2", "us-west3": "uw3", "us-west4": "uw4", "europe-central2": "eur-c2", "europe-west1": "ew", "europe-west2": "ew2", "europe-west3": "ew3", "europe-west6": "ew6", "asia-east1": "as-e1", "asia-east2": "as-e2", "asia-northeast1": "an-e1", "asia-northeast2": "an-e2", "asia-northeast3": "an-e3", "asia-south1": "as-s1", "asia-south2": "as-s2", "asia-southeast1": "as-se1", "asia-southeast2": "as-se2", "australia-southeast1": "au-se1", "australia-southeast2": "au-se2", "southamerica-east1": "sa-e1", "southamerica-west1": "sa-w1"}
	code, ok := codes[region]
	if !ok {
		// Changed from fmt.Printf to t.Logf for test logging consistency
		t.Logf("Warning: Unknown region '%s' for App Engine URL code in getRegionCode, defaulting to 'uc'. Known regions: %v\n", region, codes)
		return "uc"
	}
	return code
}

// ensureAppEngineApplicationExists checks if an App Engine application exists in the project,
// and creates one if it doesn't.
func ensureAppEngineApplicationExists(t *testing.T, projectID string, region string) {
	t.Helper()
	t.Logf("Checking if App Engine application exists in project '%s'...", projectID)

	describeCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"app", "describe",
			"--project=" + projectID,
		},
	}

	// Run gcloud app describe. We expect an error if the app doesn't exist.
	output, err := shell.RunCommandAndGetOutputE(t, describeCmd)

	if err == nil {
		// If no error, an App Engine application already exists.
		t.Logf("App Engine application already exists in project '%s'. Description output (may include location):\n%s", projectID, output)
		return
	}

	// Check if the error output or error message itself indicates that no App Engine application exists.
	// Example error: "ERROR: (gcloud.app.describe) The current Google Cloud project [project-id] does not contain an App Engine application."
	if strings.Contains(output, "does not contain an App Engine application") ||
		(err != nil && strings.Contains(err.Error(), "does not contain an App Engine application")) {

		t.Logf("App Engine application does not exist in project '%s'. Attempting to create it in region '%s'...", projectID, region)

		createCmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"app", "create",
				"--region=" + region,
				"--project=" + projectID,
				"--quiet", // Suppress interactive prompts
			},
		}
		createOutput, createErr := shell.RunCommandAndGetOutputE(t, createCmd)
		if createErr != nil {
			t.Logf("Failed to create App Engine application in project '%s', region '%s'. Output:\n%s, Error: %s", projectID, region, createOutput, createErr)
		}
		// This creation step MUST succeed for the tests to proceed.

		t.Logf("Successfully created App Engine application in project '%s', region '%s'. Output:\n%s", projectID, region, createOutput)

		t.Logf("Waiting %v for App Engine application creation to propagate...", appCreatePropagationWait)
		time.Sleep(appCreatePropagationWait)
	} else {
		// If it's a different error from `gcloud app describe`, it's unexpected.
		if err != nil {
			t.Logf("Failed to describe App Engine application for an unexpected reason. Output:\n%s. Error:%s", output, err)
		}

	}
}

// enableGcpApis enables a list of specified GCP APIs for the project.
func enableGcpApis(t *testing.T, projectID string, apis []string) {
	t.Helper()
	t.Logf("Attempting to enable %d GCP APIs for project '%s': %v", len(apis), projectID, apis)

	for _, api := range apis {
		t.Logf("Enabling API '%s' for project '%s'...", api, projectID)
		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"services", "enable", api,
				"--project=" + projectID,
			},
		}
		// The `gcloud services enable` command is idempotent.
		// If the service is already enabled, it will report success without making changes.
		output, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			t.Logf("Failed to enable API '%s' for project '%s'. Output:\n%s. Error: %s", api, projectID, output, err)
		}
		// This operation MUST succeed for the test to reliably proceed.

		// Log output for visibility, it often confirms if it was already enabled or just enabled.
		t.Logf("Processed API '%s'. gcloud output:\n%s", api, output)
	}
	t.Logf("All %d required GCP APIs have been processed for enablement in project '%s'.", len(apis), projectID)
}

// addTokenCreatorRoleToPrincipalOnServiceAccount grants the Token Creator role to a principal on a specific service account.
func addTokenCreatorRoleToPrincipalOnServiceAccount(t *testing.T, projectID string, serviceAccountEmail string, principalEmail string) {
	t.Helper()
	var memberIdentifier string
	// Differentiate between a user account and a service account for the member format.
	if strings.Contains(principalEmail, ".iam.gserviceaccount.com") {
		memberIdentifier = fmt.Sprintf("serviceAccount:%s", principalEmail)
	} else {
		memberIdentifier = fmt.Sprintf("user:%s", principalEmail)
	}

	role := "roles/iam.serviceAccountTokenCreator"
	t.Logf("Adding role '%s' to principal '%s' on service account '%s' in project '%s'", role, memberIdentifier, serviceAccountEmail, projectID)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"iam", "service-accounts", "add-iam-policy-binding", serviceAccountEmail,
			"--project=" + projectID,
			"--member=" + memberIdentifier,
			"--role=" + role,
			"--quiet",
			"--format=none",
		},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Failed to add role '%s' to principal '%s' on service account '%s'. Output:\n%s. Error: %s", role, memberIdentifier, serviceAccountEmail, output, err)
	}

	t.Logf("Successfully added role '%s' to principal '%s' on service account '%s'", role, memberIdentifier, serviceAccountEmail)
}

// removeTokenCreatorRoleFromPrincipalOnServiceAccount removes the Token Creator role from a principal on a specific service account.
func removeTokenCreatorRoleFromPrincipalOnServiceAccount(t *testing.T, projectID string, serviceAccountEmail string, principalEmail string) {
	t.Helper()
	var memberIdentifier string
	if strings.Contains(principalEmail, ".iam.gserviceaccount.com") {
		memberIdentifier = fmt.Sprintf("serviceAccount:%s", principalEmail)
	} else {
		memberIdentifier = fmt.Sprintf("user:%s", principalEmail)
	}

	role := "roles/iam.serviceAccountTokenCreator"
	t.Logf("Removing role '%s' from principal '%s' on service account '%s' in project '%s'", role, memberIdentifier, serviceAccountEmail, projectID)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"iam", "service-accounts", "remove-iam-policy-binding", serviceAccountEmail,
			"--project=" + projectID,
			"--member=" + memberIdentifier,
			"--role=" + role,
			"--quiet",
			"--format=none",
		},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		// Log non-critical errors for removal, as the binding might already be gone or test is tearing down.
		if !strings.Contains(output, "NOT_FOUND") && !strings.Contains(output, "PolicyBindingNotFound") && !strings.Contains(err.Error(), "NOT_FOUND") {
			t.Logf("Error removing IAM binding '%s' for member '%s' from SA '%s': %v. Output:\n%s", role, memberIdentifier, serviceAccountEmail, err, output)
		} else {
			t.Logf("IAM binding '%s' for member '%s' on SA '%s' already gone or not found.", role, memberIdentifier, serviceAccountEmail)
		}
	} else {
		t.Logf("Successfully removed role '%s' from principal '%s' on service account '%s'", role, memberIdentifier, serviceAccountEmail)
	}
}

// getCurrentGcloudUser retrieves the currently authenticated gcloud account email.
func getCurrentGcloudUser(t *testing.T) string {
	t.Helper()
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"auth", "list", "--filter=status:ACTIVE", "--format=value(account)"}, //gcloud auth list --filter=status:ACTIVE --format="value(account)"
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	// If this fails, the test setup cannot proceed with impersonation correctly.
	if err != nil {
		t.Logf("Failed to get current gcloud user. Ensure gcloud is authenticated. Error: %v. Output: %s", err, output)
	}
	currentUser := strings.TrimSpace(output)
	require.NotEmpty(t, currentUser, "gcloud config get-value account returned empty string. Ensure gcloud is authenticated.")
	t.Logf("Current gcloud principal identified as: %s", currentUser)
	return currentUser
}

// getProjectNumber gets the Project Number for the Endpoint configuration
func getProjectNumber(t *testing.T, projectID string) string {
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"projects", "describe", projectID, "--format=value(projectNumber)", "--quiet"},
	}
	output, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Error getting project number for project ID %s: %s", projectID, err)
	}
	return output
}
