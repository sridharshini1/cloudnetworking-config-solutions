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
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
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

var (
	projectRoot, _         = filepath.Abs("../../../../../../")
	terraformDirectoryPath = filepath.Join(projectRoot, "06-consumer/Serverless/AppEngine/Flexible")
	configFolderPath       = filepath.Join(projectRoot, "test/integration/consumer/Serverless/AppEngine/Flexible/config")
)

var (
	projectID           = os.Getenv("TF_VAR_project_id")
	instanceName        string
	region              = "us-central1"
	networkName         string
	serviceAccountName  string
	serviceAccountEmail string
	gcsBucketName       string
	gcsSourceURL        string
)

type AppEngineConfig struct {
	Project                 string `yaml:"project_id"`
	Service                 string `yaml:"service"`
	Runtime                 string `yaml:"runtime"`
	FlexibleRuntimeSettings struct {
		OperatingSystem string `yaml:"operating_system"`
		RuntimeVersion  string `yaml:"runtime_version"`
	} `yaml:"flexible_runtime_settings"`
	InstanceClass string `yaml:"instance_class"`
	Network       struct {
		Name       string `yaml:"name"`
		Subnetwork string `yaml:"subnetwork"`
	} `yaml:"network"`
	VersionID string `yaml:"version_id"`

	AutomaticScaling struct {
		CoolDownPeriod        string                 `yaml:"cool_down_period"`
		MaxConcurrentRequests int                    `yaml:"max_concurrent_requests"`
		MaxTotalInstances     int                    `yaml:"max_total_instances"`
		MinTotalInstances     int                    `yaml:"min_total_instances"`
		CPUUtilization        map[string]interface{} `yaml:"cpu_utilization"`
	} `yaml:"automatic_scaling"`

	Entrypoint struct {
		Shell string `yaml:"shell"`
	} `yaml:"entrypoint"`

	Deployment *struct {
		Zip *struct {
			SourceURL string `yaml:"source_url"`
		} `yaml:"zip,omitempty"`
	} `yaml:"deployment,omitempty"`
	LivenessCheck          map[string]interface{} `yaml:"liveness_check,omitempty"`
	ReadinessCheck         map[string]interface{} `yaml:"readiness_check,omitempty"`
	DeleteServiceOnDestroy bool                   `yaml:"delete_service_on_destroy,omitempty"`
	EnvVariables           map[string]string      `yaml:"env_variables,omitempty"`
	ServiceAccount         string                 `yaml:"service_account,omitempty"`
	Labels                 map[string]string      `yaml:"labels,omitempty"`
}

func createServiceAccount(t *testing.T, projectID, saName, displayName string) (string, error) {
	t.Logf("Attempting to create or verify service account: %s in project %s", saName, projectID)
	expectedSaEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, projectID)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"iam", "service-accounts", "create", saName,
			"--project=" + projectID,
			"--display-name=" + displayName,
		},
	}
	rawOutput, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		if strings.Contains(rawOutput, "already exists") || strings.Contains(err.Error(), "already exists") || strings.Contains(rawOutput, "Service account already exists") {
			t.Logf("Service account %s (email: %s) already exists. Assuming it's usable.", saName, expectedSaEmail)
			return expectedSaEmail, nil
		}
		return "", fmt.Errorf("failed to create service account %s: %w. Raw output: %s", saName, err, rawOutput)
	}
	t.Logf("Service account %s created/verified. Email: %s. Full gcloud output: %s", saName, expectedSaEmail, rawOutput)
	return expectedSaEmail, nil
}

func addIAMPolicyBinding(t *testing.T, projectID, saEmail, role string) error {
	t.Logf("Adding role %s to service account %s for project %s", role, saEmail, projectID)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"projects", "add-iam-policy-binding", projectID, "--member=serviceAccount:" + saEmail, "--role=" + role, "--format=json"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		return fmt.Errorf("failed to add role %s to SA %s: %w", role, saEmail, err)
	}
	t.Logf("Role %s added to %s successfully.", role, saEmail)
	return nil
}

