# Managed Instance Groups

## Overview

This Terraform solution provides a comprehensive approach to deploying and managing Google Cloud Managed Instance Groups (MIGs) using Terraform modules. Managed Instance Groups allow for the automatic management of identical virtual machine (VM) instances, ensuring high availability, scalability, and self-healing capabilities.

The solution utilizes a modular design, with the `mig.tf` file defining a Terraform module that leverages the `compute-mig` module from the Google Cloud Foundation Fabric. This module encapsulates the logic for creating and configuring MIGs based on parameters specified in a local map.

## Pre-Requisites

### Prior Step Completion:

- **Completed Prior Stages:** Successful deployment of MIG resources depends on the completion of the following stages:

    - **01-organization:** Activation of required Google Cloud APIs for MIG.
    - **02-networking:** Setup of necessary network infrastructure, including VPCs and subnets, to support MIG connectivity.
    - **03-security/MIG:** Configuration of firewall rules to allow access to MIG instances on appropriate ports and IP ranges.

### Enabled APIs:

Ensure the following Google Cloud APIs are enabled in your project:

- Compute Engine API
- Google Compute Engine Instance Group Manager API

### Permissions:

The user or service account executing Terraform must have the following roles (or equivalent permissions):

- Compute Admin (for managing VMs)
- Instance Group Manager (for managing MIGs)
- Service Account User (if using service accounts)

## Execution Steps

1. **Configuration:**

    - Define your YAML configurations for each managed instance group.
    - You can place them in the `configuration/consumer/MIG` folder.

2. **Terraform Initialization:**

    - Open your terminal and navigate to the directory containing your Terraform configuration.
    - Run the following command to initialize Terraform:

    ```bash
    terraform init
    ```

3. **Review the Execution Plan:**

    - Use the following command to generate an execution plan. This will show you the changes Terraform will make to your Google Cloud infrastructure:

    ```bash
    terraform plan
    ```

   Carefully review the plan to ensure it aligns with your intended configuration.

4. **Apply the Configuration:**

    Once you're satisfied with the plan, execute the terraform apply command to provision your Managed Instance Groups:

    ```bash
    terraform apply
    ```

   Terraform will create the corresponding MIGs in your Google Cloud project based on your configurations.

5. **Monitor and Manage:**

   After the instances are created, you can monitor their status, performance, and logs through the Google Cloud Console or using the Google Cloud CLI. Use Terraform to manage updates and changes to your Managed Instance Groups as needed.

## Important Notes

- The solution assumes that all required network and subnetwork resources already exist in your project as per previous steps.
- Ensure that you provide correct service account credentials (if applicable) to allow Terraform to interact with your Google Cloud project.
- Refer to the `variables.tf` file for a complete list of available variables and their descriptions.
- The Terraform modules used in this solution (`cloud-foundation-fabric/modules/compute-mig` and `cloud-foundation-fabric/modules/compute-vm`) might have additional configuration options. Refer to their respective documentation for further customization.

## Examples

### Sample YAMLs

Here is a sample maximum yaml example : 

```yaml
name: maximum-mig
project_id: <project-id>
location: <region> E.g. : us-central1
vpc_name : <network-name>
subnetwork_name : <subnetwork-name>
zone : <zone> E.g. : us-central1-a
target_size: 2
auto_healing_policies:
  initial_delay_sec: 30
health_check_config:
  enable_logging: true
  tcp:
    port: 80
autoscaler_config:
  max_replicas: 5
  min_replicas: 2
  cooldown_period: 60
  scaling_signals:
    cpu_utilization:
      target: 0.75
      optimize_availability: true
description: This is a maximum configuration for a managed instance group created using the CNCS repository
distribution_policy:
  target_shape: even
  zones:
    - <zone-1> E.g. : us-central1-a
    - <zone-2> E.g. : us-central1-b
    - <zone-3> E.g. : us-central1-c
```

