# Consumer Load Balancing Stage

## Overview

The Consumer Load Balancing Stage provides a comprehensive solution for deploying and managing various types of load balancers on Google Cloud Platform (GCP). This stage encompasses both **Application Load Balancers** and **Network Load Balancers**, allowing for flexible traffic distribution based on your application needs. Whether you require internal or external load balancing, this setup ensures optimal performance, scalability, and availability for your applications.

## Prerequisites

Before proceeding with the load balancing stage, ensure that the following prerequisites are met:

- **Google Cloud Project**: You must have an active Google Cloud project.
- **Enabled APIs**:
  - Compute Engine API
  - Cloud Network Connectivity API
- **Permissions**: The user or service account executing Terraform must have the following roles:
  - Load Balancer Admin (for managing load balancers)
- **Backends for Load Balancers** : Please create a backend such as a Managed Instance Group (MIG) for your Load Balancer.

## Configuration

### General Configuration Notes

- **YAML Configuration Files**: Load balancer settings are defined in YAML configuration files. These files specify parameters such as backend services, health checks, logging options, and network configurations.
- **Variable Management**: Utilize the `variables.tf` file to define and manage variables that can be reused across different configurations, ensuring consistency and ease of updates.
- **Local Variables**: The `locals.tf` file is used to process the YAML configurations and create a structured representation of your load balancer instances.

## Execution Steps

1. **Input/Configure** the YAML files based on your requirements.
   - Modify the provided sample YAML configuration files to reflect your specific project details, such as project ID, target tags, firewall networks, and backend settings.

2. **Terraform Stages**:

    - **Initialize**: Run the following command to initialize Terraform in your working directory:

      ```bash
      terraform init
      ```

    - **Plan**: Review the planned changes before applying them by running:

      ```bash
      terraform plan
      ```
      This command will show you what resources will be created or modified.

    - **Apply**: If the plan looks good, execute the following command to create or update the resources:
    
      ```bash
      terraform apply
      ```

## Additional Notes

- **Load Balancer Configuration**: Carefully review and customize the load balancer configuration to match your organization's requirements. Pay particular attention to backend service settings, health checks, logging configurations, and network settings.
- **Types of Load Balancers**:
  - **Application Load Balancers**: Best for HTTP(S) traffic; operates at Layer 7 and provides advanced routing capabilities.
  - **Network Load Balancers**: Ideal for TCP/UDP traffic; operates at Layer 4 and supports various IP protocols.
  - Choose between internal or external load balancers based on your application's architecture and traffic sources.
- **Testing**: It is recommended to test your load balancer configuration in a staging environment before deploying it to production to ensure that it behaves as expected under load.

By following this README, you will be able to successfully configure and deploy both Application and Network Load Balancers using Terraform in Google Cloud.