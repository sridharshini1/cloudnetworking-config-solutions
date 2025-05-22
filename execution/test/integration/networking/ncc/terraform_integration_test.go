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

const (
	region       = "us-west2"
	yamlFileName = "instance.yaml"
)

var (
	projectID          = os.Getenv("TF_VAR_project_id")
	uniqueID           = rand.Int()
	networkName        = fmt.Sprintf("test-vpc-ncc-%d", uniqueID)
	subnetworkName     = fmt.Sprintf("test-subnet-ncc-%d", uniqueID)
	subnetworkIPCIDR   = "10.0.2.0/24"
	psaRangeName       = "testpsarange-ncc-1"
	psaRange           = "10.0.64.0/20"
	testHubName        = fmt.Sprintf("ncc-hub-test-%d", uniqueID)
	testHubDescription = "Test NCC Hub for integration"
	testHubLabels      = map[string]string{
		"environment": "testing",
	}
	testSpokeLabels = map[string]string{
		"team": "testing",
	}

	testVPCSpokeName      = "spoke1-test"
	testProducerSpokeName = "prodspoke1-test"
	groupName             = "default"
	groupDescription      = "Test group"

	// Default values for variables
	defaultExportPSC          = true
	defaultPolicyMode         = "PRESET"
	defaultPresetTopology     = "MESH"
	defaultAutoAcceptProjects = []string{}
	projectRoot, _            = filepath.Abs("../../../../")
	terraformNCCDirectoryPath = filepath.Join(projectRoot, "02-networking/NCC")
	configFolderPathNCC       = filepath.Join(projectRoot, "test/integration/networking/ncc/config")
)

// NCCConfig struct to match the new YAML structure
type NCCConfig struct {
	Hubs   []HubConfig   `yaml:"hubs"`
	Spokes []SpokeConfig `yaml:"spokes"`
}

type HubConfig struct {
	Name               string            `yaml:"name"`
	ProjectID          string            `yaml:"project_id"`
	Description        string            `yaml:"description"`
	Labels             map[string]string `yaml:"labels"`
	ExportPSC          bool              `yaml:"export_psc"`
	PolicyMode         string            `yaml:"policy_mode"`
	PresetTopology     string            `yaml:"preset_topology"`
	AutoAcceptProjects []string          `yaml:"auto_accept_projects"`
	CreateNewHub       bool              `yaml:"create_new_hub"`
	ExistingHubURI     string            `yaml:"existing_hub_uri"`
	GroupName          string            `yaml:"group_name"`
	GroupDescription   string            `yaml:"group_decription"`
	SpokeLabels        map[string]string `yaml:"spoke_labels"`
}

type SpokeConfig struct {
	Type                string            `yaml:"type"`
	Name                string            `yaml:"name"`
	ProjectID           string            `yaml:"project_id"`
	Location            string            `yaml:"location,omitempty"`
	URI                 string            `yaml:"uri"`
	Description         string            `yaml:"description"`
	Labels              map[string]string `yaml:"labels"`
	Peering             string            `yaml:"peering,omitempty"`
	ExcludeExportRanges []string          `yaml:"exclude_export_ranges,omitempty"`
	IncludeExportRanges []string          `yaml:"include_export_ranges,omitempty"`
}

func TestNCC(t *testing.T) {
	if projectID == "" {
		t.Skip("Skipping test because TF_VAR_project_id is not set.")
	}

	// Setup: create YAML config and VPC/subnet/PSA
	createConfigYAMLNCC(t, true, "", false, testHubName)
	createVPCAndSubnetWithPSA(t, projectID, networkName, subnetworkName, region, psaRangeName, psaRange)

	tfVars := map[string]interface{}{
		"config_folder_path":   configFolderPathNCC,
		"create_new_hub":       true,
		"existing_hub_uri":     nil,
		"export_psc":           defaultExportPSC,
		"policy_mode":          defaultPolicyMode,
		"preset_topology":      defaultPresetTopology,
		"auto_accept_projects": append(defaultAutoAcceptProjects, projectID),
		"ncc_hub_description":  testHubDescription,
		"ncc_hub_labels":       testHubLabels,
		"spoke_labels":         testSpokeLabels,
		"group_name":           groupName,
		"group_decription":     groupDescription,
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		Vars:         tfVars,
		TerraformDir: terraformNCCDirectoryPath,
		Reconfigure:  true,
		Lock:         true,
		NoColor:      true,
	})

	defer deleteVPCAndSubnet(t, projectID, networkName, subnetworkName, region)
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)
	time.Sleep(30 * time.Second)

	verifyNCCResources(t, terraformOptions, testHubName)
}

