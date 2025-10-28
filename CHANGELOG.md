# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]


## [v0.5.0] - 2025-10-26

### Added
- Webhooks/Certificates: Align Admission/Conversion configuration with Kubebuilder best practices; resolve Service references via kustomize nameReference/namespace mapping.
- Add name-merge patches to `capt-*-webhook-configuration` to enforce `clientConfig.service = capt-webhook-service/capt-system`.
- Standardize CA injection into MWC/VWC and CRD conversion via cert-manager `inject-ca-from` annotations.
- Add `make wait-ca` to wait until caBundle injection completes for both Admission and CRD conversion.
- e2e: Add readiness waits before admission (Deployment rollout, TLS Secret, Service Endpoints, short settle time).
- Samples: Update to CAPI v1beta2 (Cluster `spec.topology.classRef.name`, ClusterClass `*.templateRef`, KubeadmConfigTemplate v1beta2 minimal no-op).
- API: Expose Terraform workspace name for observability via `status.workspaceTemplateStatus.workspaceName` on `CAPTControlPlane` and `CAPTCluster`.

### Changed
- Conversion Webhook: Remove manual `/convert` registration; rely on controller-runtime auto-registration.
- Controllers: Avoid mutating `spec.workspaceTemplateApplyName` when resources are managed by ClusterTopology. A deterministic name is resolved internally and the resulting workspace name is surfaced in `status.workspaceTemplateStatus.workspaceName`.
- Tests: Updated unit tests to assert deterministic naming and status-based introspection rather than spec mutation side-effects.

### Deprecated
- Field usage: Reliance on `spec.workspaceTemplateApplyName` by controllers under ClusterTopology is deprecated. The field remains for compatibility but is not written by controllers and may be removed in a future release.

### Fixed
- Resolve `unknown authority` (missing CA) and connection failures caused by mismatched webhook Service name/namespace.

### Notes
- Known caveat: Plan to strengthen immutability validation on `CAPTControlPlane` (forbid updates to immutable fields).
- Webhooks: Strengthening immutability validation for `CAPTControlPlane` on `v1beta2` is planned as a follow-up to cover all served versions consistently.

## [v0.4.0] - 2025-10-24

### Added
- Add ClusterClass support (aligned with the clusterctl ClusterClass flow)

### Changed
- Clarify and tidy up CAPI contract labels in manifests (v1beta1)
- Update image tags in `config/*/kustomization.yaml` to `v0.4.0`

### Notes
- Heads up: v0.5.0 will implement v1beta2 compatibility

## [v0.2.1] - 2024-01-25

### Added
- CAPTEP-0046: Release planning and documentation for v0.2.1
- Added detailed release process documentation
- Added troubleshooting guide for GitHub Actions releases

### Changed
- Improved release process with clear documentation
- Enhanced GitHub Actions workflow documentation

## [v0.2.0] - 2024-01-25

### Added
- CAPTEP-0031: FluxCD integration with ClusterResourceSet
- CAPTEP-0032: Variable resolution mechanism for ClusterResourceSet
- CAPTEP-0033: Migration from Crossplane to Upbound Terraform Provider
- CAPTEP-0034: Dedicated WorkspaceTemplate for kubeconfig generation
- CAPTEP-0035: Status condition helper functions
- CAPTEP-0036: Kubeconfig generation improvements and design decisions
- CAPTEP-0037: Kubeconfig secret update mechanism improvements
- CAPTEP-0040: Added WorkspaceStatus to track Workspace state and atProvider details
- CAPTEP-0041: Documentation for cluster endpoint cleanup during deletion
- CAPTEP-0042: Migration of Karpenter installation to HelmChartProxy
- CAPTEP-0043: Detailed design and implementation of Karpenter HelmChartProxy migration
- CAPTEP-0044: Analysis and discussion of Go template support in WorkspaceTemplate
- CAPTEP-0045: Release planning and documentation for v0.2.0
- Added new eks-controlplane-template-without-karpenter for Terraform-only infrastructure
- Added HelmChartProxy manifests for Karpenter and default NodePool installation
- Added demo-cluster-with-helm sample for HelmChartProxy-based Karpenter installation

