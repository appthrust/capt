# Control Plane Creation Issue

## Current Behavior

1. Control Plane Configuration:
   ```yaml
   workspaceTemplateRef:
     name: eks-controlplane-template
     namespace: default
   ```

2. Current Status:
   - Control plane shows `Phase: Creating`
   - Controller is waiting for VPC readiness
   - WorkspaceTemplateApply for control plane exists and is ready:
     ```
     demo-eks-controlplane-apply-workspace   True     True    23h
     ```

## Issue Analysis

1. Incorrect Dependency Check:
   - Controller is focusing on VPC readiness
   - Should be checking EKS control plane workspace status instead
   - VPC dependency should be handled by WorkspaceTemplateApply's `WaitForWorkspaces`

2. Implementation Problems:
   - Controller is mixing infrastructure (VPC) concerns with control plane concerns
   - Not properly utilizing WorkspaceTemplateApply's built-in dependency management
   - Status updates are not reflecting the actual control plane workspace state

## Required Changes

1. CAPTControlPlane Controller:
   - Remove direct VPC status checking
   - Focus on control plane workspace status
   - Utilize WorkspaceTemplateApply's status for control plane readiness

2. Status Management:
   - Update status based on control plane workspace state
   - Properly propagate workspace conditions to CAPTControlPlane status
   - Add appropriate events for status changes

3. Dependency Management:
   - Rely on WorkspaceTemplateApply's `WaitForWorkspaces` for VPC dependency
   - Remove VPC-specific timeout handling
   - Add proper error handling for workspace creation failures

## Next Steps

1. Implementation Updates:
   - Modify controller to focus on control plane workspace
   - Update status handling logic
   - Remove VPC-specific code

2. Testing:
   - Verify control plane creation workflow
   - Test dependency handling through WorkspaceTemplateApply
   - Validate status updates

## Related Components

- CAPTControlPlane Controller
- EKS Control Plane WorkspaceTemplate
- WorkspaceTemplateApply Controller

## References

- Current implementation in `internal/controller/controlplane/captcontrolplane_controller.go`
- Control plane spec in `api/controlplane/v1beta1/captcontrolplane_types.go`
- WorkspaceTemplateApply in `api/v1beta1/workspacetemplateapply_types.go`
