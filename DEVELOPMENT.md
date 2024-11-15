# Development Guide

This guide explains how to set up your development environment and contribute to CAPT (Cluster API Provider for Tofu/Terraform).

## Prerequisites

- Go 1.22 or later
- Docker
- kubectl
- A Kubernetes cluster (for testing)
- AWS credentials (for testing with AWS)

## Setting Up Your Development Environment

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/capt.git
   cd capt
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Install required tools:
   ```bash
   # These will be installed in ./bin/
   make kustomize    # Install kustomize
   make controller-gen # Install controller-gen
   make envtest     # Install setup-envtest
   make golangci-lint # Install golangci-lint
   ```

## Development Workflow

### Building

1. Build the manager binary:
   ```bash
   make build
   ```

2. Build the Docker image:
   ```bash
   make docker-build
   ```

3. Build multi-architecture Docker image:
   ```bash
   make docker-buildx
   ```

### Testing

1. Run unit tests:
   ```bash
   make test
   ```

2. Run end-to-end tests:
   ```bash
   make test-e2e
   ```

3. Run linter:
   ```bash
   make lint
   ```

### Generating Code and Manifests

1. Generate CRDs and RBAC manifests:
   ```bash
   make manifests
   ```

2. Generate code (DeepCopy methods, etc.):
   ```bash
   make generate
   ```

### Local Development

1. Install CRDs in your cluster:
   ```bash
   make install
   ```

2. Run the controller locally:
   ```bash
   make run
   ```

3. Deploy the controller to a cluster:
   ```bash
   make deploy
   ```

## Project Structure

```
.
├── api/                    # API definitions
│   └── v1beta1/           # API version
├── cmd/
│   └── main.go            # Main entry point
├── config/                # Kubernetes manifests
│   ├── crd/              # Custom Resource Definitions
│   ├── default/          # Default configurations
│   ├── manager/          # Controller manager configurations
│   ├── rbac/            # RBAC configurations
│   └── samples/          # Example CR manifests
├── docs/                  # Documentation
├── hack/                  # Development scripts
├── internal/
│   └── controller/       # Controller implementations
└── test/                 # Test files
    └── e2e/             # End-to-end tests
```

## Making Changes

### Adding a New API Field

1. Add the field to the appropriate type in `api/v1beta1/`
2. Run:
   ```bash
   make generate  # Generate DeepCopy methods
   make manifests # Update CRDs
   ```

### Adding a New Controller

1. Create a new controller in `internal/controller/`
2. Add tests in `internal/controller/`
3. Register the controller in `cmd/main.go`
4. Update RBAC if needed:
   ```bash
   make manifests
   ```

## Building and Testing

### Local Build

```bash
make build
```

This creates a binary in `bin/manager`

### Docker Build

```bash
make docker-build
```

Builds a Docker image with the controller

### Multi-arch Build

```bash
make docker-buildx
```

Builds and pushes multi-architecture images (linux/amd64, linux/arm64)

## Testing

### Prerequisites

The test suite requires:
- Go 1.22+
- Access to a Kubernetes cluster (for e2e tests)
- AWS credentials (for AWS-related tests)

### Running Tests

1. Unit tests:
   ```bash
   make test
   ```

2. E2E tests:
   ```bash
   make test-e2e
   ```

### Test Coverage

Generate a coverage report:
```bash
make test
```

Coverage report will be in `cover.out`

## Debugging

### Local Debugging

1. Run the controller locally:
   ```bash
   make run
   ```

2. Watch the logs:
   ```bash
   kubectl logs -n capt-system capt-controller-manager-xxx
   ```

### Remote Debugging

For remote debugging, you can use delve:

1. Build with debug info:
   ```bash
   go build -gcflags="all=-N -l" -o bin/manager cmd/main.go
   ```

2. Run with delve:
   ```bash
   dlv --listen=:2345 --headless=true --api-version=2 exec ./bin/manager
   ```

## Release Process

1. Update version:
   ```bash
   make update-version VERSION=x.y.z
   ```

2. Update CHANGELOG.md:
   ```bash
   make update-changelog
   ```

3. Build and push release:
   ```bash
   make release VERSION=x.y.z
   ```

This will:
- Update version numbers
- Update CHANGELOG.md
- Build and push multi-arch Docker images
- Generate installer YAML
- Create GitHub release

## Additional Resources

- [Cluster API Book](https://cluster-api.sigs.k8s.io/)
- [Kubernetes Controller Runtime](https://github.com/kubernetes-sigs/controller-runtime)
- [Crossplane Documentation](https://crossplane.io/docs/)
