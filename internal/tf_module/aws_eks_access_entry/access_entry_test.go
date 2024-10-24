package aws_eks_access_entry

import (
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestAccessEntryConfig(t *testing.T) {
	// Read desired HCL
	desired, err := os.ReadFile("desired.hcl")
	if err != nil {
		t.Fatalf("Failed to read desired.hcl: %v", err)
	}

	// Format desired HCL
	formattedDesired := string(hclwrite.Format(desired))

	// Generate HCL using default values
	builder := NewAccessEntryConfig()
	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build Access Entry config: %v", err)
	}

	hcl, err := config.GenerateHCL()
	if err != nil {
		t.Fatalf("Failed to generate HCL: %v", err)
	}

	// Format generated HCL
	formattedGenerated := string(hclwrite.Format([]byte(hcl)))

	if formattedGenerated != formattedDesired {
		t.Errorf("Generated HCL does not match desired HCL:\nGenerated:\n%s\nDesired:\n%s", formattedGenerated, formattedDesired)
	}
}

func TestAccessEntryConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		modifyBuilder func(*AccessEntryConfigBuilder)
		expectError   bool
	}{
		{
			name: "valid config with default values",
			modifyBuilder: func(b *AccessEntryConfigBuilder) {
				// Default values should be valid
			},
			expectError: false,
		},
		{
			name: "valid config with custom values",
			modifyBuilder: func(b *AccessEntryConfigBuilder) {
				b.SetClusterName("custom.cluster.name").
					SetPrincipalARN("custom:arn").
					SetType("EC2_LINUX").
					SetKubernetesGroups([]string{"system:masters"})
			},
			expectError: false,
		},
		{
			name: "empty cluster name",
			modifyBuilder: func(b *AccessEntryConfigBuilder) {
				b.config.ClusterName = nil
			},
			expectError: true,
		},
		{
			name: "empty principal ARN",
			modifyBuilder: func(b *AccessEntryConfigBuilder) {
				b.config.PrincipalARN = nil
			},
			expectError: true,
		},
		{
			name: "empty type",
			modifyBuilder: func(b *AccessEntryConfigBuilder) {
				b.SetType("")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewAccessEntryConfig()
			tt.modifyBuilder(builder)
			_, err := builder.Build()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
