# CAPTEP-0012: Region Configuration for Control Plane

## Summary
Add region configuration capability to CAPTControlPlane to support AWS region-specific deployments.

## Motivation
The EKS control plane template requires AWS region information for proper configuration, especially for kubeconfig generation. Currently, there is no standardized way to specify the region in the CAPTControlPlane API, which could lead to inconsistencies and potential issues in multi-region deployments.

## Goals
- Add region configuration to the CAPTControlPlane API
- Ensure backward compatibility with existing deployments
- Provide clear documentation for region configuration
- Support AWS region validation

## Proposal

### API Changes

Add region field to the ControlPlaneConfig struct in the CAPTControlPlane API:

```go
type ControlPlaneConfig struct {
    // Region specifies the AWS region where the control plane will be created
    // +optional
    Region string `json:"region,omitempty"`

    // Existing fields...
    EndpointAccess *EndpointAccess `json:"endpointAccess,omitempty"`
    Addons []Addon `json:"addons,omitempty"`
    Timeouts *TimeoutConfig `json:"timeouts,omitempty"`
}
```

### Implementation Details

1. Update the CAPTControlPlane API types
2. Add region validation in the webhook
3. Pass region information to the WorkspaceTemplate
4. Update documentation and examples

### Migration Strategy

The region field will be optional to maintain backward compatibility. If not specified:
1. Default to the region specified in AWS provider configuration
2. Log a warning about missing region configuration

### Example Usage

```yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: CAPTControlPlane
metadata:
  name: example-control-plane
spec:
  version: "1.31"
  controlPlaneConfig:
    region: "ap-northeast-1"
    endpointAccess:
      public: true
      private: true
```

## Alternatives Considered

1. Using annotations for region configuration
   - Pros: No API changes needed
   - Cons: Less explicit, harder to validate

2. Global region configuration
   - Pros: Simpler configuration
   - Cons: Less flexible for multi-region deployments

## Implementation History

- [ ] 2024-11-12: Initial proposal
- [ ] Implementation of API changes
- [ ] Implementation of validation
- [ ] Documentation updates