func deleteServiceAccount(t *testing.T, projectID, saEmail string) {
	if saEmail == "" {
		return
	}
	t.Logf("--- Deleting service account: %s in project %s ---", saEmail, projectID)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{"iam", "service-accounts", "delete", saEmail, "--project=" + projectID, "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Error deleting service account %s: %v. Manual cleanup might be required.", saEmail, err)
	} else {
		t.Logf("Service account %s deleted successfully or did not exist.", saEmail)
	}
}

func createGCSBucket(t *testing.T, projectID, bucketName string) error {
	t.Logf("Creating GCS bucket: gs://%s in project %s", bucketName, projectID)
	cmd := shell.Command{
		Command: "gsutil",
		Args:    []string{"mb", "-p", projectID, "-l", region, fmt.Sprintf("gs://%s", bucketName)},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		if strings.Contains(err.Error(), "SERVICE_ACCOUNT_NOT_SET_UP") || strings.Contains(err.Error(), "Service account has not been granted Legacy Bucket Writer role") {
			t.Logf("Storage service account might not be set up yet for project %s. This can happen on new projects. Trying one more time after a delay or check project setup.", projectID)
			time.Sleep(30 * time.Second)
			_, err = shell.RunCommandAndGetOutputE(t, cmd)
		}
		if err != nil && !strings.Contains(err.Error(), "you already own it") {
			return fmt.Errorf("failed to create GCS bucket gs://%s: %w", bucketName, err)
		}
	}
	t.Logf("GCS bucket gs://%s created/verified successfully.", bucketName)
	return nil
}

func downloadFile(t *testing.T, url string, destDir string, fileName string) (string, error) {
	filePath := filepath.Join(destDir, fileName)
	t.Logf("Downloading %s to %s", url, filePath)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: status %s", url, resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func createZipArchive(t *testing.T, sourceDir string, targetZipPath string, filesToZip map[string]string) error {
	t.Logf("Creating zip archive %s from contents of %s", targetZipPath, sourceDir)
	zipFile, err := os.Create(targetZipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file %s: %w", targetZipPath, err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for sourcePath, pathInZip := range filesToZip {
		fullSourcePath := filepath.Join(sourceDir, sourcePath)
		t.Logf("Adding to zip: %s as %s", fullSourcePath, pathInZip)
		fileToZip, err := os.Open(fullSourcePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s for zipping: %w", fullSourcePath, err)
		}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", fullSourcePath, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", fullSourcePath, err)
		}
		header.Name = pathInZip
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip writer for %s: %w", pathInZip, err)
		}
		if _, err = io.Copy(writer, fileToZip); err != nil {
			return fmt.Errorf("failed to write file %s to zip: %w", pathInZip, err)
		}
	}
	t.Logf("Zip archive %s created successfully.", targetZipPath)
	return nil
}

func prepareAppSourceZip(t *testing.T) (zipFilePath string, tempDirPath string, err error) {
	tempDir, err := os.MkdirTemp("", "appengine-source-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	t.Logf("Created temporary directory for app source: %s", tempDir)

	baseURL := "https://raw.githubusercontent.com/GoogleCloudPlatform/python-docs-samples/main/appengine/flexible/hello_world/"
	filesToIncludeInZip := map[string]string{
		"main.py":          "main.py",
		"requirements.txt": "requirements.txt",
		"app.yaml":         "app.yaml",
	}

	for localFileName, _ := range filesToIncludeInZip {
		_, err := downloadFile(t, baseURL+localFileName, tempDir, localFileName)
		if err != nil {
			_ = os.RemoveAll(tempDir)
			return "", tempDir, fmt.Errorf("failed to download %s: %w", localFileName, err)
		}
	}
	zipOutputFilePath := filepath.Join(tempDir, "app_source.zip")
	err = createZipArchive(t, tempDir, zipOutputFilePath, filesToIncludeInZip)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", tempDir, fmt.Errorf("failed to create zip archive: %w", err)
	}
	return zipOutputFilePath, tempDir, nil
}