// createConfigYAMLNCC creates the configuration YAML file for NCC.
func createConfigYAMLNCC(t *testing.T, createNewHub bool, existingHubURI string, existingSpoke bool, nccHubName string) {
	t.Helper()

	spokes := []SpokeConfig{
		{
			Type:        "linked_vpc_network",
			Name:        testVPCSpokeName,
			ProjectID:   projectID,
			URI:         fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName),
			Description: "Test VPC Spoke",
			Labels:      testSpokeLabels,
		},
		{
			Type:                "linked_producer_vpc_network",
			Name:                testProducerSpokeName,
			ProjectID:           projectID,
			Location:            "global",
			URI:                 fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName),
			Peering:             "servicenetworking-googleapis-com",
			ExcludeExportRanges: []string{},
			IncludeExportRanges: []string{psaRange},
			Labels:              testSpokeLabels,
			Description:         "Test Producer VPC Spoke",
		},
	}
	hubs := []HubConfig{
		{
			Name:               nccHubName,
			ProjectID:          projectID,
			Description:        testHubDescription,
			Labels:             testHubLabels,
			ExportPSC:          defaultExportPSC,
			PolicyMode:         defaultPolicyMode,
			PresetTopology:     defaultPresetTopology,
			AutoAcceptProjects: append(defaultAutoAcceptProjects, projectID),
			CreateNewHub:       createNewHub,
			ExistingHubURI:     existingHubURI,
			SpokeLabels:        testSpokeLabels,
			GroupName:          groupName,
			GroupDescription:   groupDescription,
		},
	}
	nccInstance := NCCConfig{
		Hubs:   hubs,
		Spokes: spokes,
	}

	yamlData, err := yaml.Marshal(&nccInstance)
	if err != nil {
		t.Fatalf("Error while marshaling: %v", err)
	}

	if err := os.MkdirAll(configFolderPathNCC, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	filePath := filepath.Join(configFolderPathNCC, yamlFileName)
	t.Logf("Created YAML config at %s with content:\n%s", filePath, string(yamlData))

	err = os.WriteFile(filePath, []byte(yamlData), 0644)
	if err != nil {
		t.Fatalf("Unable to write data into the file: %v", err)
	}
}

// createVPCAndSubnetWithPSA creates a VPC, a subnet with PSA enabled.
func createVPCAndSubnetWithPSA(t *testing.T, projectID, networkName, subnetworkName, region, psaRangeName, psaRange string) {
	t.Helper()
	text := "compute"
	// Create VPC
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", networkName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom", "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating VPC: %v", err)
	}

	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "create", subnetworkName,
			"--project=" + projectID,
			"--network=" + networkName,
			"--region=" + region,
			"--range=" + subnetworkIPCIDR,
			"--enable-private-ip-google-access",
			"--quiet",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating subnet: %v", err)
	}

	// Create Allocated PSA Range
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "addresses", "create", psaRangeName,
			"--global",
			"--purpose=VPC_PEERING",
			"--addresses=" + psaRange[:strings.Index(psaRange, "/")],       // Extract the IP from the CIDR
			"--prefix-length=" + psaRange[strings.Index(psaRange, "/")+1:], // Extract the prefix length from the CIDR
			"--network=" + networkName,
			"--project=" + projectID,
			"--quiet",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating allocated PSA range: %v", err)
	}

	// Create PSA Range Peering
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"services", "vpc-peerings", "connect",
			"--project=" + projectID,
			"--service=servicenetworking.googleapis.com",
			"--ranges=" + psaRangeName,
			"--network=" + networkName,
			"--quiet",
		},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("Error creating PSA range: %v", err)
	}

	t.Logf("Successfully created VPC '%s' with subnet '%s' and PSA range '%s'.", networkName, subnetworkName, psaRangeName)
	time.Sleep(60 * time.Second)
}

