package model

import "testing"

func TestRuntimeAnchorPrefix_AllCanonicalsMatch(t *testing.T) {
	cases := []NodeType{
		NodeTypeBMARuntimeFlagNormDrift,
		NodeTypeBMARuntimeObsZDDetected,
		NodeTypeBMARuntimeObsCounter,
		NodeTypeBMARuntimeObsFault,
	}
	for _, c := range cases {
		if !IsRuntimeAnchor(c) {
			t.Errorf("IsRuntimeAnchor(%q) = false, want true", c)
		}
	}
}

func TestIsRuntimeAnchor_NonRuntimeRejected(t *testing.T) {
	cases := []NodeType{
		"",
		"bma.engram.tier-1.semantic",
		"cth.anchor.measurement",
		"contextus.signal",
		"contextus.scope.physical",
		"wyrd.internal",
		"bma.runtime",  // exact prefix without trailing "."; not a runtime anchor type
		"bma.runtimer", // visually similar but distinct namespace
	}
	for _, c := range cases {
		if IsRuntimeAnchor(c) {
			t.Errorf("IsRuntimeAnchor(%q) = true, want false (not under reserved prefix)", c)
		}
	}
}

func TestIsRuntimeAnchor_FutureAdditionsClassified(t *testing.T) {
	// IsRuntimeAnchor classifies any value under the reserved prefix,
	// not only the four canonical constants. This lets future M1+M2
	// additions register without a code change to consumers.
	future := NodeType("bma.runtime.obs-fault-recovered")
	if !IsRuntimeAnchor(future) {
		t.Errorf("IsRuntimeAnchor(%q) = false, want true (under reserved prefix)", future)
	}
}

func TestRuntimeAnchorConstants_StableValues(t *testing.T) {
	// Pin the exact string values; downstream consumers (CTH ρ_net
	// classifier, BMA observer) match on these. Drift here is a
	// breaking change.
	want := map[NodeType]string{
		RuntimeAnchorPrefix:             "bma.runtime.",
		NodeTypeBMARuntimeFlagNormDrift: "bma.runtime.flag-norm-drift",
		NodeTypeBMARuntimeObsZDDetected: "bma.runtime.obs-zd-detected",
		NodeTypeBMARuntimeObsCounter:    "bma.runtime.obs-runtime-counter",
		NodeTypeBMARuntimeObsFault:      "bma.runtime.obs-fault",
	}
	for k, v := range want {
		if string(k) != v {
			t.Errorf("constant value drift: got %q, want %q", string(k), v)
		}
	}
}
