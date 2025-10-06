# druidctl

This module provides CLI tool for interacting with the etcd-druid managed Etcd clusters in a Kubernetes cluster.

## Structure
- `cli/`   - Cobra-based CLI commands
- `client/` - For generating typed, generic and discovery clients
- `internal/`  - Shared logic for operator interaction
- `pkg/`   - Utilities and shared types

## Getting Started
- Run `make plugin-install` in this directory after cloning to make it into a kubectl plugin.
- Target any K8s cluster and make sure you're able to access it using `kubectl`.
- Then use it like `kubectl druid ...`.
- The list of implemented commands are: 
  - `reconcile`
  - `suspend-reconcile`
  - `resume-reconcile`
  - `remove-component-protection`
  - `add-component-protection`
  - `list-resources`

## Requirements
- Go 1.24+
- Access to a Kubernetes cluster (or kubeconfig)
