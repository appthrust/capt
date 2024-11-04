# Fargate Profile Implementation Details

## Overview

This document provides detailed technical information about the implementation of Fargate Profile management in CAPT. It covers the specific implementation details, data structures, and workflows used to manage EKS Fargate profiles.

## API Design

### CAPTControlPlane Extensions

```go
type CAPTControlPlaneSpec struct {
    // ... existing fields ...

    // AdditionalFargateProfiles defines additional Fargate profiles to be created
    // +optional
    AdditionalFargateProfiles []AdditionalFargateProfile `json:"additionalFargateProfiles,omitempty"`
}

type AdditionalFargateProfile struct {
    // Name is the name of the Fargate profile
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // Selectors is a list of label selectors to use for pods
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinItems=1
    Selectors []FargateSelector `json:"selectors"`

    // WorkspaceTemplateRef is a reference to the WorkspaceTemplate
    // +kubebuilder:validation:Required
    WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`
}
```

### Status Management

```go
type FargateProfileStatus struct {
    // Name is the name of the Fargate profile
    Name string `json:"name"`

    // Ready indicates if the Fargate profile is ready
    Ready bool `json:"ready"`

    // WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply resource
    WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName"`

    // FailureReason indicates that there is a problem with the Fargate profile
    // +optional
    FailureReason *string `json:"failureReason,omitempty"`

    // FailureMessage provides more detail about the failure
    // +optional
    FailureMessage *string `json:"failureMessage,omitempty"`
}
```

## Controller Implementation

### Reconciliation Flow

1. EKS Cluster Check:
```go
// Check if EKS cluster is ready by checking for eks-connection secret
eksSecret := &corev1.Secret{}
if err := r.Get(ctx, types.NamespacedName{
    Name:      eksConnectionSecret,
    Namespace: controlPlane.Namespace,
}, eksSecret); err != nil {
    return ctrl.Result{RequeueAfter: requeueInterval}, nil
}
```

2. Profile Management:
```go
func (r *CAPTControlPlaneReconciler) reconcileFargateProfiles(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
    // Initialize status if needed
    if controlPlane.Status.FargateProfileStatuses == nil {
        controlPlane.Status.FargateProfileStatuses = []controlplanev1beta1.FargateProfileStatus{}
    }

    // Process each profile
    for _, profile := range controlPlane.Spec.AdditionalFargateProfiles {
        // ... profile reconciliation logic ...
    }
}
```

### WorkspaceTemplateApply Creation

```go
func (r *CAPTControlPlaneReconciler) reconcileFargateProfileWorkspaceApply(
    ctx context.Context,
    controlPlane *controlplanev1beta1.CAPTControlPlane,
    profile controlplanev1beta1.AdditionalFargateProfile,
    status *controlplanev1beta1.FargateProfileStatus,
) error {
    // Convert selectors to JSON
    selectorsJSON, err := json.Marshal(profile.Selectors)
    if err != nil {
        return fmt.Errorf("failed to marshal selectors: %v", err)
    }

    // Create WorkspaceTemplateApply
    workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{
        ObjectMeta: metav1.ObjectMeta{
            Name:      status.WorkspaceTemplateApplyName,
            Namespace: controlPlane.Namespace,
        },
        Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
            // ... template and variable configuration ...
        },
    }

    // Create or update
    return r.Create(ctx, workspaceApply)
}
```

## Terraform Module Integration

### Variable Passing

Variables passed to the Terraform module:

```hcl
variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster"
}

variable "profile_name" {
  type        = string
  description = "Name of the Fargate profile"
}

variable "selectors" {
  type = list(object({
    namespace = string
    labels    = map(string)
  }))
  description = "Pod selectors for the Fargate profile"
}
```

### Resource Creation

```hcl
resource "aws_eks_fargate_profile" "this" {
  cluster_name           = var.cluster_name
  fargate_profile_name   = var.profile_name
  pod_execution_role_arn = data.aws_eks_cluster.cluster.role_arn
  subnet_ids            = var.private_subnets

  dynamic "selector" {
    for_each = var.selectors
    content {
      namespace = selector.value.namespace
      labels   = lookup(selector.value, "labels", {})
    }
  }

  tags = var.tags
}
```

## Error Handling and Recovery

### Status Updates

```go
// Update profile status based on WorkspaceTemplateApply status
for _, condition := range workspaceApply.Status.Conditions {
    if condition.Type == xpv1.TypeReady {
        if condition.Status == corev1.ConditionTrue {
            profileStatus.Ready = true
            profileStatus.FailureReason = nil
            profileStatus.FailureMessage = nil
        } else {
            profileStatus.Ready = false
            reason := controlplanev1beta1.ReasonFargateProfileFailed
            message := condition.Message
            profileStatus.FailureReason = &reason
            profileStatus.FailureMessage = &message
            return fmt.Errorf("Fargate profile %s failed: %s", profile.Name, message)
        }
        break
    }
}
```

### Dependency Management

```go
// Wait for required secrets
WaitForSecrets: []xpv1.SecretReference{
    {
        Name:      eksConnectionSecret,
        Namespace: controlPlane.Namespace,
    },
    {
        Name:      "vpc-connection",
        Namespace: controlPlane.Namespace,
    },
}
```

## Testing Considerations

1. Unit Tests:
   - Profile creation validation
   - Status updates
   - Error handling

2. Integration Tests:
   - Secret dependency management
   - WorkspaceTemplateApply creation
   - Profile lifecycle management

3. End-to-End Tests:
   - Complete profile creation workflow
   - Error recovery scenarios
   - Multi-profile management
