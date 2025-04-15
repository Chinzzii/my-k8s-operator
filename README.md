# Smooth K8s Operator

Smooth K8s Operator is a custom Kubernetes operator written in Go that manages a custom resource definition (CRD) called **StaticPage**. This operator automatically creates and updates a Deployment and a ConfigMap based on the configuration specified in your StaticPage custom resource. It’s designed to provide a simple example for managing content pages with dynamic configuration changes.

## Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Project Structure](#project-structure)
- [Installation and Deployment](#installation-and-deployment)
- [Additional Considerations](#additional-considerations)

## Project Overview

This operator allows you to define a **StaticPage** CR that contains:

- **Contents:** The HTML content to be served.
- **Image:** The container image to use (e.g., _nginx:latest_).
- **Replicas:** The number of pod replicas to run.

Once you create a StaticPage resource, the operator creates or updates the following Kubernetes resources in the same namespace:

- A **Deployment** for running the specified container.
- A **ConfigMap** that holds the HTML content.

The operator monitors changes to the custom resource and performs reconciliation to ensure that the live state of the cluster reflects the desired specification.

## Features

- **Custom Resource Definition (CRD):** Define a `StaticPage` custom resource.
- **Reconciliation Loop:** Automatically creates, updates, or deletes related resources (Deployment and ConfigMap) based on changes to the StaticPage resource.
- **Logging:** Uses structured logging (via zap) for better troubleshooting.
- **RBAC Integration:** Provides YAML manifest files for required RBAC permissions.
- **Docker Integration:** Multi-stage Dockerfile builds a minimal runtime image for the operator.

## Project Structure

```bash
├── api
│ └── v1
│    ├── deepcopy.go    # Custom deep copy implementations
│    ├── register.go    # API registration to the runtime scheme
│    └── staticpage.go  # CRD type definitions for StaticPage
├── cmd
│    └── controller
│       └── main.go     # The operator's main entry point
├── yaml
│    ├── crd.yaml               # CRD definition for StaticPage
│    ├── deploy-controller.yaml # Deployment, ServiceAccount, and RBAC for the operator
│    └── example.yaml           # An example StaticPage custom resource
└── Dockerfile      # Dockerfile for building the operator image
```

## Installation and Deployment

### 1. Deploy CRD and Operator Resources

Apply the provided YAML manifests in the following order:

1. **CRD** – Deploy the CustomResourceDefinition for StaticPage:
   ```bash
   kubectl apply -f yaml/crd.yaml
   ```
2. **Operator Deployment and RBAC** – Deploy the operator and associated RBAC settings:
   ```bash
   kubectl apply -f yaml/deploy-controller.yaml
   ```
3. **Example Custom Resource** – Create an example StaticPage custom resource:
   ```bash
   kubectl apply -f yaml/example.yaml
   ```

### 2. Building and Running Locally

You can run the operator locally from VS Code. Ensure that your kubeconfig is properly configured. To start the operator from your local machine:

```bash
go run cmd/controller/main.go
```

Check the logs to verify that the operator is correctly reconciling the `StaticPage` resource.

## Additional Considerations

- **RBAC and API Groups:**
  Verify that your RBAC definitions in `deploy-controller.yaml` use the correct API group matching your CRD (`kubernetes.chinzzii.com`).

- **Future Enhancements:**
  You might consider adding support for additional fields in the CRD, more advanced reconciliation logic, custom events, or metrics.
