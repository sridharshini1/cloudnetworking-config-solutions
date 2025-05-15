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
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

type AcceleratorConfig struct {
	Type      string `yaml:"type"`
	CoreCount int    `yaml:"core_count"`
}

type DataDisk struct {
	DiskSizeGB int    `yaml:"disk_size_gb"`
	DiskType   string `yaml:"disk_type"`
}

type VMImage struct {
	Project string `yaml:"project"`
	Family  string `yaml:"family"`
}

type NetworkInterface struct {
	Network        string `yaml:"network"`
	Subnet         string `yaml:"subnet"`
	NICType        string `yaml:"nic_type"`
	InternalIPOnly bool   `yaml:"internal_ip_only"`
}

type GCESetup struct {
	MachineType        string              `yaml:"machine_type"`
	DisablePublicIP    bool                `yaml:"disable_public_ip"`
	AcceleratorConfigs []AcceleratorConfig `yaml:"accelerator_configs"`
	Tags               []string            `yaml:"tags"`
	Labels             map[string]string   `yaml:"labels"`
	DataDisks          []DataDisk          `yaml:"data_disks"`
	VMImage            VMImage             `yaml:"vm_image"`
	NetworkInterfaces  []NetworkInterface  `yaml:"network_interfaces"`
	Metadata           map[string]string   `yaml:"metadata"`
	DisableProxyAccess bool                `yaml:"disable_proxy_access"`
}

type WorkbenchConfig struct {
	Name      string   `yaml:"name"`
	ProjectID string   `yaml:"project_id"`
	Location  string   `yaml:"location"`
	Region    string   `yaml:"region"`
	GCESetup  GCESetup `yaml:"gce_setup"`
}

var (
	projectRoot, _         = filepath.Abs("../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Workbench")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Workbench/config")

	projectID             = os.Getenv("TF_VAR_project_id")
	workbenchInstanceName = fmt.Sprintf("wb-%d", rand.Int())
	region                = "us-central1"
	zone                  = "us-central1-a"
	vpcName               = fmt.Sprintf("testing-net-wb-%d", rand.Int())
	subnetName            = fmt.Sprintf("testing-subnet-wb-%d", rand.Int())
	firewallRuleName      = fmt.Sprintf("testing-fw-wb-%d", rand.Int()) // Added firewall rule name
	yamlFileName          = "instance.yaml"
	machineType           = "e2-standard-4"
	acceleratorType       = "NVIDIA_TESLA_T4"
	acceleratorCoreCount  = 1
	tags                  = []string{"deeplearning-vm", "notebook-instance"}
	labels                = map[string]string{"purpose": "workbench-demo-1"}
	dataDiskSizeGB        = 200
	dataDiskType          = "PD_SSD"
	vmImageProject        = "cloud-notebooks-managed"
	vmImageFamily         = "workbench-instances"
	nicType               = "GVNIC"
	disablePublicIP       = true
	internalIPOnly        = true
	disableProxyAccess    = true

	// Metadata map declared as a variable
	metadata = map[string]string{
		"framework":       "TensorFlow:2.17",
		"notebooks-api":   "PROD",
		"shutdown-script": "/opt/deeplearning/bin/shutdown_script.sh",
	}
)

