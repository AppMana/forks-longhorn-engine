package meta

import (
	"runtime"
	"slices"
	"testing"
)

func TestPlatformCapabilities(t *testing.T) {
	capabilities := GetVersion().Capabilities
	if !slices.Contains(capabilities.Controller, CapabilityV1) {
		t.Fatal("controller does not advertise V1")
	}
	if !slices.Contains(capabilities.Replica, CapabilityRWO) {
		t.Fatal("replica does not advertise RWO")
	}
	if runtime.GOOS == "windows" {
		for _, unsupported := range []string{CapabilityRWX, CapabilityEncryption} {
			if slices.Contains(capabilities.Replica, unsupported) {
				t.Fatalf("Windows replica unexpectedly advertises %q", unsupported)
			}
		}
		for role, advertised := range map[string][]string{
			"controller": capabilities.Controller,
			"replica":    capabilities.Replica,
		} {
			if !slices.Contains(advertised, CapabilityStrictLocal) {
				t.Fatalf("Windows %s does not advertise strict-local", role)
			}
		}
		if !slices.Contains(capabilities.Frontend, CapabilityReFS) {
			t.Fatal("Windows frontend does not advertise ReFS")
		}
	} else {
		for _, filesystem := range []string{CapabilityExt4, CapabilityXFS} {
			if !slices.Contains(capabilities.Frontend, filesystem) {
				t.Fatalf("Linux frontend does not advertise %q", filesystem)
			}
		}
	}
}