func uploadToGCS(t *testing.T, localFilePath, bucketName, objectName string) (string, error) {
	t.Logf("Uploading %s to gs://%s/%s", localFilePath, bucketName, objectName)
	gsutilUploadPath := fmt.Sprintf("gs://%s/%s", bucketName, objectName)

	cmd := shell.Command{
		Command: "gsutil",
		Args:    []string{"cp", localFilePath, gsutilUploadPath},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to upload %s to %s: %w", localFilePath, gsutilUploadPath, err)
	}
	httpsURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	t.Logf("File %s uploaded successfully. App Engine source URL: %s", filepath.Base(localFilePath), httpsURL)
	return httpsURL, nil
}

func deleteGCSBucket(t *testing.T, bucketName string) {
	if bucketName == "" {
		return
	}
	t.Logf("--- Deleting GCS bucket: gs://%s ---", bucketName)
	cmd := shell.Command{
		Command: "gsutil",
		Args:    []string{"rm", "-r", "-f", fmt.Sprintf("gs://%s", bucketName)},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Logf("Note: Error during gsutil rm for bucket gs://%s (may be benign if bucket was already empty/gone or due to permissions): %v", bucketName, err)
	} else {
		t.Logf("GCS bucket gs://%s and its contents deleted successfully or did not exist.", bucketName)
	}
}
func getBaseAppEngineConfig(t *testing.T) AppEngineConfig {
	return AppEngineConfig{
		Project: projectID,
		Runtime: "python",
		FlexibleRuntimeSettings: struct {
			OperatingSystem string `yaml:"operating_system"`
			RuntimeVersion  string `yaml:"runtime_version"`
		}{
			OperatingSystem: "ubuntu22",
			RuntimeVersion:  "3.12",
		},
		Network: struct {
			Name       string `yaml:"name"`
			Subnetwork string `yaml:"subnetwork"`
		}{
			Name:       networkName,
			Subnetwork: fmt.Sprintf("%s-subnet", networkName),
		},
		VersionID: "v1",
		AutomaticScaling: struct {
			CoolDownPeriod        string                 `yaml:"cool_down_period"`
			MaxConcurrentRequests int                    `yaml:"max_concurrent_requests"`
			MaxTotalInstances     int                    `yaml:"max_total_instances"`
			MinTotalInstances     int                    `yaml:"min_total_instances"`
			CPUUtilization        map[string]interface{} `yaml:"cpu_utilization"`
		}{
			CoolDownPeriod:        "120s",
			MaxConcurrentRequests: 50,
			MaxTotalInstances:     10,
			MinTotalInstances:     2,
			CPUUtilization: map[string]interface{}{
				"target_utilization":        0.6,
				"aggregation_window_length": "60s",
			},
		},
		Entrypoint: struct {
			Shell string `yaml:"shell"`
		}{
			Shell: "pip3 install gunicorn flask && gunicorn -b :8080 main:app"},
		LivenessCheck: map[string]interface{}{
			"path":          "/",
			"initial_delay": "300s",
		},
		ReadinessCheck: map[string]interface{}{
			"path":              "/",
			"app_start_timeout": "300s",
		},
		DeleteServiceOnDestroy: true,
	}
}
func createConfigYAML(t *testing.T, currentSaEmail string, currentGcsSourceURL string) []AppEngineConfig {
	t.Log("Generating YAML configuration files aligned with the minimal working example, with dynamic SA and GCS URL...")
	baseConfig := getBaseAppEngineConfig(t) // This now returns a minimal config
	service1Config := baseConfig
	service1Config.Service = "test-service1"
	service1Config.InstanceClass = "F4_1G"
	service1Config.Deployment = &struct {
		Zip *struct {
			SourceURL string `yaml:"source_url"`
		} `yaml:"zip,omitempty"`
	}{
		Zip: &struct {
			SourceURL string `yaml:"source_url"`
		}{
			SourceURL: currentGcsSourceURL,
		},
	}
	service1Config.ServiceAccount = currentSaEmail
	service1Config.EnvVariables = map[string]string{
		"SERVICE_ID": "1",
		"GCS_SOURCE": currentGcsSourceURL,
	}
	service1Config.Labels = map[string]string{
		"test-type":   "integration",
		"managed-by":  "terratest",
		"specific-to": "service1",
	}
	service2Config := baseConfig
	service2Config.Service = "test-service2"
	service2Config.AutomaticScaling.MinTotalInstances = 1
	service2Config.AutomaticScaling.MaxTotalInstances = 3

	service2Config.Deployment = &struct {
		Zip *struct {
			SourceURL string `yaml:"source_url"`
		} `yaml:"zip,omitempty"`
	}{
		Zip: &struct {
			SourceURL string `yaml:"source_url"`
		}{
			SourceURL: currentGcsSourceURL,
		},
	}
	service2Config.ServiceAccount = currentSaEmail
	service2Config.EnvVariables = map[string]string{
		"SERVICE_ID": "2",
		"GCS_SOURCE": currentGcsSourceURL,
	}
	service2Config.Labels = map[string]string{
		"test-type":   "integration",
		"managed-by":  "terratest",
		"specific-to": "service2",
		"feature-x":   "enabled", // Example of an additional label for service2
	}
	servicesToCreate := []AppEngineConfig{service1Config, service2Config}
	if err := os.RemoveAll(configFolderPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to clean config directory %s: %v", configFolderPath, err)
	}
	if err := os.MkdirAll(configFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory %s: %v", configFolderPath, err)
	}
	t.Logf("Using config folder for generated YAMLs: %s", configFolderPath)

	// Write each service configuration to its own YAML file
	for _, serviceCfg := range servicesToCreate {
		yamlData, err := yaml.Marshal(&serviceCfg)
		if err != nil {
			t.Fatalf("Error marshaling YAML for service %s: %v", serviceCfg.Service, err)
		}

		filePath := filepath.Join(configFolderPath, fmt.Sprintf("%s.yaml", serviceCfg.Service))
		err = os.WriteFile(filePath, yamlData, 0644)
		if err != nil {
			t.Fatalf("Error writing YAML file for service %s at %s: %v", serviceCfg.Service, filePath, err)
		}
		t.Logf("Created YAML config file: %s\n--- Content ---\n%s\n---------------", filePath, string(yamlData))
	}

	return servicesToCreate
}

