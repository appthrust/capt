# VPC Configuration Migration Tasks

This document outlines the tasks required to migrate the VPC configuration from the current VPCConfig-based approach to using WorkspaceTemplate and WorkspaceTemplateApply.

## Background

Currently, the CAPT Cluster creates VPCs using the VPCConfig specification defined in CAPTClusterSpec. This approach tightly couples VPC configuration with the cluster configuration. Moving to a WorkspaceTemplate-based approach will provide better separation of concerns and reusability.

## Current Implementation

The current implementation uses:
- VPCConfig embedded in CAPTClusterSpec
- Direct VPC creation through the cluster controller
- Tightly coupled VPC and cluster lifecycle

## Target Implementation

The new implementation will:
- Use WorkspaceTemplate to define VPC configurations
- Use WorkspaceTemplateApply to manage VPC lifecycle
- Decouple VPC management from cluster lifecycle
- Enable VPC reuse across multiple clusters
- Support better resource management and cleanup

## Migration Tasks

### 1. API Changes

1.1. CAPTClusterSpec Updates:
- Remove VPCConfig from CAPTClusterSpec
- Add WorkspaceTemplateRef field to reference VPC WorkspaceTemplate
- Add optional VPCSelector field for existing VPC selection

1.2. Create VPC-specific WorkspaceTemplate:
- Define standard VPC template structure
- Include all current VPCConfig options
- Add additional metadata for better resource tracking

### 2. Controller Changes

2.1. CAPTCluster Controller:
- Remove direct VPC creation logic
- Add WorkspaceTemplateApply creation/management
- Implement VPC lookup/selection logic
- Update reconciliation flow to handle WorkspaceTemplate references

2.2. WorkspaceTemplate Controller:
- Enhance validation for VPC-specific templates
- Add VPC-specific status conditions
- Implement VPC resource tracking

### 3. Infrastructure Changes

3.1. Terraform Configurations:
- Migrate VPC Terraform configurations to WorkspaceTemplate format
- Update module references and dependencies
- Add support for VPC resource sharing

3.2. Resource Management:
- Implement VPC lifecycle management through WorkspaceTemplateApply
- Add cleanup and garbage collection for orphaned resources
- Handle VPC dependencies and references

### 4. Testing

4.1. Unit Tests:
- Update existing tests to reflect new VPC management
- Add tests for WorkspaceTemplate VPC configurations
- Add tests for VPC reference handling

4.2. Integration Tests:
- Add tests for VPC creation through WorkspaceTemplate
- Test VPC sharing across clusters
- Verify cleanup and resource management

4.3. E2E Tests:
- Update e2e test scenarios for new VPC management
- Add migration test cases
- Test failure scenarios and recovery

### 5. Documentation

5.1. User Documentation:
- Document new VPC configuration approach
- Provide migration guide for existing users
- Add examples and best practices

5.2. API Documentation:
- Update API reference for new fields
- Document WorkspaceTemplate VPC specifications
- Add migration-specific documentation

### 6. Migration Support

6.1. Migration Tools:
- Create tools to convert existing VPCConfig to WorkspaceTemplate
- Add validation for converted configurations
- Provide rollback mechanisms

6.2. Compatibility:
- Implement temporary backward compatibility
- Add deprecation warnings for VPCConfig
- Plan deprecation timeline

## Implementation Phases

1. **Phase 1: Development**
   - Implement API changes
   - Update controllers
   - Add basic testing

2. **Phase 2: Testing**
   - Comprehensive testing
   - Bug fixes and improvements
   - Documentation updates

3. **Phase 3: Migration**
   - Release migration tools
   - Support user migration
   - Monitor and address issues

4. **Phase 4: Cleanup**
   - Remove deprecated code
   - Finalize documentation
   - Complete migration support

## Success Criteria

1. All VPC management successfully migrated to WorkspaceTemplate
2. No regression in existing functionality
3. Improved resource management and reusability
4. Complete test coverage
5. Comprehensive documentation
6. Smooth migration path for existing users

## Risks and Mitigation

1. **Risk**: Breaking changes for existing users
   - **Mitigation**: Provide clear migration path and tools

2. **Risk**: Resource management complexity
   - **Mitigation**: Comprehensive testing and monitoring

3. **Risk**: Performance impact
   - **Mitigation**: Benchmark and optimize as needed

4. **Risk**: Migration failures
   - **Mitigation**: Implement rollback mechanisms

## Timeline

- Development: 2-3 weeks
- Testing: 1-2 weeks
- Documentation: 1 week
- Migration Support: Ongoing

## Next Steps

1. Review and approve migration plan
2. Create detailed technical design
3. Begin implementation of Phase 1
4. Schedule regular progress reviews