// TestWorkbenchInstances is an integration test that validates the creation, configuration and BigQuery integration.
func TestWorkbenchInstances(t *testing.T) {
	createConfigYAML(t)

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

	createVPC(t, projectID, vpcName)
	time.Sleep(30 * time.Second)
	// Create firewall rule to allow SSH
	createFirewallRule(t, projectID, firewallRuleName, vpcName)
	time.Sleep(30 * time.Second)

	defer deleteVPC(t, projectID, vpcName)
	defer deleteFirewallRule(t, projectID, firewallRuleName) // Added defer to delete firewall rule
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	instanceIDsOutput := terraform.OutputJson(t, terraformOptions, "workbench_instance_ids")
	proxyURIsOutput := terraform.OutputJson(t, terraformOptions, "workbench_instance_proxy_uris")

	t.Logf("Instance IDs: %s", instanceIDsOutput)
	t.Logf("Proxy URIs: %s", proxyURIsOutput)

	instanceIDs := gjson.Parse(instanceIDsOutput).Map()
	proxyURIs := gjson.Parse(proxyURIsOutput).Map()

	for instanceName := range instanceIDs {
		instanceID := instanceIDs[instanceName].String()
		proxyURI := proxyURIs[instanceName].String()

		t.Logf("Instance Name: %s, ID: %s, Proxy URI: %s", instanceName, instanceID, proxyURI)

		expectedProxyURI := ""
		if !disableProxyAccess {
			expectedProxyURI = proxyURI
		}
		assert.Equal(t, expectedProxyURI, proxyURI, fmt.Sprintf("Proxy URI mismatch for instance %s. Expected: %s, Got: %s", instanceName, expectedProxyURI, proxyURI))

		stdout, stderr, err := shell.RunCommandAndGetStdOutErrE(t, shell.Command{
			Command: "gcloud",
			Args:    []string{"compute", "instances", "describe", instanceName, "--project", projectID, "--zone", zone, "--format=json"},
		})
		if err != nil {
			t.Fatalf("Failed to describe instance %s: %v\nStderr: %s", instanceName, err, stderr)
		}
		instanceDetailsOutput := stdout

		var actualInstance map[string]interface{}
		err = json.Unmarshal([]byte(instanceDetailsOutput), &actualInstance)
		if err != nil {
			t.Fatalf("Failed to unmarshal instance details: %v", err)
		}

		serviceAccountEmail, err := getServiceAccountEmail(actualInstance, t)
		if err != nil {
			t.Fatalf("Failed to get service account email: %v", err)
		}
		// Validate and assign roles, including the Service Account User role to itself.
		err = validateAndAssignRoles(t, serviceAccountEmail, projectID)
		if err != nil {
			t.Fatalf("Failed to validate and assign roles: %v", err)
		}
		// Create BigQuery resources and populate them
		bigqueryDataset, bigqueryTable, err := createAndPopulateBigQuery(t, projectID, serviceAccountEmail)
		if err != nil {
			t.Fatalf("Failed to create and populate BigQuery resources: %v", err)
		}
		defer deleteBigQueryResources(t, projectID, bigqueryDataset)

		// Validate BigQuery connectivity from the Workbench instance
		err = validateBigQueryFromWorkbenchInstance(t, instanceName, projectID, zone, bigqueryDataset, bigqueryTable)
		if err != nil {
			t.Fatalf("Failed to validate BigQuery connectivity from Workbench instance: %v", err)
		}

		// Execute BigQuery query directly and validate
		err = validateBigQueryDirectly(t, projectID, bigqueryDataset, bigqueryTable)
		if err != nil {
			t.Fatalf("Failed to validate BigQuery query directly: %v", err)
		}

		yamlFilePath := filepath.Join(configFolderPath, yamlFileName)
		yamlFile, err := os.ReadFile(yamlFilePath)
		if err != nil {
			t.Fatalf("Error reading YAML file at %s: %s", yamlFilePath, err)
		}
		var expectedInstance WorkbenchConfig
		err = yaml.Unmarshal(yamlFile, &expectedInstance)
		if err != nil {
			t.Fatalf("Error unmarshaling YAML: %v", err)
		}
		instanceZone := extractZoneFromInstanceDetails(actualInstance, t)
		assert.Equal(t, zone, instanceZone, fmt.Sprintf("Zone mismatch. Expected: %s, Actual: %s", zone, instanceZone))

		expectedMachineType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/%s", projectID, zone, expectedInstance.GCESetup.MachineType)
		if expectedMachineType != actualInstance["machineType"] {
			t.Logf("Machine type mismatch. Expected: %s, Actual: %s", expectedMachineType, actualInstance["machineType"])
		}
		assert.Equal(t, expectedMachineType, actualInstance["machineType"], "Machine type mismatch")

		tagsInterface, ok := actualInstance["tags"].(map[string]interface{})["items"].([]interface{})
		if !ok {
			t.Fatalf("Tags are not of type []interface{}")
		}
		var tags []string
		for _, tag := range tagsInterface {
			tags = append(tags, tag.(string))
		}
		expectedTags := []string{"deeplearning-vm", "notebook-instance"}
		if !assert.ObjectsAreEqual(expectedTags, tags) {
			t.Logf("Tags mismatch. Expected: %v, Actual: %v", expectedTags, tags)
		}

		if projectID != actualInstance["labels"].(map[string]interface{})["consumer-project-id"] {
			t.Logf("Project ID mismatch. Expected: %s, Actual: %s", projectID, actualInstance["labels"].(map[string]interface{})["consumer-project-id"])
		}
		assert.Equal(t, projectID, actualInstance["labels"].(map[string]interface{})["consumer-project-id"], "Project ID mismatch")

		expectedZoneURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s", projectID, zone)
		if expectedZoneURL != actualInstance["zone"] {
			t.Logf("Location mismatch. Expected: %s, Actual: %s", expectedZoneURL, actualInstance["zone"])
		}
		assert.Equal(t, expectedZoneURL, actualInstance["zone"], "Location mismatch")
	}
}

