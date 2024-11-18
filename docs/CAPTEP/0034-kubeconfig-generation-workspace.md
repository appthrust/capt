# CAPTEP-0034: Kubeconfig Generation Workspace Template

## Summary

This proposal outlines the separation of kubeconfig generation functionality into a dedicated WorkspaceTemplate. This change aims to improve the modularity and maintainability of the EKS cluster management process.

## Motivation

Currently, kubeconfig generation is embedded within the EKS control plane WorkspaceTemplate. This coupling makes the template more complex and harder to maintain. By separating this functionality into its own WorkspaceTemplate, we can:

1. Improve modularity and separation of concerns
2. Make the kubeconfig generation process more maintainable
3. Allow for independent updates to the authentication mechanism
4. Reduce complexity in the main EKS control plane template

### Goals

- Create a dedicated WorkspaceTemplate for kubeconfig generation
- Ensure seamless integration with the existing EKS control plane workflow
- Support token-based authentication for EKS clusters
- Maintain security best practices for credential handling

### Non-Goals

- Modifying the existing EKS control plane template beyond removing kubeconfig generation
- Supporting authentication methods other than token-based authentication
- Implementing custom token refresh mechanisms

## Proposal

### User Stories

#### Story 1: Cluster Administrator

As a cluster administrator, I want the kubeconfig generation to be handled separately from the EKS control plane creation, so that I can manage and troubleshoot authentication issues more effectively.

#### Story 2: DevOps Engineer

As a DevOps engineer, I want to be able to regenerate kubeconfig without modifying the EKS control plane configuration, so that I can handle authentication updates independently.

### Implementation Details

1. Create a new WorkspaceTemplate for kubeconfig generation:

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: WorkspaceTemplate
metadata:
  name: eks-kubeconfig-template
spec:
  template:
    metadata:
      description: "Template for generating EKS kubeconfig"
    spec:
      forProvider:
        source: Inline
        module: |
          variable "cluster_name" {
            type = string
            description = "Name of the EKS cluster"
          }

          variable "region" {
            type = string
            description = "AWS region of the EKS cluster"
          }

          variable "cluster_endpoint" {
            type = string
            description = "Endpoint URL of the EKS cluster"
          }

          variable "cluster_certificate_authority_data" {
            type = string
            description = "Certificate authority data for the EKS cluster"
          }

          data "external" "eks_token" {
            program = ["sh", "-c", "aws eks get-token --output json --cluster-name ${var.cluster_name} --region ${var.region} | jq -r '{token: .status.token}'"]
          }

          output "kubeconfig" {
            sensitive = true
            value = yamlencode({
              apiVersion = "v1"
              clusters = [{
                cluster = {
                  certificate-authority-data = var.cluster_certificate_authority_data
                  server = var.cluster_endpoint
                }
                name = "${var.cluster_name}.${var.region}.eksctl.io"
              }]
              contexts = [{
                context = {
                  cluster = "${var.cluster_name}.${var.region}.eksctl.io"
                  user = "rancher-installer@${var.cluster_name}.${var.region}.eksctl.io"
                }
                name = "rancher-installer@${var.cluster_name}.${var.region}.eksctl.io"
              }]
              current-context = "rancher-installer@${var.cluster_name}.${var.region}.eksctl.io"
              kind = "Config"
              preferences = {}
              users = [{
                name = "rancher-installer@${var.cluster_name}.${var.region}.eksctl.io"
                user = {
                  token = data.external.eks_token.result.token
                }
              }]
            })
          }
```

2. Modify CAPTControlPlane controller to create WorkspaceTemplateApply for kubeconfig:

```go
// After EKS cluster is ready
if cluster.Status.Ready {
    // Create WorkspaceTemplateApply for kubeconfig
    kubeconfigApply := &infrastructurev1beta1.WorkspaceTemplateApply{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-kubeconfig", cluster.Name),
            Namespace: cluster.Namespace,
        },
        Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
            TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
                Name: "eks-kubeconfig-template",
            },
            Variables: map[string]string{
                "cluster_name": cluster.Name,
                "region":      cluster.Spec.Region,
                "cluster_endpoint": eksOutput.ClusterEndpoint,
                "cluster_certificate_authority_data": eksOutput.ClusterCertificateAuthorityData,
            },
            WaitForWorkspaces: []infrastructurev1beta1.WorkspaceReference{
                {
                    Name: fmt.Sprintf("%s-controlplane", cluster.Name),
                },
            },
        },
    }
    // Apply the kubeconfig workspace
    if err := r.Create(ctx, kubeconfigApply); err != nil {
        return ctrl.Result{}, err
    }
}
```

### Risks and Mitigations

1. Risk: Token expiration during cluster operations
   - Mitigation: Implement proper error handling and retry mechanisms

2. Risk: AWS CLI availability in the container
   - Mitigation: Ensure AWS CLI and jq are included in the container image

3. Risk: Increased complexity in workspace management
   - Mitigation: Clear documentation and proper error handling in the controller

## Design Details

### Test Plan

1. Unit Tests
   - Test kubeconfig generation with various input combinations
   - Verify token retrieval functionality
   - Test error handling scenarios

2. Integration Tests
   - Verify WorkspaceTemplate creation and application
   - Test interaction with EKS control plane
   - Validate kubeconfig format and usability

### Graduation Criteria

1. All tests passing
2. Documentation updated
3. Successful deployment in test environments
4. No regression in existing functionality

## Implementation History

- 2024-01-25: Initial proposal

## Alternatives Considered

1. Keep kubeconfig generation in EKS control plane template
   - Rejected due to coupling and maintainability concerns

2. Use exec-based authentication instead of token-based
   - Rejected due to complexity and potential security issues

3. Generate kubeconfig in the controller
   - Rejected to maintain consistency with Terraform-based approach
