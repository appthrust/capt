# Introduction of Machine Concept

## Background

In the current CAPT implementation, compute resources (Fargate profiles, managed node groups, etc.) are defined as part of the ControlPlane resource. This creates the following challenges:

- Lack of flexibility in compute resource management
- Difficulties in scaling and lifecycle management
- Deviation from standard Cluster API (CAPI) patterns

## Proposal

We propose introducing the Machine concept to CAPT, managing compute resources as independent resources.

### Key Changes

1. Introduction of New Custom Resources
   - CAPTMachineTemplate
   - CAPTMachineDeployment
   - CAPTMachine

2. Separation of WorkspaceTemplates
   - WorkspaceTemplate for ControlPlane
   - WorkspaceTemplate for Machine

3. Restructuring of Terraform Modules
   - Separation of node group configurations from eks module
   - Creation of new module for node groups

## Expected Benefits

1. Improved Operability
   - Enable individual management of node groups
   - Flexible scaling control
   - Enhanced lifecycle management

2. Architecture Improvements
   - Clear separation of responsibilities
   - Alignment with CAPI patterns
   - Enhanced future extensibility
