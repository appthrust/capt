package eks_blueprints_addons_v2

import (
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestEKSBlueprintsAddonsConfig(t *testing.T) {
	// Read desired HCL
	desired, err := os.ReadFile("desired_hcl_output.hcl")
	if err != nil {
		t.Fatalf("Failed to read desired_hcl_output.hcl: %v", err)
	}

	// Format desired HCL
	formattedDesired := string(hclwrite.Format(desired))

	// Generate HCL using default values
	builder := NewEKSBlueprintsAddonsConfig()
	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build EKS Blueprints Addons config: %v", err)
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

func TestEKSBlueprintsAddonsConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		modifyBuilder func(*EKSBlueprintsAddonsConfigBuilder)
		expectError   bool
	}{
		{
			name: "valid config with default values",
			modifyBuilder: func(b *EKSBlueprintsAddonsConfigBuilder) {
				// Default values should be valid
			},
			expectError: false,
		},
		{
			name: "valid config with custom values",
			modifyBuilder: func(b *EKSBlueprintsAddonsConfigBuilder) {
				b.SetEnableKarpenter(false).
					SetKarpenterHelmCacheDir("/custom/cache/dir").
					SetKarpenterNodeConfig(true)
			},
			expectError: false,
		},
		{
			name: "valid config with custom CoreDNS values",
			modifyBuilder: func(b *EKSBlueprintsAddonsConfigBuilder) {
				b.SetCoreDNSConfig(`{"computeType":"EC2"}`)
			},
			expectError: false,
		},
		{
			name: "valid config with custom cluster settings",
			modifyBuilder: func(b *EKSBlueprintsAddonsConfigBuilder) {
				b.SetClusterName("custom-cluster").
					SetClusterEndpoint("https://custom-endpoint").
					SetClusterVersion("1.24").
					SetOIDCProviderARN("arn:aws:iam::123456789012:oidc-provider")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewEKSBlueprintsAddonsConfig()
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