func createVPC(t *testing.T, projectID string, networkName string) bool {
	text := "compute"
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	t.Logf("Creating VPC Network: %s in project %s", networkName, projectID)
	cmdVPC := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "create", networkName,
			"--project=" + projectID,
			"--subnet-mode=custom",
			"--mtu=1460",
			"--bgp-routing-mode=regional",
			"--format=json",
		},
	}
	_, errVPC := shell.RunCommandAndGetOutputE(t, cmdVPC)
	if errVPC != nil {
		if strings.Contains(errVPC.Error(), "already exists") {
			t.Logf("VPC Network %s already exists. Proceeding.", networkName)
		} else {
			t.Errorf("Error creating VPC Network %s: %v", networkName, errVPC)
			return false
		}
	} else {
		t.Logf("VPC Network %s created successfully.", networkName)
	}
	t.Logf("Creating Subnet: %s in network %s, region %s", subnetName, networkName, region)
	cmdSubnet := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "create", subnetName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--region=" + region,
			"--range=10.128.0.0/20",
			"--enable-private-ip-google-access",
			"--format=json",
		},
	}
	_, errSubnet := shell.RunCommandAndGetOutputE(t, cmdSubnet)
	if errSubnet != nil {
		if strings.Contains(errSubnet.Error(), "already exists") {
			t.Logf("Subnet %s in network %s already exists. Proceeding.", subnetName, networkName)
		} else {
			t.Errorf("Error creating Subnet %s in network %s: %v", subnetName, networkName, errSubnet)
			return false
		}
	} else {
		t.Logf("Subnet %s in network %s created successfully.", subnetName, networkName)
	}
	return true
}

