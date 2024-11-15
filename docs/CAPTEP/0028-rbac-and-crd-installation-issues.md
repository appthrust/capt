# CAPTEP-0028: RBAC and CRD Installation Issues

## Summary

This CAPTEP documents the RBAC permission issues encountered with the CAPTControlPlane controller and the solution implemented to resolve them.

## Motivation

The CAPTControlPlane controller was experiencing permission errors when attempting to list and manage CAPTControlPlane resources. This was preventing proper functionality of the control plane management features.

### Error Message

```
failed to list *v1beta1.CAPTControlPlane: captcontrolplanes.controlplane.cluster.x-k8s.io is forbidden: User "system:serviceaccount:capt-system:capt-controller-manager" cannot list resource "captcontrolplanes" in API group "controlplane.cluster.x-k8s.io" at the cluster scope
```

## Proposal

### Implementation Details

1. Added RBAC markers to the CAPTControlPlane controller:
```go
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes/finalizers,verbs=update
```

2. Regenerated RBAC manifests using controller-gen to include the necessary permissions.

### Testing

1. Applied the updated RBAC configuration
2. Verified that the controller can successfully list and manage CAPTControlPlane resources
3. Confirmed that no permission errors are present in the controller logs

## Implementation History

- 2024-11-15: Initial implementation and documentation

## Alternative Considerations

1. Manually adding permissions to the ClusterRole
   - Rejected because using RBAC markers is the recommended approach and ensures permissions are properly maintained with the code
   - Manual changes could be lost when regenerating manifests

2. Using a more permissive role
   - Rejected due to security best practices of least privilege

## References

- [Kubebuilder RBAC Markers Documentation](https://book.kubebuilder.io/reference/markers/rbac.html)
- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
