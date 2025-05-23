// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integrationtest

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"math/rand"
	"os"
	"testing"
	"time"
)

var (
	projectID                     = os.Getenv("TF_VAR_project_id")
	region                        = "us-central1"
	global                        = "global"
	terraformDirectoryPath        = "../../../../../03-security/Firewall/FirewallPolicy"
	configFolderPath              = "../../../test/integration/security/Firewall/FirewallPolicy/config"
	uniqueIdentifier              = fmt.Sprint(rand.Int())
	networkName                   = fmt.Sprintf("vpc-%s-test", uniqueIdentifier)
	regionalFirewallPolicy        = fmt.Sprintf("regionalfirewallpolicy-%s", "test")
	globalFirewallPolicy          = fmt.Sprintf("globalfirewallpolicy-%s", "test")
	regionalFirewallPolicyVPCName = fmt.Sprintf("regionalfirewallpolicyvpc-%s", "test")
	globalFirewallPolicyVPCName   = fmt.Sprintf("globalfirewallpolicyvpc-%s", "test")
)

type FirewallPolicyStruct struct {
	Name        string            `yaml:"name"`
	ParentID    string            `yaml:"parent_id"`
	Region      string            `yaml:"region"`
	Attachments map[string]string `yaml:"attachments"`
}

type AttachmentStruct struct {
	VPC interface{} `yaml:"vpc"`
}

/*
This test creates all the pre-requsite resources including the vpc network
It then validates if Regional and Global network firewall policies are created and validates the same.
*/
func TestCreateFirewallPolicy(t *testing.T) {
	// Initialize Network Firewall Policy config YAML files
	createConfigYAMLs(t, region, projectID, regionalFirewallPolicyVPCName, regionalFirewallPolicy)
	createConfigYAMLs(t, global, projectID, globalFirewallPolicyVPCName, globalFirewallPolicy)

	var (
		tfVars = map[string]any{
			"config_folder_path": configFolderPath,
		}
	)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// Set the path to the Terraform code that will be tested.
		Vars:                 tfVars,
		TerraformDir:         terraformDirectoryPath,
		Reconfigure:          true,
		Lock:                 true,
		NoColor:              true,
		SetVarsAfterVarFiles: true,
	})
	// Create VPC outside of the terraform module.
	vpcNameList := []string{globalFirewallPolicyVPCName, regionalFirewallPolicyVPCName}
	for _, vpcName := range vpcNameList {
		err := createVPC(t, projectID, vpcName)
		if err != nil {
			t.Fatalf("Error parsing output: %s", err)
		}
	}

	// Delete VPC created outside of the terraform module.
	for _, vpcName := range vpcNameList {
		defer deleteVPC(t, projectID, vpcName)
	}

	// Clean up resources with "terraform destroy" at the end of the test.
	defer terraform.Destroy(t, terraformOptions)

	// Run "terraform init" and "terraform apply". Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)
	t.Log("Waiting for 60 seconds to let resource achieve stable state")
	// Wait for 60 seconds to let resource acheive stable state.
	time.Sleep(60 * time.Second)
	t.Log("Waiting for 60 seconds to let resource achieve stable state")
	// Run `terraform output` to get the values of output variables
	firewallPolicyOutputValue := terraform.OutputJson(t, terraformOptions, "id")
	if !gjson.Valid(firewallPolicyOutputValue) {
		t.Fatalf("Error parsing output, invalid JSON: %s", firewallPolicyOutputValue)
	}

	result := gjson.Parse(firewallPolicyOutputValue)
	regionalFirewallIDPath := fmt.Sprintf("%s.id", regionalFirewallPolicy)
	globalFirewallIDPath := fmt.Sprintf("%s.id", globalFirewallPolicy)

	wantRegionalFirewallID := fmt.Sprintf("projects/%s/regions/%s/firewallPolicies/%s", projectID, region, regionalFirewallPolicy)
	wantGlobalFirewallID := fmt.Sprintf("projects/%s/%s/firewallPolicies/%s", projectID, global, globalFirewallPolicy)

	gotRegionalFirewallID := gjson.Get(result.String(), regionalFirewallIDPath).String()
	gotGlobalFirewallID := gjson.Get(result.String(), globalFirewallIDPath).String()
	if wantRegionalFirewallID != gotRegionalFirewallID {
		t.Errorf("Firewall with invalid regional ID Path created = %v, want = %v", gotRegionalFirewallID, wantRegionalFirewallID)
	}

	if wantGlobalFirewallID != gotGlobalFirewallID {
		t.Errorf("Firewall with invalid global ID Path created = %v, want = %v", gotGlobalFirewallID, wantGlobalFirewallID)
	}
}

// /*
// deleteVPC is a helper function which deletes the VPC after
// completion of the test.
// */
func deleteVPC(t *testing.T, projectID string, networkName string) {
	text := "compute"
	t.Log("Waiting for 60 seconds to let resource achieve stable state")
	time.Sleep(60 * time.Second)
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("===Error %s Encountered while executing %s", err, text)
	}
}

// /*
// createVPC is a helper function which creates the VPC before the
// execution of the test.
// */
func createVPC(t *testing.T, projectID string, networkName string) error {
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", networkName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Fatalf("===Error %s Encountered while executing %s", err, text)
	}
	return err
}

/*
createConfigYAML is a helper function which creates the configuration YAML file
for an network firewall policy instance range before the.
*/
func createConfigYAMLs(t *testing.T, region string, projectID string, vPCName string, firewallPolicyType string) {

	instance := FirewallPolicyStruct{
		Name:     fmt.Sprintf("%s", firewallPolicyType),
		ParentID: projectID,
		Region:   region,
		Attachments: map[string]string{
			"vpc": fmt.Sprintf("projects/%s/global/networks/%s", projectID, vPCName),
		},
	}

	yamlData, err := yaml.Marshal(&instance)
	if err != nil {
		t.Errorf("Error marshalling instance for %s: %v", firewallPolicyType, err)
	}
	filePath := fmt.Sprintf("%s/%s-%s", "config", firewallPolicyType, "instance.yaml")
	err = os.WriteFile(filePath, []byte(yamlData), 0666)
	if err != nil {
		t.Errorf("Unable to write instance data for %s: %v", filePath, err)
	}
}
