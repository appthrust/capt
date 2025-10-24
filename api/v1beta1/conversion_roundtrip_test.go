package v1beta1

import (
	"reflect"
	"testing"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	controlplanev1beta2 "github.com/appthrust/capt/api/controlplane/v1beta2"
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
)

func TestCAPTCluster_RoundTrip(t *testing.T) {
	src := &CAPTCluster{
		Spec: CAPTClusterSpec{
			Region:                     "ap-northeast-1",
			VPCTemplateRef:             &WorkspaceTemplateReference{Name: "tpl", Namespace: "capt-system"},
			ExistingVPCID:              "vpc-123",
			RetainVPCOnDelete:          true,
			VPCConfig:                  &VPCConfig{Name: "custom-vpc"},
			WorkspaceTemplateApplyName: "wta-1",
		},
	}

	var hub v1beta2.CAPTCluster
	if err := src.ConvertTo(&hub); err != nil {
		t.Fatalf("ConvertTo hub failed: %v", err)
	}

	var dst CAPTCluster
	if err := dst.ConvertFrom(&hub); err != nil {
		t.Fatalf("ConvertFrom hub failed: %v", err)
	}

	if !reflect.DeepEqual(src.Spec, dst.Spec) {
		t.Fatalf("roundtrip mismatch\nsource=%#v\nresult=%#v", src.Spec, dst.Spec)
	}
}

func TestCAPTControlPlane_RoundTrip(t *testing.T) {
	src := &controlplanev1beta1.CAPTControlPlane{
		Spec: controlplanev1beta1.CAPTControlPlaneSpec{
			Version: "v1.30.4",
			WorkspaceTemplateRef: controlplanev1beta1.WorkspaceTemplateReference{
				Name:      "cp-tpl",
				Namespace: "capt-system",
			},
			ControlPlaneConfig: &controlplanev1beta1.ControlPlaneConfig{
				Region: "eu-west-1",
				EndpointAccess: &controlplanev1beta1.EndpointAccess{
					Public:      true,
					Private:     false,
					PublicCIDRs: []string{"0.0.0.0/0"},
				},
				Addons:   []controlplanev1beta1.Addon{{Name: "vpc-cni", Version: "1.18.2"}},
				Timeouts: &controlplanev1beta1.TimeoutConfig{ControlPlaneTimeout: intPtr(40), VPCReadyTimeout: intPtr(20)},
			},
			AdditionalTags: map[string]string{"env": "e2e"},
			ControlPlaneEndpoint: struct {
				Host string "json:\"host\""
				Port int32  "json:\"port\""
			}{Host: "1.2.3.4", Port: 6443},
			WorkspaceTemplateApplyName: "cp-wta-1",
		},
	}

	var hub controlplanev1beta2.CAPTControlPlane
	if err := src.ConvertTo(&hub); err != nil {
		t.Fatalf("ConvertTo hub failed: %v", err)
	}

	var dst controlplanev1beta1.CAPTControlPlane
	if err := dst.ConvertFrom(&hub); err != nil {
		t.Fatalf("ConvertFrom hub failed: %v", err)
	}

	if !reflect.DeepEqual(src.Spec, dst.Spec) {
		t.Fatalf("roundtrip mismatch\nsource=%#v\nresult=%#v", src.Spec, dst.Spec)
	}
}

func intPtr(v int) *int { return &v }