### Changed
- Updated RBAC permissions to use tf.upbound.io API group instead of tf.crossplane.io
- Migrated Terraform workspace management to use Upbound provider
- Separated kubeconfig generation into a dedicated WorkspaceTemplate
- Refactored EKS control plane template to remove kubeconfig generation
- Refactored status condition handling to use common helper functions
- Enhanced kubeconfig secret management to support automatic updates
- Migrated Karpenter installation from Terraform to HelmChartProxy
- Separated Karpenter core installation and NodePool configuration
- Improved Karpenter installation reliability with fixed release names and namespace isolation

### Fixed
- Added missing RBAC permissions for clusters/status resource
- CAPTEP-0033: Documentation for RBAC permissions update
- Implemented automatic updates for kubeconfig secrets to ensure latest configuration
- CAPTEP-0038: Added missing RBAC permissions for kubeconfig secret creation and updates
- Fixed cluster endpoint not being cleared during CAPTControlPlane deletion
- Improved endpoint update handling using patch-based updates to prevent race conditions
- Fixed Workspace deletion order in CAPTControlPlane cleanup to ensure proper resource cleanup
- Enhanced error handling and retry mechanism for Workspace deletion confirmation
- Fixed Karpenter installation issues with proper dependency management and namespace configuration

## [v0.1.11] - 2024-01-24

### Fixed
- Added missing RBAC permissions for clusters/status resource
- CAPTEP-0033: Documentation for RBAC permissions update

## [v0.1.10] - 2024-11-15

### Added
- CAPTEP-0026: Documentation for release workflow optimization
- Improved GitHub Container Registry integration
- Enhanced release automation process
- CAPTEP-0028: Documentation for RBAC and CRD installation issues

### Fixed
- RBAC permissions for CAPTControlPlane controller
- Added missing RBAC markers for proper permission generation

### Changed
- Updated generated API files
- Updated configuration files and RBAC roles

## [v0.1.6] - 2024-11-15

### Added
- Quick Start Guide in README.md for easier CAPT deployment and integration with Cluster API
- Detailed CRD analysis and documentation in CAPTEP-0025
- Comprehensive RBAC security analysis
- Controller configuration documentation including resource limits and health checks

### Enhanced
- Updated CAPTEP-0025 with detailed architecture and design decisions
- Improved documentation of resource relationships and dependencies
- Added security considerations and best practices

## [v0.1.5] - 2024-11-15

### Added
- Support for customizing VPC name through CaptCluster spec
- CAPTEP-0024: Documentation for VPC name customization
- CAPTEP-0025: Analysis and plan for releasing CAPT as Cluster API Provider

### Changed
- Updated vpc.yaml template to use ${vpc_name} and ${cluster_name} variables
- Modified VPC naming to use cluster name as default with -vpc suffix
- Enhanced error handling and logging for EC2 Spot Service-Linked Role management
- Improved Cluster API specification compliance for CAPTCluster and CAPTControlPlane

### Improved
- Optimized WorkspaceTemplateApply processing for better performance
- Strengthened integration with Karpenter for Infrastructure Provider

## [v0.1.4] - 2024-11-14

### Changed
- Improved EC2 Spot Service-Linked Role handling using WorkspaceTemplate
- Separated role management from EKS template for better control

### Added
- CAPTEP-0023: Documentation for improved EC2 Spot Service-Linked Role handling

## [v0.1.3] - 2024-11-13

### Fixed
- Fixed EC2 Spot Instance provisioning failure
- Added automatic creation of Service-Linked Role (skips if role already exists)

### Added
- CAPTEP-0023: Documentation for EC2 Spot Instance Service-Linked Role automation

## [v0.1.1] - 2024-11-13

### Fixed
- Variable references in kubeconfig output to use `var.region` format
- WorkspaceTemplate to properly handle variable expansion in Terraform modules

### Added
- CAPTEP-0021 documenting the variable expansion issue and solution

## [v0.1.0] - 2024-11-12

### Added
- Initial release of CAPT
- Basic EKS cluster management functionality
- Support for Terraform-based infrastructure provisioning
- WorkspaceTemplate and WorkspaceTemplateApply controllers
- CAPTControlPlane implementation
