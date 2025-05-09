# Consumer Stage

## Overview

This Consumer stage is responsible for provisioning consumer service instances such as Google Compute Engine (GCE) virtual machines and Cloud Run Jobs & Cloud Run Services. It uses Terraform modules to manage the creation and configuration of consumers such as VMs based on input parameters defined in YAML files.

The stage is designed to be highly flexible. For GCE, it allows customizations such as instance type, boot disk, network configuration, and attached storage.


## Prerequisites

- **Completed Prior Stages:** Successful deployment of networking resources depends on the completion of the following stages:

    - **01-organization:** This stage handles the activation of required Google Cloud APIs.
    - **02-networking:** This stage handles the creation of networking resources such as VPCs, HA-VPNs etc.
    - **03-security:** This stage handles the creation of key security components such firewall rules. 
            - For GCE, the folder to use is 03-security/GCE.
            - For MIG, the folder to use is 03-security/MIG.
- Enable the following APIs :
    - [Compute Engine API](https://cloud.google.com/compute/docs/reference/rest/v1): Used for creating and managing GCE VMs and MIGs.
    - [Cloud Run API](https://cloud.google.com/run/docs/reference/rest): Used for creating and managing Cloud Run jobs and Cloud Run services.
    - [AI Platform Notebooks API](https://cloud.google.com/vertex-ai/docs/workbench/reference/rest): Required for provisioning and managing Workbench environments.
    - [App Engine Admin API](https://cloud.google.com/appengine/docs/admin-api): Required for creating and managing App Engine services.
        - In addition to the above api , you might also need to enable the following apis for it to work perfectly
            - [Artifact Registry API](https://cloud.google.com/artifact-registry/docs/reference/rest): Used for storing and managing build artifacts.
            - [Cloud Build API](https://cloud.google.com/cloud-build/docs/api/reference/rest): Used for building source code.
            - [Cloud Logging API](https://cloud.google.com/logging/docs/reference/v2/rest): Used for writing and managing logs.
            - [Cloud Storage API](https://cloud.google.com/storage/docs/json_api/v1): Used for storing application source code, build artifacts, and other data.

    - [Serverless VPC Access API](https://cloud.google.com/vpc/docs/reference/vpcaccess/rest): Required for creating and managing Serverless VPC Access connectors.
                
- Permissions required :

    - [Compute Admin role](https://cloud.google.com/compute/docs/access/iam#compute.admin) : Used to create and manage GCE VMs and MIGs.
    - [Service Account User](https://cloud.google.com/compute/docs/access/iam#iam.serviceAccountUser) : Lets a principal attach a service account to a resource.
    - [Cloud Run Admin](https://cloud.google.com/run/docs/reference/iam/roles#run.admin) : User to create and manage cloud run jobs and cloud run services.
    - [App Engine Admin role](https://cloud.google.com/appengine/docs/standard/roles#appengine.admin): Required for full control over App Engine applications, including creation and management.
        - In addition to the above role , you might also need to enable the following roles for it to work perfectly
            - [Cloud Build Editor role](https://cloud.google.com/build/docs/iam-roles-permissions): Allows starting builds and viewing build information.
            - [Artifact Registry Writer role](https://cloud.google.com/artifact-registry/docs/access-control#roles): Allows pulling container images or other artifacts.
            - [Artifact Registry Reader role](https://cloud.google.com/artifact-registry/docs/access-control#roles): Allows pulling container images or other artifacts.
            - [Storage Object Admin role](https://cloud.google.com/storage/docs/access-control/iam-roles#standard-roles): Allows managing Cloud Storage buckets, which App Engine uses for storing source code and created container images.
            - [Logs Writer role](https://cloud.google.com/logging/docs/access-control#roles): Allows services like App Engine to write logs.
    - [Serverless VPC Access Admin role](https://cloud.google.com/vpc/docs/configure-serverless-vpc-access#permissions): Required to create, delete, and manage Serverless VPC Access connectors.
    - [Notebooks Admin role](https://cloud.google.com/vertex-ai/docs/workbench/security/iam#roles): Required for full control over Vertex AI Workbench instances.

## Configuration

### General Configuration Notes

- YAML Configuration Files: Place YAML files defining each instance's configuration within the configs/ directory of the respective service's folder (e.g., configuration/consumer/GCE/config).

- Terraform Variables: You can customize the input variables in the .tf files according to your project's requirements.

Configurations would be different for different consumer services as listed below :
1. **GCE**: For configuration of the GCE VM, you can read more in the [GCE README](cloudnetworking-config-solution/execution/06-consumer/GCE/README.md).

2. **Cloud Run**: For configuration of the Cloud Run Job and Service, you can read more in the following sections:

    2.1. [Cloud Run Job README](cloudnetworking-config-solution/execution/06-consumer/CloudRun/Job/README.md).

    2.2. [Cloud Run Service README](cloudnetworking-config-solution/execution/06-consumer/CloudRun/Service/README.md).

3. **MIG**: For configuration of an MIG, you can read more in the [MIG README](cloudnetworking-config-solution/execution/06-consumer/MIG/README.md).

4. **Workbench**: For configuration of the Workbench environment, you can read more in the [Workbench README](cloudnetworking-config-solution/execution/06-consumer/Workbench/README.md).

5. **App Engine**: For configuration of the App Engine service, you can read more in the following sections:

    5.1. [App Engine Standard README](cloudnetworking-config-solution/execution/06-consumer/Serverless/AppEngine/Flexible/README.md).

    5.2. [App Engine Flexible README](cloudnetworking-config-solution/execution/06-consumer/Serverless/AppEngine/Standard/README.md).

6. **VPC Access Connector**: For configuration of the VPC Access Connector, you can read more in the [VPC Access Connector README](cloudnetworking-config-solution/execution/06-consumer/Serverless/VPCAccessConnector/README.md).

For every consumer, you can define .yaml files for the consumer configuration. With every .yaml file in the configs/ folder, our terraform module would create an instance. For an example, for GCE an example yaml files to create two instances are :

- instance1.yaml :

  ```
  name: instance1
  project_id: <project-id>
  region: us-central1
  zone: us-central1-a
  image: ubuntu-os-cloud/ubuntu-2204-lts
  network: projects/<project-id>/global/networks/<network-name>
  subnetwork: projects/<project-id>/regions/us-central1/subnetworks/<subnetwork-name>
  ```

- instance2.yaml :

  ```
  name: instance2
  project_id: <project-id>
  region: us-central1
  zone: us-central1-a
  image: ubuntu-os-cloud/ubuntu-2204-lts
  network: projects/<project-id>/global/networks/<network-name>
  subnetwork: projects/<project-id>/regions/us-central1/subnetworks/<subnetwork-name>
  ```

## Execution Steps

1. **Input/Configure** the yaml files based on your requirements.

2. **Terraform Stages** :

    - Initialize: Run `terraform init`.
    - Plan: Run `terraform plan` to review the planned changes.
    - Apply:  If the plan looks good, run `terraform apply` to create or update the resources.


## Additional Notes

- **Instance configuration**: Carefully review and customize the instance configuration to match your organization's requirements.