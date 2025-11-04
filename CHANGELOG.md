# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.3] - 2025-11-04

### Changed
- Relaxed kubeconfig generation dependencies; removed EKS Workspace readiness wait and `WaitForWorkspaces` requirement.
- Switched the kubeconfig WorkspaceTemplate to fetch `endpoint` and `certificate_authority.data` via Terraform `data "aws_eks_cluster"`.
- Expanded region resolution precedence (`ControlPlaneConfig.Region` → Cluster topology variable `region` → annotation `cluster.x-k8s.io/region` → annotation `installation.appthrust.com/aws-primary-region`).

### Notes
- Resolved the circular dependency that caused a deadlock between EKS provisioning and kubeconfig retrieval.

## [v0.4.1] - 2025-10-28

### Added
 - CAPTCluster.status.workspaceStatus (v1beta1): Expose Workspace Ready/State/AtProvider fields.
 - CAPTCluster controller: Add collection logic for atProvider from Workspace.

### Changed
 - Switched CAPTCluster status updates to patch-based to improve conflict resilience and preserve atProvider.
 - Reworked ClusterClass samples to the v1beta1 structure (full compliance with the 0.4.x contract).

### Notes
 - Dependencies remain on CAPI v1beta1; ClusterClass maintains full v1beta1 compatibility.
 - Generated artifacts (DeepCopy/CRDs) are updated; apply with `make install`.

## [Unreleased]

## [v0.4.0] - 2025-10-24

### Added
 - Added ClusterClass support (compatible with clusterctl's ClusterClass workflow).

### Changed
 - Clarified and organized CAPI contract labels in manifests (v1beta1).
 - Updated image tags in `config/*/kustomization.yaml` to `v0.4.0`.

### Notes
 - v1beta2 compatibility planned for `v0.5.0`.

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
