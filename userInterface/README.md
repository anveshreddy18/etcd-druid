# userInterface

This module provides user-facing interfaces (CLI and TUI) for interacting with the etcd-druid operator in Kubernetes.

## Structure
- `cli/`   - Cobra-based CLI commands
- `tui/`   - Bubbletea-based TUI
- `core/`  - Shared logic for operator interaction
- `pkg/`   - Utilities and shared types

## Getting Started
- Run `go mod tidy` in this directory after cloning.
- To build the CLI: `go run ./cli`
- To run the TUI: `go run ./tui`

## Requirements
- Go 1.24+
- Access to a Kubernetes cluster (or kubeconfig)
