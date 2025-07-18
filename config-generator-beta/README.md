### Terraform Config Generator [Beta] - User Guide

Welcome to the Terraform Config Generator\! This tool is designed to simplify the process of setting up complex Google Cloud networking architectures. Instead of editing numerous configuration files, you can now declare your high-level intent in a single, simple JSON file, and this tool will generate all the necessary, ready-to-deploy Terraform configurations for you.

Config Generator is a one-way one-time automation, any changes made to generated configuration files (`.yaml` or `.tfvars`) are not reflected back in the architecture-spec JSON file. This is also a one-time automation tool - any required overlapping changes to configuration needs existing infrastructure to be destroyed and configuration/ folder to be regenerated.

### Prerequisites

Before running the generator, please ensure your environment meets the following requirements:

1.  **Python 3.10 or newer:** The generator scripts use syntax and features available in Python 3.10 and later. You can download the latest version from the [official Python website](https://www.python.org/downloads/).

2.  **Terraform:** The generated files are Terraform configurations. Ensure the Terraform CLI is installed on your system. You can find installation instructions on the [official Terraform website](https://developer.hashicorp.com/terraform/downloads).

3.  **Install Python Dependencies:** The script requires a few Python libraries. It is highly recommended to use a virtual environment to manage these dependencies.

    ```bash
    # It is best practice to use a virtual environment
    python3 -m venv venv
    source venv/bin/activate

    # Install the required packages
    python3 -m pip install absl-py Jinja2
    ```

4.  **Set Default Applications:** For a smoother interactive experience, it's helpful to have a default application set for opening `.tfvars` and `.yaml` files. The script will attempt to open these files for your review, and this ensures it works seamlessly.

5.  **Check for an Existing `configuration` Directory:**

    > **Important [Beta] Notice:**
    > **Warning:** The generator creates a new `configuration/` directory. If you already have a folder with this name from a previous run, you must rename or remove it before starting.
    > If that directory was used to deploy resources, **you must run `terraform destroy`** from within that configuration's deployment folders before removing it to avoid orphaning cloud resources.

-----

#### The `architecture-spec` File

The core of the generator is the `architecture-spec` file (in `.json` format). This is where you define the architecture you want to build.

##### Top-Level Structure

Every `architecture-spec` file has a few keys at the top level:

  * **`description`** : A clear, detailed description of the architecture, its purpose, and its business value.
  * **`defaultRegion`** (Optional): A default GCP region (e.g., "us-central1") to use for resources if a region is not specified on the resource itself.
  * **`namePrefix`** (Optional): A string that will be prepended to the `name` of every generated resource. Useful for distinguishing between different deployments.
  * **`nameSuffix`** (Optional): A string that will be appended to the `name` of every generated resource.
  * **`projects`** (Required): A list of all the GCP projects involved in this architecture.

<!-- end list -->

```json
{
  "description": "A detailed description of the architecture and its purpose...",
  "defaultRegion": "us-central1",
  "namePrefix": "demo-",
  "nameSuffix": "-test",
  "projects": [
    // Project definitions go here
  ]
}
```

##### Defining Projects and the VPC Host

The `projects` key holds a list of project objects. Each project must have a `projectId`.

> **Note:** Projects are not created by this tool. We expect your GCP projects to pre-exist before running the deployment.

By default, if only one project in your configuration contains a `vpc` block, it is automatically designated as the `Shared VPC Host` Project.

You can override this default behavior or resolve ambiguity if multiple projects contain a VPC. In such cases, you must explicitly set the `"hostProject": true` flag on the project you intend to be the host. Failure to do so when multiple VPCs are defined will result in a configuration error.

The `createNat` and `createInterconnect` flags should be set within the host project's VPC block.

**Example (Explicit Host Declaration):**

```json
"projects": [
    {
      "projectId": "network-host-project",
      "hostProject": true,
      "vpc": [
        {
          "type": "vpc",
          "name": "central-shared-vpc",
          "createNat": true,
          "createInterconnect": false,
          "subnets": [
            // Subnet definitions go here
          ]
        }
      ]
    },
    {
      "projectId": "app-services-project",
       // This is a service project
    }
]
```

##### Defining Subnets

Subnets are defined within a VPC's `"subnets"` list.

  * For **VPC-native GKE**, you must define `secondaryIpRanges`.

**Example:**

```json
"subnets": [
  {
    "name": "app-services-subnet",
    "ipCidrRange": "10.100.0.0/20",
    "region": "us-central1"
  },
  {
    "name": "psc-subnet",
    "ipCidrRange": "10.100.32.0/24",
    "region": "us-central1"
  },
  {
    "name": "gke-subnet",
    "ipCidrRange": "10.50.0.0/20",
    "region": "us-central1",
    "secondaryIpRanges": [
      { "rangeName": "pods-range", "ipCidrRange": "192.168.0.0/16" },
      { "rangeName": "services-range", "ipCidrRange": "192.168.128.0/20" }
    ]
  }
]
```

-----

#### Defining Producers (Services)

Producers are the services you want to make available (e.g., databases). They are defined inside a project's `producers` list.

##### 1. Producers with Private Service Connect (PSC)

This is the recommended way to expose services like Cloud SQL or AlloyDB privately. The network and subnet for the PSC endpoint can be defined **directly on the producer**.

**Example:**

```json
"producers": [
  {
    "type": "cloudsql",
    "name": "my-database-instance",
    "region": "us-central1",
    "createRequiredFwRules": true,
    "allowedConsumersTags": ["db-accessor"],
    "network": "central-shared-vpc",
    "subnet": "psc-subnet"
  }
]
```

A global `pscSettings` block is also supported as a fallback if `network` and `subnet` are not found on the producer.

##### 2. Memorystore for Redis with Service Connecting Policy (SCP)

This connects Redis directly to your VPC. You must explicitly provide the `subnet` for the instance.

**Example:**

```json
"producers": [
  {
    "type": "memorystore_redis_cluster",
    "name": "my-main-redis-cluster",
    "region": "us-central1",
    "subnet": "main-app-subnet"
  }
]
```

-----

#### Defining Consumers (Applications)

Consumers are the applications that access the producers. They are defined in a project's `consumers` list. The `tags` are crucial for security, as they must match the `allowedConsumersTags` on a producer for firewall rules to be automatically created along with manual Private Service Connect connections.

**Example GCE VM:**

```json
"consumers": [
  {
    "type": "vm",
    "name": "my-app-vm",
    "count": 1,
    "zone": "us-central1-a",
    "networkInterfaces": [
      {
        "network": "projects/network-host-project/global/networks/central-shared-vpc",
        "subnetwork": "projects/network-host-project/regions/us-central1/subnetworks/app-services-subnet"
      }
    ],
    "tags": { "items": ["db-accessor"] }
  }
]
```

-----

### How to Run the Generator

1.  Navigate to the project's root directory in your terminal.
2.  Run the command:
    ```bash
    python3 config_generator.py --all
    ```
3.  The script will prompt you to select an architecture file from the list.
4.  It will then guide you through the interactive process of reviewing the generated files and confirming deployment.

Enjoy simplifying your cloud networking\!