# TODO List

## Control Plane Endpoint Management

### WorkspaceTemplateApply Output Requirements
- [ ] Define required output format for EKS endpoint in WorkspaceTemplate
- [ ] Document the expected output key name ("endpoint") in WorkspaceTemplate design
- [ ] Add validation for endpoint output format (must be a valid hostname)

### CAPTControlPlane Controller
- [ ] Add error handling for missing endpoint output
- [ ] Add logging for endpoint retrieval and updates
- [ ] Consider adding timeout for endpoint availability
- [ ] Add status condition for endpoint availability

### Cluster Integration
- [ ] Verify endpoint propagation to owner Cluster in all scenarios
- [ ] Add proper error handling when endpoint update fails
- [ ] Consider adding readiness check before setting endpoint

## Documentation
- [ ] Update WorkspaceTemplate examples with endpoint output
- [ ] Document endpoint propagation flow in architecture docs
- [ ] Add troubleshooting guide for endpoint issues