// extractZoneFromInstanceDetails extracts the zone from the instance details.
// It checks if the zone information is present and returns it.
func extractZoneFromInstanceDetails(instanceDetails map[string]interface{}, t *testing.T) string {
	zoneURL, ok := instanceDetails["zone"].(string)
	if !ok {
		t.Fatalf("Zone information not found in instance details")
		return ""
	}
	parts := strings.Split(zoneURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// getServiceAccountEmail extracts the service account email from the instance details.
func getServiceAccountEmail(instanceDetails map[string]interface{}, t *testing.T) (string, error) {
	serviceAccounts, ok := instanceDetails["serviceAccounts"].([]interface{})
	if !ok || len(serviceAccounts) == 0 {
		return "", fmt.Errorf("Service ccounts not found in instance details")
	}
	serviceAccount, ok := serviceAccounts[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid service account format")
	}
	email, ok := serviceAccount["email"].(string)
	if !ok {
		return "", fmt.Errorf("Service account email not found")
	}
	return email, nil
}

// createConfigYAML creates a YAML configuration file for the Workbench instance.
func createConfigYAML(t *testing.T) {
	t.Log("========= YAML File =========")

	dynamicNetwork := fmt.Sprintf("projects/%s/global/networks/%s", projectID, vpcName)
	dynamicSubnet := fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectID, region, subnetName)

	workbenchInstance := WorkbenchConfig{
		Name:      workbenchInstanceName,
		ProjectID: projectID,
		Location:  zone,
		Region:    region,
		GCESetup: GCESetup{
			MachineType: machineType,
			AcceleratorConfigs: []AcceleratorConfig{
				{
					Type:      acceleratorType,
					CoreCount: acceleratorCoreCount,
				},
			},
			Tags:   tags,
			Labels: labels,
			DataDisks: []DataDisk{
				{
					DiskSizeGB: dataDiskSizeGB,
					DiskType:   dataDiskType,
				},
			},
			VMImage: VMImage{
				Project: vmImageProject,
				Family:  vmImageFamily,
			},
			NetworkInterfaces: []NetworkInterface{
				{
					Network:        dynamicNetwork,
					Subnet:         dynamicSubnet,
					NICType:        nicType,
					InternalIPOnly: internalIPOnly,
				},
			},
			Metadata:           metadata,
			DisablePublicIP:    disablePublicIP,
			DisableProxyAccess: disableProxyAccess,
		},
	}

	yamlData, err := yaml.Marshal(&workbenchInstance)
	if err != nil {
		t.Fatalf("Error while marshaling: %v", err)
	}

	configDir := "config"
	filePath := filepath.Join(configDir, yamlFileName)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	t.Logf("Created YAML config at %s with content:\n%s", filePath, string(yamlData))

	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		t.Fatalf("Unable to write data into the file: %v", err)
	}
}

// createVPC creates a VPC network and a subnet in the specified region.
func createVPC(t *testing.T, projectID string, vpcName string) {
	text := "compute"

	// Create the VPC network
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", vpcName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating VPC network %s: %v", vpcName, err)
		return
	}

	// Create the subnet with Private Google Access enabled
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--region=" + region,
			"--range=10.0.0.0/24",
			"--enable-private-ip-google-access",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating subnet %s: %v", subnetName, err)
		return
	}

}

// deleteVPC deletes the VPC and subnet configuration after the test.
func deleteVPC(t *testing.T, projectID string, vpcName string) {
	text := "compute"
	time.Sleep(60 * time.Second) // Wait for resources to be in a deletable state

	// Delete Subnet
	cmd := shell.Command{
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

	time.Sleep(150 * time.Second)

	// Delete VPC
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", vpcName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("===Error %s Encountered while deleting VPC: %s", err, vpcName)
	}
}

