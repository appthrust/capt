package eks_v2

import (
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestEKSConfig(t *testing.T) {
	// Read desired HCL
	desired, err := os.ReadFile("desired.hcl")
	if err != nil {
		t.Fatalf("Failed to read desired.hcl: %v", err)
	}

	// Format desired HCL
	formattedDesired := string(hclwrite.Format(desired))

	// Generate HCL
	builder := NewEKSConfig()
	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build EKS config: %v", err)
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

func TestEKSConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		modifyBuilder func(*EKSConfigBuilder)
		expectError   bool
	}{
		{
			name:          "valid config",
			modifyBuilder: func(b *EKSConfigBuilder) {},
			expectError:   false,
		},
		{
			name: "empty cluster version",
			modifyBuilder: func(b *EKSConfigBuilder) {
				b.SetClusterVersion("")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewEKSConfig()
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
