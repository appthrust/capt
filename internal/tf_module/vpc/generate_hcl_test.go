package vpc

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
)

func setupVPCConfigForHCLTest() *VPCConfigBuilder {
	builder := NewVPCConfig()
	builder.SetName("eks-vpc")
	builder.SetCIDR("10.0.0.0/16")
	builder.SetAZsExpression("local.azs")
	builder.SetPrivateSubnetsExpression("[for k, v in local.azs : cidrsubnet(var.vpc_cidr, 4, k)]")
	builder.SetPublicSubnetsExpression("[for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 48)]")
	builder.SetEnableNATGateway(true)
	builder.SetSingleNATGateway(true)
	builder.AddPublicSubnetTag("kubernetes.io/role/elb", "1")
	builder.AddPrivateSubnetTag("kubernetes.io/role/internal-elb", "1")
	return builder
}

func TestGenerateHCL(t *testing.T) {
	builder := setupVPCConfigForHCLTest()
	builder.AddTag("Environment", "dev")
	builder.AddTag("Terraform", "true")

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

	// Read the expected HCL content
	expectedHCL, err := os.ReadFile("testdata/expected_vpc.hcl")
	if err != nil {
		t.Fatalf("Failed to read expected HCL file: %v", err)
	}

	// Compare the generated HCL with the expected content
	if !compareHCL(string(expectedHCL), hcl) {
		t.Errorf("Generated HCL does not match expected HCL.\nExpected:\n%s\nGot:\n%s", string(expectedHCL), hcl)
	}

	// Check if the generated HCL contains expected keys and values
	expectedContent := map[string][]string{
		"module":              {"vpc"},
		"source":              {"terraform-aws-modules/vpc/aws"},
		"version":             {"5.0.0"},
		"name":                {"eks-vpc"},
		"cidr":                {"10.0.0.0/16"},
		"azs":                 {"local.azs"},
		"private_subnets":     {"[for k, v in local.azs : cidrsubnet(var.vpc_cidr, 4, k)]"},
		"public_subnets":      {"[for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 48)]"},
		"enable_nat_gateway":  {"true"},
		"single_nat_gateway":  {"true"},
		"public_subnet_tags":  {"kubernetes.io/role/elb", "1"},
		"private_subnet_tags": {"kubernetes.io/role/internal-elb", "1"},
		"tags":                {"Environment", "dev", "Terraform", "true"},
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

// compareHCL compares two HCL strings, ignoring whitespace differences
func compareHCL(expected, actual string) bool {
	// Remove all whitespace and newlines
	expected = strings.ReplaceAll(strings.ReplaceAll(expected, " ", ""), "\n", "")
	actual = strings.ReplaceAll(strings.ReplaceAll(actual, " ", ""), "\n", "")
	return expected == actual
}
