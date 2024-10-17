package vpc

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
)

func TestGenerateHCL(t *testing.T) {
	builder := setupVPCConfig()
	builder.SetPrivateSubnets([]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"})
	builder.SetPublicSubnets([]string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"})
	builder.AddTag("Environment", "dev")

	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Unexpected error building VPCConfig: %v", err)
	}

	hcl, err := config.GenerateHCL()
	if err != nil {
		t.Fatalf("GenerateHCL failed: %v", err)
	}

	// Print the generated HCL for inspection
	t.Logf("Generated HCL:\n%s", hcl)

	// Parse the generated HCL
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(hcl), "generated.tf")
	if diags.HasErrors() {
		t.Fatalf("Generated HCL is invalid: %v", diags)
	}

	// Check if the generated HCL contains expected keys and values
	expectedContent := map[string][]string{
		"name":                {"eks-vpc"},
		"cidr":                {"10.0.0.0/16"},
		"azs":                 {"a", "b", "c"},
		"private_subnets":     {"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
		"public_subnets":      {"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"},
		"enable_nat_gateway":  {"true"},
		"single_nat_gateway":  {"true"},
		"public_subnet_tags":  {"kubernetes.io/role/elb", "1"},
		"private_subnet_tags": {"kubernetes.io/role/internal-elb", "1"},
		"tags":                {"Environment", "dev"},
	}

	for key, values := range expectedContent {
		for _, value := range values {
			if !strings.Contains(hcl, key) || !strings.Contains(hcl, value) {
				t.Errorf("Generated HCL does not contain expected key-value pair: %s = %s", key, value)
			}
		}
	}

	// Check for unexpected content
	unexpectedContent := []string{
		"resource",
		"data",
		"variable",
		"output",
	}

	for _, content := range unexpectedContent {
		if strings.Contains(hcl, content) {
			t.Errorf("Generated HCL contains unexpected content: %s", content)
		}
	}
}
