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

package common_utils

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/shell"
	"testing"
	"time"
)

/*
CreateVPCSubnets is a helper function which creates the VPC and subnets before
execution of the test expecting to use existing VPC and subnets.
*/

func CreateVPCSubnets(t *testing.T, projectID string, networkName string, subnetworkName string, region string) {
	subnetworkIPCIDR := "10.0.1.0/24"
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "create", networkName, "--project=" + projectID, "--format=json", "--bgp-routing-mode=global", "--subnet-mode=custom", "--verbosity=none"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
	time.Sleep(60 * time.Second)
	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "subnets", "create", subnetworkName, "--network=" + networkName, "--project=" + projectID, "--range=" + subnetworkIPCIDR, "--region=" + region, "--format=json", "--enable-private-ip-google-access", "--enable-flow-logs", "--verbosity=none"},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
}

/*
DeleteVPCSubnets is a helper function which deletes the VPC and subnets after
completion of the test expecting to use existing VPC and subnets.
*/
func DeleteVPCSubnets(t *testing.T, projectID string, networkName string, subnetworkName string, region string) {
	text := "compute"
	cmd := shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "subnets", "delete", subnetworkName, "--region=" + region, "--project=" + projectID, "--quiet"},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}

	// Sleep for 60 seconds to ensure the deleted subnets is reliably reflected.
	time.Sleep(60 * time.Second)

	cmd = shell.Command{
		Command: "gcloud",
		Args:    []string{text, "networks", "delete", networkName, "--project=" + projectID, "--quiet"},
	}
	_, err = shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("===Error %s Encountered while executing %s", err, text)
	}
}

/*
CreateServiceConnectionPolicy is a helped function that creates the service
connection policy.
*/
func CreateServiceConnectionPolicy(t *testing.T, projectID string, region string, networkName string, policyName string, subnetworkName string, serviceClass string, connectionLimit int) {
	// Get subnet self link from subnet ID using gcloud command
	subnetSelfLink := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/subnetworks/%s", projectID, region, subnetworkName)

	cmd := shell.Command{
		Command: "gcloud",
		Args: []string{
			"network-connectivity", "service-connection-policies", "create",
			policyName, // Add the policyName here as the first argument after "create"
			"--project", projectID,
			"--region", region,
			"--network", networkName,
			"--service-class", serviceClass,
			"--subnets", subnetSelfLink,
			"--psc-connection-limit", fmt.Sprintf("%d", connectionLimit),
			"--quiet",
		},
	}
	_, err := shell.RunCommandAndGetOutputE(t, cmd)
	if err != nil {
		t.Errorf("Error creating Service Connection Policy: %s", err)
	}
}
