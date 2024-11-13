# CAPTEP-0012: Region Configuration for Control Plane

## Summary
Add region configuration capability to CAPTControlPlane to support AWS region-specific deployments.

## Motivation
The EKS control plane template currently retrieves the AWS region information from AdditionalTags, which is not an appropriate way to handle such a critical configuration parameter. Region information is essential for proper EKS cluster configuration, especially for kubeconfig generation and AWS resource creation.

### Problems with Current Approach
1. Using AdditionalTags for region configuration:
   - Tags are meant for metadata and resource organization, not for critical configuration
   - No validation of region values
   - Implicit dependency that's not clear from the API
2. Kubeconfig generation requires region information:
   - Currently relies on tag value which could be missing or incorrect
   - No type safety or validation
3. AWS resource creation:
   - Region is a fundamental parameter for AWS resource creation
   - Should be explicitly defined in the API

## Goals
- Add region configuration to the CAPTControlPlane API as a first-class field
- Ensure backward compatibility with existing deployments
- Provide clear documentation for region configuration
- Support AWS region validation
- Improve type safety and validation for region configuration

## Non-Goals
- Changing the underlying AWS provider configuration
- Supporting multi-region deployments within a single control plane
- Modifying existing tag-based functionality for other purposes

## Proposal

### API Changes

Add region field to the ControlPlaneConfig struct in the CAPTControlPlane API:

```go
type ControlPlaneConfig struct {
    // Region specifies the AWS region where the control plane will be created
    // +kubebuilder:validation:Required
    Region string `json:"region"`

    // Existing fields...
    EndpointAccess *EndpointAccess `json:"endpointAccess,omitempty"`
    Addons []Addon `json:"addons,omitempty"`
    Timeouts *TimeoutConfig `json:"timeouts,omitempty"`
}
```

### Implementation Details

1. Update CAPTControlPlane API types:
   - Add Region field to ControlPlaneConfig
   - Add validation for AWS region format
   - Update CRD generation

2. Update WorkspaceTemplate handling:
   - Modify eks-controlplane-template to use spec.controlPlaneConfig.region
   - Remove region from AdditionalTags usage
   - Update template variable mapping

3. Add region validation in the webhook:
   - Validate AWS region format
   - Ensure region is provided when required
   - Add validation tests

4. Update documentation and examples:
   - Update all example manifests
   - Add migration guide
   - Update user documentation

### Migration Strategy

To maintain backward compatibility:

1. Short term:
   - Support both new Region field and legacy tag-based approach
   - Log warning when using tag-based region
   - Prioritize Region field over tag when both are present

2. Long term:
   - Deprecate tag-based region configuration
   - Add validation webhook to require Region field
   - Remove tag-based region support in future release

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

### Validation

The controller will validate:
1. Region format matches AWS region pattern
2. Region is a valid AWS region
3. Region is consistent with AWS provider configuration

## Alternatives Considered

1. Keep using AdditionalTags:
   - Pros: No API changes needed
   - Cons: Continues problematic pattern, lacks validation
   - Decision: Rejected due to poor design practice

2. Add Region at root level of CAPTControlPlaneSpec:
   - Pros: More visible, clearer importance
   - Cons: Inconsistent with other AWS-specific configs
   - Decision: Rejected for consistency with existing pattern

3. Use annotations for region:
   - Pros: No API changes needed
   - Cons: Same issues as tags, harder to validate
   - Decision: Rejected for same reasons as current approach

## Implementation History

- [x] 2024-11-12: Initial proposal
- [ ] Implementation of API changes
- [ ] Implementation of validation
- [ ] Documentation updates
- [ ] Migration guide
- [ ] Release notes

## Technical Leads
- @reoring

## References
- [AWS Regions and Endpoints](https://docs.aws.amazon.com/general/latest/gr/rande.html)
- [EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/clusters.html)
