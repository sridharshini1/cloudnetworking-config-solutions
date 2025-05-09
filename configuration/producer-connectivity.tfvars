psc_endpoints = [
  // Configuration for a PSC endpoint with a CloudSQL instance
  // This configuration includes a producer_instance_name field that specifies the CloudSQL instance name.
  // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-1"
    network_name                 = "network-1"
    ip_address_literal           = "10.128.0.26"
    region                       = "" # Example: us-central1
    producer_cloudsql = {
      instance_name = "psc-instance-name"
    }
  },
  // Configuration for a PSC endpoint with an AlloyDB instance
  // This configuration includes producer_alloydb_instance_name and cluster_id fields that specify the AlloyDB instance and cluster.
  // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-2"
    network_name                 = "network-2"
    ip_address_literal           = "10.128.0.27"
    region                       = "" # Example: us-central2
    producer_alloydb = {
      instance_name = "your-alloydb-instance-name"
      cluster_id    = "your-cluster-id"
    }
  },
  // Configuration for a PSC endpoint with a target
  // This configuration includes a target field that specifies the service attachment URL.
  // Example values are provided for subnetwork_name, network_name, ip_address_literal, and region.
  {
    endpoint_project_id          = "your-endpoint-project-id"
    producer_instance_project_id = "your-producer-instance-project-id"
    subnetwork_name              = "subnetwork-3"
    network_name                 = "network-3"
    ip_address_literal           = "10.0.0.10"
    region                       = "" # Example: us-central1
    target                       = "projects/your-project-id/regions/us-central1/serviceAttachments/your-service-attachment-id"
  }
]