// createFirewallRule creates a firewall rule to allow SSH access (port 22)
func createFirewallRule(t *testing.T, projectID, firewallRuleName, vpcName string) {
	t.Logf("Creating firewall rule: %s", firewallRuleName)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "create", firewallRuleName,
			"--project=" + projectID,
			"--network=" + vpcName,
			"--allow=tcp:22",
			"--source-ranges=0.0.0.0/0",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating firewall rule %s: %v", firewallRuleName, err)
	}
	t.Logf("Firewall rule %s created successfully", firewallRuleName)
}

// deleteFirewallRule deletes the specified firewall rule
func deleteFirewallRule(t *testing.T, projectID, firewallRuleName string) {
	t.Logf("Deleting firewall rule: %s", firewallRuleName)
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "firewall-rules", "delete", firewallRuleName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Error deleting firewall rule %s: %v", firewallRuleName, err)
	}
	t.Logf("Firewall rule %s deleted successfully", firewallRuleName)
}

// validateAndAssignRoles validates and assigns necessary roles to the service account.
func validateAndAssignRoles(t *testing.T, serviceAccountEmail string, projectID string) error {
	t.Helper()
	requiredRoles := []string{
		"roles/bigquery.jobUser",
		"roles/bigquery.dataViewer",
		"roles/serviceusage.serviceUsageConsumer",
		"roles/notebooks.admin",        // Or a more specific role if needed
		"roles/iam.serviceAccountUser", // To allow the SA to act as itself
	}

	member := fmt.Sprintf("serviceAccount:%s", serviceAccountEmail)

	for _, role := range requiredRoles {
		t.Logf("Adding role %s to %s for project %s", role, member, projectID)
		cmd := shell.Command{
			Command: "gcloud",
			Args:    []string{"projects", "add-iam-policy-binding", projectID, "--member=" + member, "--role=" + role, "--condition=None"},
		}
		output, err := shell.RunCommandAndGetOutputE(t, cmd)
		t.Logf("Output of IAM binding command for %s: %s", role, output)

		if err != nil {
			t.Logf("Failed to add IAM binding '%s' for member '%s' to project '%s'. Output:\n%s, Error: %v", role, member, projectID, output, err)
			return fmt.Errorf("failed to add IAM binding '%s' for member '%s': %w", role, member, err)
		}
		t.Logf("Successfully added role %s to %s for project %s", role, member, projectID)
	}

	return nil
}

// validateBigQueryFromWorkbenchInstance validates BigQuery connectivity from the Workbench instance using the bq CLI.
func validateBigQueryFromWorkbenchInstance(t *testing.T, instanceName, projectID, zone, datasetID, tableID string) error {
	t.Logf("Validating BigQuery connectivity from Workbench instance: %s", instanceName)

	// Construct the BigQuery query command
	queryCommand := fmt.Sprintf("bq query --use_legacy_sql=false 'SELECT COUNT(*) FROM `%s.%s`'", datasetID, tableID)

	// Execute the query command on the Workbench instance via SSH
	execCmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "ssh", fmt.Sprintf("jupyter@%s", instanceName),
			"--project", projectID,
			"--zone", zone,
			"--command", queryCommand,
		},
	}

	output, err := shell.RunCommandAndGetOutputE(t, execCmd)
	if err != nil {
		return fmt.Errorf("failed to execute BigQuery query on Workbench instance: %w", err)
	}
	// Check if the output contains the expected result
	expectedOutput := "3"
	if !strings.Contains(output, expectedOutput) {
		return fmt.Errorf("expected output '%s' not found in the output: %s", expectedOutput, output)
	}

	t.Logf("BigQuery connectivity validated from Workbench instance. Output:\n%s", output)
	return nil
}