func deleteVPC(t *testing.T, projectID string, networkName string) {
	text := "compute"
	subnetName := fmt.Sprintf("%s-subnet", networkName)
	t.Logf("--- Starting VPC Cleanup for network: %s ---", networkName)

	t.Logf("Attempting to delete Subnet: %s in region %s", subnetName, region)
	cmdSubnet := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "delete", subnetName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet", // Suppress confirmation prompts
		},
	}
	_, errSubnet := shell.RunCommandAndGetOutputE(t, cmdSubnet)
	if errSubnet != nil {
		t.Logf("Note: Error deleting Subnet %s (may be benign if already gone or due to dependencies): %v", subnetName, errSubnet)
	} else {
		t.Logf("Subnet %s deleted successfully or did not exist.", subnetName)
	}

	time.Sleep(15 * time.Second)
	t.Logf("Attempting to delete VPC Network: %s", networkName)
	cmdVPC := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "delete", networkName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	_, errVPC := shell.RunCommandAndGetOutputE(t, cmdVPC)
	if errVPC != nil {
		t.Logf("Note: Error deleting VPC Network %s (may be benign if already gone or due to dependencies like firewall rules not fully detached): %v", networkName, errVPC)
	} else {
		t.Logf("VPC Network %s deleted successfully or did not exist.", networkName)
	}
	t.Logf("--- VPC Cleanup for network %s finished ---", networkName)
}

func createFirewallRules(t *testing.T, projectID string, networkName string, instanceName string) bool {
	t.Logf("Creating firewall rules for network %s (instance suffix %s)...", networkName, instanceName)
	rulesToCreate := map[string]string{
		fmt.Sprintf("fw-allow-ssh-%s", instanceName):   "tcp:22",
		fmt.Sprintf("fw-allow-http-%s", instanceName):  "tcp:80",
		fmt.Sprintf("fw-allow-https-%s", instanceName): "tcp:443",
	}

	allSucceeded := true
	for ruleName, ruleProtoPort := range rulesToCreate {
		t.Logf("Creating firewall rule: %s for %s", ruleName, ruleProtoPort)
		cmd := shell.Command{
			Command: "gcloud",
			Args: []string{
				"compute", "firewall-rules", "create", ruleName,
				"--project=" + projectID,
				"--network=" + networkName,
				"--direction=INGRESS",
				"--priority=1000",
				"--action=ALLOW",
				"--rules=" + ruleProtoPort,
				"--source-ranges=0.0.0.0/0",
				"--format=json",
			},
		}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				t.Logf("Firewall rule %s already exists. Proceeding.", ruleName)
			} else {
				t.Errorf("Error creating firewall rule %s: %v", ruleName, err)
				allSucceeded = false
			}
		} else {
			t.Logf("Firewall rule %s created successfully.", ruleName)
		}
	}
	if !allSucceeded {
		t.Error("One or more firewall rules failed to create properly.")
	}
	return allSucceeded
}

func deleteFirewallRules(t *testing.T, projectID string, instanceName string) {
	t.Logf("--- Starting Firewall Rule Cleanup for instance suffix: %s ---", instanceName)
	rulesToDelete := []string{
		fmt.Sprintf("fw-allow-ssh-%s", instanceName),
		fmt.Sprintf("fw-allow-http-%s", instanceName),
		fmt.Sprintf("fw-allow-https-%s", instanceName),
	}

	for _, ruleName := range rulesToDelete {
		t.Logf("Attempting to delete firewall rule: %s", ruleName)
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
			t.Logf("Note: Error deleting firewall rule %s (may be benign if already gone): %v", ruleName, err)
		} else {
			t.Logf("Firewall rule %s deleted successfully or did not exist.", ruleName)
		}
	}
	t.Logf("--- Firewall Rule Cleanup for instance suffix %s finished ---", instanceName)
}