func deleteVPCAndSubnet(t *testing.T, projectID, networkName, subnetworkName, region string) {
	t.Helper()
	text := "compute"
	time.Sleep(60 * time.Second) // Wait for resources to be in a deletable state

	// Delete subnet
	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			text, "networks", "subnets", "delete", subnetworkName,
			"--project=" + projectID,
			"--region=" + region,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Error deleting subnet: %v", err)
	}

	time.Sleep(60 * time.Second)

	// Delete PSA range
	cmd = shell.Command{
		Command: "gcloud",
		Args: []string{
			"compute", "addresses", "delete", psaRangeName,
			"--global",
			"--project=" + projectID,
			"--quiet",
		},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Error deleting PSA range: %v", err)
	}

	time.Sleep(60 * time.Second)

	// Delete VPC
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	if _, err := shell.RunCommandAndGetOutputE(t, cmd); err != nil {
		t.Errorf("Error deleting VPC: %v", err)
	}

	time.Sleep(60 * time.Second)
}

func verifyNCCResources(t *testing.T, terraformOptions *terraform.Options, testHubName string) {
	t.Helper()
	nccOutputValue := terraform.OutputJson(t, terraformOptions, "ncc_module")
	if !gjson.Valid(nccOutputValue) {
		t.Fatalf("Error parsing network_connectivity_center output, invalid json: %s", nccOutputValue)
	}
	resultNCC := gjson.Parse(nccOutputValue)

	// Use the hub name as the key
	hubKey := testHubName

	// Verify NCC Hub
	hubIDPath := fmt.Sprintf("%s.ncc_hub.0.id", hubKey)
	gotHubID := gjson.Get(resultNCC.String(), hubIDPath).String()
	wantHubID := fmt.Sprintf("projects/%s/locations/global/hubs/%s", projectID, testHubName)
	if gotHubID != wantHubID {
		t.Errorf("Hub with invalid ID created. Got: %v, Want: %v", gotHubID, wantHubID)
	} else {
		t.Logf("Verified NCC Hub ID: %s", gotHubID)
	}

	hubStatePath := fmt.Sprintf("%s.ncc_hub.0.state", hubKey)
	gotHubState := gjson.Get(resultNCC.String(), hubStatePath).String()
	wantHubState := "ACTIVE"
	if gotHubState != wantHubState {
		t.Errorf("Hub with invalid state created. Got: %v, Want: %v", gotHubState, wantHubState)
	} else {
		t.Logf("Verified NCC Hub State: %s", gotHubState)
	}

	// Verify VPC Spoke
	vpcSpokePath := fmt.Sprintf("%s.vpc_spokes.%s", hubKey, testVPCSpokeName)
	vpcSpoke := gjson.Get(resultNCC.String(), vpcSpokePath)
	if !vpcSpoke.Exists() {
		t.Errorf("VPC Spoke '%s' not found in NCC output", testVPCSpokeName)
	} else {
		// Get the first linked_vpc_network.uri
		gotVPCSpokeURI := vpcSpoke.Get("linked_vpc_network.0.uri").String()
		wantVPCSpokeURI := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
		if !strings.HasSuffix(gotVPCSpokeURI, wantVPCSpokeURI) {
			t.Errorf("VPC Spoke '%s' with invalid URI. Got: %v, Want: %v", testVPCSpokeName, gotVPCSpokeURI, wantVPCSpokeURI)
		} else {
			t.Logf("Verified NCC VPC Spoke '%s' URI: %s", testVPCSpokeName, gotVPCSpokeURI)
		}
	}

	// Verify Producer VPC Spoke
	producerSpokePath := fmt.Sprintf("%s.producer_vpc_spokes.%s", hubKey, testProducerSpokeName)
	producerSpoke := gjson.Get(resultNCC.String(), producerSpokePath)
	if !producerSpoke.Exists() {
		t.Errorf("Producer VPC Spoke '%s' not found in NCC output", testProducerSpokeName)
	} else {
		gotProducerSpokeURI := producerSpoke.Get("linked_producer_vpc_network.0.network").String()
		wantProducerSpokeURI := fmt.Sprintf("projects/%s/global/networks/%s", projectID, networkName)
		if !strings.HasSuffix(gotProducerSpokeURI, wantProducerSpokeURI) {
			t.Errorf("Producer VPC Spoke '%s' with invalid URI. Got: %v, Want: %v", testProducerSpokeName, gotProducerSpokeURI, wantProducerSpokeURI)
		} else {
			t.Logf("Verified NCC Producer VPC Spoke '%s' URI: %s", testProducerSpokeName, gotProducerSpokeURI)
		}
	}

	t.Log("NCC Resources verification finished.")
}