Here is a sample minimum yaml example : 

```yaml
name: minimal-mig
project_id: <project-id>
location: <region> E.g. : us-central1
zone : <zone> E.g. : us-central1-a
vpc_name : <network-name>
subnetwork_name : <subnetwork-name>
health_check_config:
  enable_logging: true
  tcp:
    port: 80
```
<!-- BEGIN_TF_DOCS -->

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_mig"></a> [mig](#module\_mig) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/compute-mig | n/a |
| <a name="module_mig-template"></a> [mig-template](#module\_mig-template) | github.com/GoogleCloudPlatform/cloud-foundation-fabric//modules/compute-vm | n/a |


## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_addresses"></a> [addresses](#input\_addresses) | List of static IP addresses to assign to the instances. | `string` | `null` | no |
| <a name="input_all_instances_config"></a> [all\_instances\_config](#input\_all\_instances\_config) | Metadata and labels set to all instances in the group. | <pre>object({<br>    labels   = optional(map(string))<br>    metadata = optional(map(string))<br>  })</pre> | `null` | no |
| <a name="input_auto_healing_policies"></a> [auto\_healing\_policies](#input\_auto\_healing\_policies) | Auto-healing policies for this group. | <pre>object({<br>    health_check      = optional(string)<br>    initial_delay_sec = number<br>  })</pre> | <pre>{<br>  "health_check": null,<br>  "initial_delay_sec": 30<br>}</pre> | no |
| <a name="input_autoscaler_config"></a> [autoscaler\_config](#input\_autoscaler\_config) | Optional autoscaler configuration. | <pre>object({<br>    max_replicas    = number<br>    min_replicas    = number<br>    cooldown_period = optional(number)<br>    mode            = optional(string) # OFF, ONLY_UP, ON<br>    scaling_control = optional(object({<br>      down = optional(object({<br>        max_replicas_fixed   = optional(number)<br>        max_replicas_percent = optional(number)<br>        time_window_sec      = optional(number)<br>      }))<br>      in = optional(object({<br>        max_replicas_fixed   = optional(number)<br>        max_replicas_percent = optional(number)<br>        time_window_sec      = optional(number)<br>      }))<br>    }), {})<br>    scaling_signals = optional(object({<br>      cpu_utilization = optional(object({<br>        target                = number<br>        optimize_availability = optional(bool)<br>      }))<br>      load_balancing_utilization = optional(object({<br>        target = number<br>      }))<br>      metrics = optional(list(object({<br>        name                       = string<br>        type                       = optional(string) # GAUGE, DELTA_PER_SECOND, DELTA_PER_MINUTE<br>        target_value               = optional(number)<br>        single_instance_assignment = optional(number)<br>        time_series_filter         = optional(string)<br>      })))<br>      schedules = optional(list(object({<br>        duration_sec          = number<br>        name                  = string<br>        min_required_replicas = number<br>        cron_schedule         = string<br>        description           = optional(bool)<br>        timezone              = optional(string)<br>        disabled              = optional(bool)<br>      })))<br>    }), {})<br>  })</pre> | <pre>{<br>  "cooldown_period": null,<br>  "max_replicas": 3,<br>  "min_replicas": 1,<br>  "scaling_signals": {<br>    "cpu_utilization": {<br>      "optimize_availability": false,<br>      "target": 0.65<br>    }<br>  }<br>}</pre> | no |
| <a name="input_config_folder_path"></a> [config\_folder\_path](#input\_config\_folder\_path) | Location of YAML files holding GCE configuration values. | `string` | `"../../../configuration/consumer/MIG/config"` | no |
| <a name="input_create_nat"></a> [create\_nat](#input\_create\_nat) | True or False to create NAT for template network interface. | `bool` | `true` | no |
| <a name="input_create_template"></a> [create\_template](#input\_create\_template) | True or False to create a template | `bool` | `true` | no |
| <a name="input_default_version_name"></a> [default\_version\_name](#input\_default\_version\_name) | Name used for the default version. | `string` | `"default"` | no |
| <a name="input_description"></a> [description](#input\_description) | Optional description used for all resources managed by this module. | `string` | `"Terraform managed."` | no |
| <a name="input_distribution_policy"></a> [distribution\_policy](#input\_distribution\_policy) | DIstribution policy for regional MIG. | <pre>object({<br>    target_shape = optional(string)<br>    zones        = optional(list(string))<br>  })</pre> | `null` | no |
| <a name="input_health_check_config"></a> [health\_check\_config](#input\_health\_check\_config) | Optional auto-created health check configuration, use the output self-link to set it in the auto healing policy. Refer to examples for usage. | <pre>object({<br>    check_interval_sec  = optional(number)<br>    description         = optional(string, "Terraform managed.")<br>    enable_logging      = optional(bool, false)<br>    healthy_threshold   = optional(number)<br>    timeout_sec         = optional(number)<br>    unhealthy_threshold = optional(number)<br>    grpc = optional(object({<br>      port               = optional(number)<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      service_name       = optional(string)<br>    }))<br>    http = optional(object({<br>      host               = optional(string)<br>      port               = optional(number)<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      proxy_header       = optional(string)<br>      request_path       = optional(string)<br>      response           = optional(string)<br>    }))<br>    http2 = optional(object({<br>      host               = optional(string)<br>      port               = optional(number)<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      proxy_header       = optional(string)<br>      request_path       = optional(string)<br>      response           = optional(string)<br>    }))<br>    https = optional(object({<br>      host               = optional(string)<br>      port               = optional(number)<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      proxy_header       = optional(string)<br>      request_path       = optional(string)<br>      response           = optional(string)<br>    }))<br>    tcp = optional(object({<br>      port               = optional(number) # This will be overridden in the default<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      proxy_header       = optional(string)<br>      request            = optional(string)<br>      response           = optional(string)<br>    }))<br>    ssl = optional(object({<br>      port               = optional(number)<br>      port_name          = optional(string)<br>      port_specification = optional(string) # USE_FIXED_PORT USE_NAMED_PORT USE_SERVING_PORT<br>      proxy_header       = optional(string)<br>      request            = optional(string)<br>      response           = optional(string)<br>    }))<br>  })</pre> | <pre>{<br>  "check_interval_sec": null,<br>  "description": "Terraform managed.",<br>  "enable_logging": true,<br>  "grpc": null,<br>  "healthy_threshold": null,<br>  "http": null,<br>  "http2": null,<br>  "https": null,<br>  "ssl": null,<br>  "tcp": {<br>    "port": 80,<br>    "port_name": null,<br>    "port_specification": null,<br>    "proxy_header": null,<br>    "request": null,<br>    "response": null<br>  },<br>  "timeout_sec": 90,<br>  "unhealthy_threshold": null<br>}</pre> | no |
| <a name="input_metadata"></a> [metadata](#input\_metadata) | Metadata of the instances being created as a part of the MIG | `string` | `"#!/bin/bash\n\n# Update and install Apache2 & PHP\nsudo apt-get update\nsudo apt-get install -y apache2\nsudo apt-get install -y php libapache2-mod-php\nsudo a2enmod php\n\n# Restart Apache2 to apply changes\nsudo systemctl restart apache2\n\n# Create the PHP file\ncat << EOL > /var/www/html/hostname.php\n<!DOCTYPE html>\n<html>\n<body>\n\n<p>Hostname: <?php echo gethostname(); ?></p>\n\n</body>\n</html>\nEOL\n\n# Start Apache2 service\nsudo service apache2 start\n\n# Enable Apache2 to start on boot\nsudo update-rc.d apache2 enable\n"` | no |
| <a name="input_mig_image"></a> [mig\_image](#input\_mig\_image) | Image for the MIG instance. | `string` | `"projects/debian-cloud/global/images/family/debian-11"` | no |
| <a name="input_mig_template_name"></a> [mig\_template\_name](#input\_mig\_template\_name) | Name for the MIG instance template. | `string` | `"mig-template"` | no |
| <a name="input_named_ports"></a> [named\_ports](#input\_named\_ports) | Named ports. | `map(number)` | `null` | no |
| <a name="input_region"></a> [region](#input\_region) | Region for the resources. | `string` | `"us-central1"` | no |
| <a name="input_stateful_config"></a> [stateful\_config](#input\_stateful\_config) | Stateful configuration for individual instances. | <pre>map(object({<br>    minimal_action          = optional(string)<br>    most_disruptive_action  = optional(string)<br>    remove_state_on_destroy = optional(bool)<br>    preserved_state = optional(object({<br>      disks = optional(map(object({<br>        source                      = string<br>        delete_on_instance_deletion = optional(bool)<br>        read_only                   = optional(bool)<br>      })))<br>      metadata = optional(map(string))<br>    }))<br>  }))</pre> | `{}` | no |
| <a name="input_stateful_disks"></a> [stateful\_disks](#input\_stateful\_disks) | Stateful disk configuration applied at the MIG level to all instances, in device name => on permanent instance delete rule as boolean. | `map(bool)` | `{}` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Tags for the instance. | `list(string)` | <pre>[<br>  "allow-health-checks"<br>]</pre> | no |
| <a name="input_target_pools"></a> [target\_pools](#input\_target\_pools) | Optional list of URLs for target pools to which new instances in the group are added. | `list(string)` | `[]` | no |
| <a name="input_target_size"></a> [target\_size](#input\_target\_size) | Group target size, leave null when using an autoscaler. | `number` | `1` | no |
| <a name="input_update_policy"></a> [update\_policy](#input\_update\_policy) | Update policy. Minimal action and type are required. | <pre>object({<br>    minimal_action = string<br>    type           = string<br>    max_surge = optional(object({<br>      fixed   = optional(number)<br>      percent = optional(number)<br>    }))<br>    max_unavailable = optional(object({<br>      fixed   = optional(number)<br>      percent = optional(number)<br>    }))<br>    min_ready_sec                = optional(number)<br>    most_disruptive_action       = optional(string)<br>    regional_redistribution_type = optional(string)<br>    replacement_method           = optional(string)<br>  })</pre> | `null` | no |
| <a name="input_versions"></a> [versions](#input\_versions) | Additional application versions, target\_size is optional. | <pre>map(object({<br>    instance_template = string<br>    target_size = optional(object({<br>      fixed   = optional(number)<br>      percent = optional(number)<br>    }))<br>  }))</pre> | `{}` | no |
| <a name="input_wait_for_instances"></a> [wait\_for\_instances](#input\_wait\_for\_instances) | Wait for all instances to be created/updated before returning. | <pre>object({<br>    enabled = bool<br>    status  = optional(string)<br>  })</pre> | `null` | no |
| <a name="input_zone"></a> [zone](#input\_zone) | Zone for the resources. | `string` | `"us-central1-b"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_autoscaler"></a> [autoscaler](#output\_autoscaler) | Auto-created autoscaler resource. |
| <a name="output_autoscaler_config"></a> [autoscaler\_config](#output\_autoscaler\_config) | Configuration details of the autoscaler. |
| <a name="output_group_manager"></a> [group\_manager](#output\_group\_manager) | Instance group resource. |
| <a name="output_health_check"></a> [health\_check](#output\_health\_check) | Auto-created health-check resource. |
| <a name="output_id"></a> [id](#output\_id) | Fully qualified group manager id. |
| <a name="output_instance_template"></a> [instance\_template](#output\_instance\_template) | The self-link of the instance template used for the MIG. |

<!-- END_TF_DOCS -->