// createAndPopulateBigQuery creates a BigQuery dataset and table, populates the table with sample data,
// and returns the dataset and table IDs along with any error encountered.
func createAndPopulateBigQuery(t *testing.T, projectID string, serviceAccountEmail string) (string, string, error) {
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get default credentials: %w", err)
	}
	t.Logf("Credentials obtained with email: %s, using project ID: %s for BigQuery", credentials.Email, projectID)

	bqClient, err := bigquery.NewClient(ctx, projectID, option.WithCredentials(credentials))
	if err != nil {
		return "", "", fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	defer bqClient.Close()

	datasetID := fmt.Sprintf("test_dataset_%d", rand.Int())
	tableID := fmt.Sprintf("test_table_%d", rand.Int())

	// Create dataset
	t.Logf("Creating BigQuery Dataset: %s", datasetID)
	dataset := bqClient.Dataset(datasetID)
	if err := dataset.Create(ctx, &bigquery.DatasetMetadata{
		Location: region,
	}); err != nil {
		t.Logf("Error creating dataset %s: %+v", datasetID, err)
		return "", "", fmt.Errorf("failed to create dataset: %w", err)
	}
	t.Logf("Created BigQuery Dataset: %s", datasetID)

	// Create table
	t.Logf("Creating BigQuery Table: %s", tableID)
	schema := bigquery.Schema{
		{Name: "name", Type: bigquery.StringFieldType},
		{Name: "value", Type: bigquery.IntegerFieldType},
	}
	table := dataset.Table(tableID)
	if err := table.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
		return "", "", fmt.Errorf("failed to create table: %w", err)
	}
	t.Logf("Created BigQuery Table: %s", tableID)

	// Populate table with data
	t.Log("Populating BigQuery Table with data")
	rows := []struct {
		Name  string `bigquery:"name"`
		Value int    `bigquery:"value"`
	}{
		{Name: "Item A", Value: 10},
		{Name: "Item B", Value: 20},
		{Name: "Item C", Value: 30},
	}

	inserter := table.Inserter()
	t.Logf("Inserter created: %v", inserter)
	if err := inserter.Put(ctx, rows); err != nil {
		return "", "", fmt.Errorf("failed to insert data: %w", err)
	}
	t.Logf("Data inserted: %v", rows)
	t.Logf("Populated BigQuery Table with data")

	return datasetID, tableID, nil
}

// deleteBigQueryResources deletes a BigQuery dataset and its contents for the given project and dataset IDs.
// It ensures proper cleanup of resources during integration tests.
func deleteBigQueryResources(t *testing.T, projectID string, datasetID string) {
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		t.Fatalf("failed to get default credentials: %v", err)
		return
	}
	t.Logf("Credentials obtained with email: %s, using project ID: %s for BigQuery deletion", credentials.Email, projectID)

	bqClient, err := bigquery.NewClient(ctx, projectID, option.WithCredentials(credentials))
	if err != nil {
		t.Fatalf("failed to create BigQuery client: %v", err)
		return
	}
	defer bqClient.Close()

	dataset := bqClient.Dataset(datasetID)
	if err := dataset.DeleteWithContents(ctx); err != nil {
		t.Fatalf("failed to delete dataset: %v", err)
		return
	}
	t.Logf("Deleted BigQuery Dataset: %s", datasetID)
}

// validateBigQueryDirectly executes a query directly against BigQuery and validates the result.
func validateBigQueryDirectly(t *testing.T, projectID string, datasetID string, tableID string) error {
	t.Logf("Executing BigQuery query directly... using project ID: %s", projectID)

	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get default credentials: %w", err)
	}
	t.Logf("Credentials obtained with email: %s", credentials.Email)

	bqClient, err := bigquery.NewClient(ctx, projectID, option.WithCredentials(credentials))
	if err != nil {
		return fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	defer bqClient.Close()
	t.Logf("BigQuery client created successfully")

	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s.%s`", datasetID, tableID)
	t.Logf("Query: %s", query)
	q := bqClient.Query(query)
	job, err := q.Run(ctx)
	if err != nil {
		t.Logf("Error running query: %+v", err)
		return fmt.Errorf("failed to run query: %w", err)
	}
	t.Logf("Query job started successfully")

	status, err := job.Wait(ctx)
	if err != nil {
		t.Logf("Error waiting for query job: %+v", err)
		return fmt.Errorf("failed to wait for query job: %w", err)
	}
	if err := status.Err(); err != nil {
		t.Logf("Query job failed with error: %+v", status.Err())
		return fmt.Errorf("query failed with error: %w", err)
	}
	t.Logf("Query job finished successfully")

	it, err := job.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read query results: %w", err)
	}
	t.Logf("Query results read successfully")

	var row []bigquery.Value
	err = it.Next(&row)
	if err != nil {
		return fmt.Errorf("failed to get next row: %w", err)
	}
	if len(row) == 0 {
		return fmt.Errorf("no results returned from the query")
	}
	count := row[0].(int64)
	t.Logf("BigQuery query successful. Count: %v", count)
	assert.Equal(t, int64(3), int64(count), "The count of the table should be 3")
	return nil
}
