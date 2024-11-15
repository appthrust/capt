# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CAPTEP-0029: Standardization of Enhancement Proposal format
- ADR-0002: Decision to use Enhancement Proposals for significant changes
- CAPTEP-0030: Proposal for improving Karpenter installation reliability using ClusterResourceSet

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
