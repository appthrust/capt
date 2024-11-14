# CAPTEP-0023: EC2 Spot Service-Linked Role Management

## Summary

This CAPTEP describes the implementation of EC2 Spot Service-Linked Role management in the CAPT controller. The implementation ensures that the required IAM Service-Linked Role for EC2 Spot instances exists before creating EKS clusters.

## Motivation

When creating an EKS cluster that uses Spot instances, AWS requires the existence of a Service-Linked Role named `AWSServiceRoleForEC2Spot`. If this role doesn't exist, cluster creation fails with the error:

```
cannot apply Terraform configuration: Terraform encountered an error. Summary: creating IAM Service Linked Role (spot.amazonaws.com): operation error IAM: CreateServiceLinkedRole, https response error StatusCode: 400, RequestID: f68669ec-fb16-49a9-82bf-0f6920aa66db, InvalidInput: Service role name AWSServiceRoleForEC2Spot has been taken in this account, please try a different suffix.
```

This implementation aims to:
1. Check for the existence of the required role
2. Create the role if it doesn't exist
3. Handle the process in a Kubernetes-native way using WorkspaceTemplates

## Implementation Details

### WorkspaceTemplates

Two WorkspaceTemplates are used:

1. `spot-role-check.yaml`: Checks if the role exists
```yaml
module: |
  data "aws_iam_roles" "spot" {
    name_regex = "^AWSServiceRoleForEC2Spot$"
  }

  output "role_exists" {
    value = length(data.aws_iam_roles.spot.arns) > 0
  }
```

2. `spot-role-create.yaml`: Creates the role if needed
```yaml
module: |
  resource "aws_iam_service_linked_role" "spot" {
    aws_service_name = "spot.amazonaws.com"
    description      = "Service-linked role for EC2 Spot Instances"
  }

  output "role_arn" {
    value = aws_iam_service_linked_role.spot.arn
  }
```

### Controller Implementation

The controller:
1. Creates a WorkspaceTemplateApply for checking role existence
2. Waits for the check workspace to be ready
3. If role doesn't exist, creates another WorkspaceTemplateApply for role creation
4. Waits for the create workspace to be ready and verifies the role ARN

Key points in the implementation:
- Uses WorkspaceTemplateApply for consistent resource management
- Proper error handling and logging
- Status checking with appropriate requeuing
- Owner references for proper resource cleanup

## Test Results

Manual testing confirmed:

1. Role Check:
   - Correctly identifies when role exists/doesn't exist
   - Outputs boolean value properly

2. Role Creation:
   - Successfully creates the role when missing
   - Provides role ARN in outputs
   - Handles existing role case gracefully

3. Error Cases:
   - Handles IAM permission errors
   - Manages deletion restrictions (when role is in use)

## Lessons Learned

1. Service-Linked Role Characteristics:
   - Cannot be deleted while in use by resources (e.g., Spot instances)
   - Deletion is asynchronous with a task ID
   - Must use specific API calls for management

2. Implementation Considerations:
   - Using WorkspaceTemplate provides better abstraction
   - Provider configuration should be template-defined
   - Proper status checking is crucial

## Implementation Impact

This implementation:
1. Improves cluster creation reliability
2. Reduces manual intervention
3. Handles edge cases gracefully

## Alternatives Considered

1. Direct AWS API calls:
   - Rejected due to lack of declarative nature
   - Would require additional AWS SDK dependencies

2. Single Terraform workspace:
   - Rejected due to potential race conditions
   - Separate check/create provides better control

## References

- [AWS Service-Linked Role Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/using-service-linked-roles.html)
- [EKS Spot Instances Documentation](https://docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html#managed-node-group-capacity-types)
