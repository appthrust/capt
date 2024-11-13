# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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