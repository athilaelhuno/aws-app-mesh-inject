package patch

import (
	"strings"
	"testing"
)

func TestGeneratePatch_AppendSidecarFalse(t *testing.T) {
	meta := Meta{
		AppendImagePullSecret: false,
		HasImagePullSecret:    false,
		AppendSidecar:         false,
		AppendInit:            false,
		Init: InitMeta{
			Ports:              "80,443",
			EgressIgnoredPorts: "22",
			ContainerImage:     "059588584554.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-proxy-route-manager:v2",
			IgnoredIPs:         "169.254.169.254",
		},
		Sidecar: SidecarMeta{
			MeshName:          "global",
			VirtualNodeName:   "podinfo",
			Region:            "us-west-2",
			LogLevel:          "debug",
			ContainerImage:    "111345817488.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-envoy:v1.11.1.0-prod",
			CpuRequests:       "10m",
			MemoryRequests:    "32Mi",
			InjectXraySidecar: true,
			EnableStatsTags:   true,
		},
	}

	patch, err := GeneratePatch(meta)
	if err != nil {
		t.Fatal(err)
	}

	verifyPatch(t, string(patch), meta)
}

func TestGeneratePatch_AppendSidecarTrue(t *testing.T) {
	meta := Meta{
		AppendImagePullSecret: false,
		HasImagePullSecret:    false,
		AppendSidecar:         true,
		AppendInit:            false,
		Init: InitMeta{
			Ports:              "80,443",
			EgressIgnoredPorts: "22",
			ContainerImage:     "059588584554.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-proxy-route-manager:v2",
			IgnoredIPs:         "169.254.169.254",
		},
		Sidecar: SidecarMeta{
			MeshName:          "global",
			VirtualNodeName:   "podinfo",
			Region:            "us-west-2",
			LogLevel:          "debug",
			ContainerImage:    "111345817488.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-envoy:v1.11.1.0-prod",
			CpuRequests:       "10m",
			MemoryRequests:    "32Mi",
			InjectXraySidecar: true,
			EnableStatsTags:   true,
		},
	}

	patch, err := GeneratePatch(meta)
	if err != nil {
		t.Fatal(err)
	}

	verifyPatch(t, string(patch), meta)
}

func verifyPatch(t *testing.T, patch string, meta Meta) {
	if !strings.Contains(patch, "aws-appmesh-proxy-route-manager:v2") {
		t.Errorf("Init container image not found")
	}

	if !strings.Contains(patch, meta.Sidecar.ContainerImage) {
		t.Errorf("Sidecar container image not found")
	}

	if meta.Sidecar.InjectXraySidecar {
		if !strings.Contains(patch, "amazon/aws-xray-daemon") {
			t.Errorf("No x-ray found")
		}
	} else {
		if strings.Contains(patch, "amazon/aws-xray-daemon") {
			t.Errorf("X-Ray container found when InjectXraySidecar=false")
		}
	}
}
