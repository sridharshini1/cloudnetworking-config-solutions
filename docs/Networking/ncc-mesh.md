# Network Connectivity Center with Mesh Topology
### Mesh Topology with VPC as spokes

**On this page**

1. [Introduction](#introduction)
2. [Objectives](#objectives)
3. [Architecture](#architecture)
4. [Request flow](#request-flow)
5. [Deploy the solution](#deploy-the-solution)
6. [Prerequisites](#prerequisites)
7. [Deploy through “terraform-cli”](#deploy-through-terraform-cli)
8. [Optional: Delete the deployment](#optional-delete-the-deployment)
9. [Submit feedback](#submit-feedback)

## Introduction

This guide is designed to assist Network and Engineering teams in simplifying their cloud migration journey. Moving to a cloud-first strategy often involves managing complex networks across on-premises and multiple cloud providers. This guide focuses on automating the configuration of Google Cloud's Network Connectivity Center (NCC), a crucial component for simplifying connectivity across these heterogeneous environments.

NCC provides a single management experience for your on-premises and cloud networks, enabling consistent access, policies, and services across global regions. However, managing and configuring NCC manually can be time-consuming and error-prone. This guide leverages automation to streamline the creation of Hubs and Spokes within NCC, accelerating your cloud adoption journey.

This guide assumes familiarity with cloud networking concepts, including Virtual Private Clouds (VPCs), and basic understanding of automation tools. It provides step-by-step instructions and best practices for automating NCC configurations, ultimately improving efficiency and reducing operational overhead.

## Objectives

* Create a Network Connectivity Center Hub: Establish a central hub for managing network connectivity across your Google Cloud projects and on-premises networks.
* Establish VPC Spokes: Create three VPC spokes in different regions (e.g., us-central1, europe-west1, asia-east1) and attach them to the Network Connectivity Center hub.
* Configure Full Mesh Connectivity: Configure a full mesh topology where each VPC spoke can communicate directly with all other spokes.
* Validate Network Connectivity: Deploy workloads (e.g., virtual machines) in each VPC spoke and validate connectivity between them using tools like ping.



## Architecture

**Scenario 1:** The diagram illustrates a multi-network cloud architecture with 3 VPC spokes connected with a hub in a mesh topology.

<img src="./images/nccmeshtopologyvpcspoke.png" alt="nccmeshtopologyvpcspoke" width="800"/>

### Request flow

This scenario involves creating a hub of mesh topology and 3 VPC Spoke connecting to this hub. The architecture diagram depicts a hub-and-spoke network topology, designed to facilitate communication between multiple VPCs (Virtual Private Clouds).

1. Hub: The central point of the network, configured with a mesh topology to enable efficient traffic routing between connected spokes.

2. Spokes: Three VPCs (VPC1, VPC2, VPC3) are connected to the Hub as VPC spokes. Each spoke represents an isolated network environment. These Spokes can be from the same google cloud project or from a different google cloud project.

3. Firewall Rules: To control traffic flow between spokes, firewall rules are implemented on each VPC, allowing or denying ingress and egress traffic based on IP CIDR ranges and ports.

4. Network Connectivity: The architecture relies on Network Connectivity APIs to establish and manage connections between the hub and spokes.

The architecture diagram illustrates 3 different VPC i.e. VPC1, VPC2 and VPC3 connected as VPC spoke to the NCC Hub. The VPC networks can be located across different projects in the same google cloud organization or different organizations. The resources created within this VPC along with appropriate firewall rules would allow connectivity between the different producer and consumer services created in these VPC. For simplicity, we have created 3 GCE instances in each of the VPC. Once the NCC connection is established, each of these GCE private/internal IPs would be reachable to each other.

## Deploy the solution

This section guides you through the process of deploying the solution.

### Prerequisites

To use this configuration solution, ensure the following are installed:

1. **Terraform** : modules are for use with Terraform 1.8+ and tested using Terraform 1.8+. Choose and install the preferred Terraform binary from [here](https://releases.hashicorp.com/terraform/).
2. **gcloud SDK** : install gcloud SDK from [here](https://cloud.google.com/sdk/docs/install) to authenticate to Google Cloud while running Terraform.

### Deploy through terraform-cli

1. **Clone** the cloudnetworking-config-solutions repository repository**:**
    ```
    git clone https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions.git
    ```

2. Navigate to **cloudnetworking-config-solutions** folder and update the files containing the configuration values
   * **00-bootstrap stage**
     * Update configuration/bootstrap.tfvars **\-** update the google cloud project IDs and the user IDs/groups in the tfvars.

        ```
            folder_id                           = "<your-project-id>"
            bootstrap_project_id                = "<your-project-id>"
            network_hostproject_id              = "<your-project-id>"
            network_serviceproject_id           = "<your-project-id>"
            organization_administrator          = ["user:user-example@example.com"]
            networking_administrator            = ["user:user-example@example.com"]
            security_administrator              = ["user:user-example@example.com"]
            producer_cloudsql_administrator     = ["user:user-example@example.com"]
            producer_gke_administrator          = ["user:user-example@example.com"]
            producer_alloydb_administrator      = ["user:user-example@example.com"]
            producer_vertex_administrator       = ["user:user-example@example.com"]
            producer_mrc_administrator          = ["user:user-example@example.com"]
            producer_connectivity_administrator = ["user:user-example@example.com"]
            consumer_gce_administrator          = ["user:user-example@example.com"]
            consumer_cloudrun_administrator     = ["user:user-example@example.com"]
        ```

   * **01-organisation stage**
     * Update configuration/organization.tfvars \- update the google cloud project ID and the list of the APIs to enable the service networking API.

        ```
        activate_api_identities = {
          "project-01" = {
            project_id = "your-project-id",
            activate_apis = [
              "servicenetworking.googleapis.com",
              "iam.googleapis.com",
              "compute.googleapis.com",
              ],
          },
        }
        ```

   * **02-networking stage**
     * Update configuration/networking.tfvars \- update the Google Cloud Project ID and the parameters for additional resources such as VPC, subnet, NAT, create_NCC flag as outlined below and based on your requriements.

        ```
        project_id  = "your-project-id",
        region      = "us-central1"

        ## VPC input variables
        network_name = "CNCS_VPC"
        subnets = [
          {
            ip_cidr_range = "10.0.0.0/24"
            name          = "CNCS_VPC_Subnet_1"
            region        = "us-central1-a"
          }
        ]
        psa_range_name    = range1
        psa_range         = "10.0.64.0/20"

        ## PSC/Service Connectivity Variables
        create_scp_policy  = false

        ## Cloud Nat input variables
        create_nat = true
        ## Cloud HA VPN input variables
        create_havpn = false

        ## NCC input variables
        create_ncc = true
        ```
3. **Execute the terraform script**
   You can now deploy the stages individually using **run.sh** or you can deploy all the stages automatically using the run.sh file. Navigate to the execution/ directory and run this command to run the automatic deployment using **run.sh .**

    ```
    ./run.sh -s networking -t init-apply-auto-approve
    or
    ./run.sh --stage networking --tfcommand init-apply-auto-approve
    ```

4. **Verify NCC resource creation:**
   Once the deployment is complete, navigate to the network connectivity center section in the Google Cloud Console to confirm that your network connectivity center resources has been successfully created.

   Your network connectivity center hub and spoke are now ready to serve different producers and consumers.

## **Optional-Delete the deployment**

1. In Cloud Shell or in your terminal, make sure that the current working directory is $HOME/cloudshell\_open/\<Folder-name\>/execution. If it isn't, go to that directory.
2. Remove the resources that were provisioned by the solution guide:

    ```
    ./run.sh -s all -t destroy-auto-approve
    ```

Terraform displays a list of the resources that will be destroyed.

3. When you're prompted to perform the actions, enter yes.

## **Submit feedback**

To troubleshoot errors, check Terraform's logs and output.

To submit feedback, do the following:

* If you're looking for assistance with streamlining network configuration automation for a comparable use case, feel free to submit an issue on the [GitHub repository](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/issues).
* For unmodified Terraform code, create issues in the [GitHub repository](https://github.com/GoogleCloudPlatform/cloudnetworking-config-solutions/issues). GitHub issues are reviewed on a best-effort basis and are not intended for general use questions.
* For issues with the products that are used in the solution, contact [Cloud Customer Care](https://cloud.google.com/support-hub).