func TestCreateAppEngine(t *testing.T) {
	if projectID == "" {
		t.Fatal("TF_VAR_project_id environment variable must be set")
	}
	instanceName = fmt.Sprintf("appeng-flex-test-%d", rand.Intn(10000))
	networkName = fmt.Sprintf("vpc-%s", instanceName)
	serviceAccountName = fmt.Sprintf("sa-%s", instanceName)
	gcsBucketName = fmt.Sprintf("bkt-%s-%s", strings.ToLower(projectID), instanceName)
	gcsBucketName = strings.ReplaceAll(gcsBucketName, "_", "-")
	if len(gcsBucketName) > 63 {
		gcsBucketName = gcsBucketName[:63]
	}

	t.Logf("Test Run Config: ProjectID=%s, InstanceSuffix=%s, Network=%s, SA=%s, Bucket=%s",
		projectID, instanceName, networkName, serviceAccountName, gcsBucketName)
	var err error
	serviceAccountEmail, err = createServiceAccount(t, projectID, serviceAccountName, "AppEngine Test SA")
	if err != nil {
		t.Fatalf("Failed to create service account: %v", err)
	}
	defer deleteServiceAccount(t, projectID, serviceAccountEmail)
	rolesToGrant := []string{
		"roles/logging.logWriter",
		"roles/artifactregistry.admin",
		"roles/cloudbuild.builds.editor",
		"roles/cloudsql.client",
		"roles/storage.admin",
		"roles/monitoring.metricWriter",
	}
	for _, role := range rolesToGrant {
		err = addIAMPolicyBinding(t, projectID, serviceAccountEmail, role)
		if err != nil {
			t.Fatalf("Failed to add role %s to SA %s: %v", role, serviceAccountEmail, err)
		}
	}
	t.Logf("IAM roles granted to service account %s", serviceAccountEmail)

	err = createGCSBucket(t, projectID, gcsBucketName)
	if err != nil {
		t.Fatalf("Failed to create GCS bucket: %v", err)
	}
	defer deleteGCSBucket(t, gcsBucketName)

	localZipPath, sourceTempDir, err := prepareAppSourceZip(t)
	if err != nil {
		t.Fatalf("Failed to prepare app source zip: %v", err)
	}
	defer func() {
		t.Logf("Cleaning up temporary source directory: %s", sourceTempDir)
		if err := os.RemoveAll(sourceTempDir); err != nil {
			t.Logf("WARN: Failed to remove temp source directory %s: %v", sourceTempDir, err)
		}
	}()

	fileInfo, statErr := os.Stat(localZipPath)
	if statErr != nil {
		t.Fatalf("Zip file expected at %s was not found after prepareAppSourceZip: %v", localZipPath, statErr)
	}
	if fileInfo.Size() == 0 {
		t.Fatalf("Zip file at %s is empty after prepareAppSourceZip. Size: %d bytes", localZipPath, fileInfo.Size())
	}
	t.Logf("Local zip file %s prepared successfully, size: %d bytes. Proceeding to upload.", localZipPath, fileInfo.Size())

	gcsSourceURL, err = uploadToGCS(t, localZipPath, gcsBucketName, "app_source.zip")
	if err != nil {
		t.Fatalf("Failed to upload app source to GCS: %v", err)
	}
	t.Logf("App source uploaded to: %s", gcsSourceURL)

	generatedConfigs := createConfigYAML(t, serviceAccountEmail, gcsSourceURL)
	if len(generatedConfigs) == 0 {
		t.Fatal("No YAML configurations were generated.")
	}

	tfVars := map[string]interface{}{"config_folder_path": configFolderPath}
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars: tfVars, TerraformDir: terraformDirectoryPath, Reconfigure: true, Lock: true, NoColor: true, SetVarsAfterVarFiles: true,
	})
	if !createVPC(t, projectID, networkName) {
		t.Fatal("VPC/Subnet creation failed.")
	}
	defer deleteVPC(t, projectID, networkName)
	t.Log("VPC/Subnet created. Waiting for propagation...")
	time.Sleep(60 * time.Second)

	if !createFirewallRules(t, projectID, networkName, instanceName) {
		t.Fatal("Firewall rule creation failed.")
	}
	defer deleteFirewallRules(t, projectID, instanceName)
	t.Log("Firewall rules created.")

	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running terraform init and apply...")
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Terraform apply complete.")
	t.Log("Fetching Terraform outputs...")
	instanceServiceURLsOutput := terraform.OutputJson(t, terraformOptions, "instance_service_urls")
	instanceServiceURLsMap := gjson.Parse(instanceServiceURLsOutput).Map()
	if len(instanceServiceURLsMap) == 0 {
		t.Fatal("No instances found in 'instance_service_urls' output.")
	}
	t.Logf("Found %d instance outputs for verification.", len(instanceServiceURLsMap))

	maxRetries := 7 // Allow more retries as GAE flex can be slow
	retryInterval := 1 * time.Minute
	verifiedServiceCount := 0

	for instanceKey, serviceURLMapResult := range instanceServiceURLsMap {
		t.Logf("--- Verifying Instance from Output Key: %s ---", instanceKey)
		serviceURLMap := serviceURLMapResult.Map()
		if len(serviceURLMap) == 0 {
			t.Errorf("No service URLs for instance key %s.", instanceKey)
			continue
		}

		var serviceName string
		for sn := range serviceURLMap {
			serviceName = sn
			break
		}
		if serviceName == "" {
			t.Errorf("Could not extract service name for instance key %s.", instanceKey)
			continue
		}
		t.Logf("Verifying service: %s", serviceName)

		var expectedConfig *AppEngineConfig
		for i := range generatedConfigs {
			if generatedConfigs[i].Service == serviceName && generatedConfigs[i].Project == projectID {
				expectedConfig = &generatedConfigs[i]
				break
			}
		}
		if expectedConfig == nil {
			t.Errorf("No matching YAML config for service %s in project %s.", serviceName, projectID)
			continue
		}

		serviceIsReady := false
		versionID := expectedConfig.VersionID
		for i := 0; i < maxRetries; i++ {
			t.Logf("Describing %s/%s (Attempt %d/%d)...", serviceName, versionID, i+1, maxRetries)
			cmd := shell.Command{Command: "gcloud", Args: []string{"app", "versions", "describe", versionID, "--service", serviceName, "--verbosity=none","--project", projectID, "--format", "json"}}
			gcloudOutput, errCmd := shell.RunCommandAndGetOutputE(t, cmd)
			if errCmd != nil {
				t.Logf("gcloud error for %s/%s: %v. Retrying...", serviceName, versionID, errCmd)
				time.Sleep(retryInterval)
				continue
			}
			if gcloudOutput == "" {
				t.Logf("Empty gcloud output for %s/%s. Retrying...", serviceName, versionID)
				time.Sleep(retryInterval)
				continue
			}
			t.Logf("gcloud Output : %s",gcloudOutput)
			actualServiceInfo := gjson.Parse(gcloudOutput)
			t.Logf("Actual Service Info : %s",actualServiceInfo)
			status := gjson.Get(actualServiceInfo.String(),"servingStatus").String()
			t.Logf("Status for %s/%s: %s", serviceName, versionID, status)
			if status == "SERVING" {
				t.Logf("Service %s/%s is SERVING. Performing assertions...", serviceName, versionID)
				assert.Equal(t, expectedConfig.Runtime, actualServiceInfo.Get("runtime").String(), "Runtime mismatch")
				assert.Equal(t, expectedConfig.InstanceClass, actualServiceInfo.Get("instanceClass").String(), "InstanceClass mismatch")
				assert.Equal(t, expectedConfig.ServiceAccount, actualServiceInfo.Get("serviceAccount").String(), "ServiceAccount mismatch")
				assert.Equal(t, expectedConfig.Network.Name, actualServiceInfo.Get("network.name").String(), "Network name mismatch")
				actualSubnetworkPath := actualServiceInfo.Get("network.subnetworkName").String()
				assert.Equal(t, expectedConfig.Network.Subnetwork, actualSubnetworkPath, "Subnetwork name mismatch")

				if expectedConfig.Deployment != nil && expectedConfig.Deployment.Zip != nil {
					t.Logf("Verifying deployment source URL for %s/%s against expected %s (actual field may vary).",
						serviceName, versionID, expectedConfig.Deployment.Zip.SourceURL)
				}

				serviceIsReady = true
				verifiedServiceCount++
				break
			}
			t.Logf("Waiting %v for %s/%s...", retryInterval, serviceName, versionID)
			time.Sleep(retryInterval)
		}
		if !serviceIsReady {
			t.Fatalf("Service %s/%s did not reach SERVING after %d retries.", serviceName, versionID, maxRetries)
		}
	}
	assert.Equal(t, len(generatedConfigs), verifiedServiceCount, "Number of verified services did not match generated configs.")
	t.Logf("Successfully verified %d service(s).", verifiedServiceCount)
	t.Log("Test completed. Cleanup will run via deferred calls.")
}
