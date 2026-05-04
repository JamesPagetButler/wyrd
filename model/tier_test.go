package model

import (
	"encoding/json"
	"testing"
)

func TestTier_String(t *testing.T) {
	cases := []struct {
		tier Tier
		want string
	}{
		{TierComplex, TierNameComplex},
		{TierQuaternion, TierNameQuaternion},
		{TierOctonion, TierNameOctonion},
		{TierSedenion, TierNameSedenion},
	}
	for _, c := range cases {
		if got := c.tier.String(); got != c.want {
			t.Errorf("Tier(%d).String() = %q, want %q", int(c.tier), got, c.want)
		}
	}
}

func TestTier_Components(t *testing.T) {
	cases := []struct {
		tier Tier
		want int
	}{
		{TierComplex, 2},
		{TierQuaternion, 4},
		{TierOctonion, 8},
		{TierSedenion, 16},
	}
	for _, c := range cases {
		if got := c.tier.Components(); got != c.want {
			t.Errorf("Tier(%s).Components() = %d, want %d", c.tier, got, c.want)
		}
	}
}

func TestTier_RoundTripJSON(t *testing.T) {
	for _, original := range []Tier{TierComplex, TierQuaternion, TierOctonion, TierSedenion} {
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal %s: %v", original, err)
		}
		var decoded Tier
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal %s: %v", original, err)
		}
		if decoded != original {
			t.Errorf("round-trip %s: got %s", original, decoded)
		}
	}
}

func TestTier_UnknownNameRejected(t *testing.T) {
	var tier Tier
	if err := json.Unmarshal([]byte(`"hexpentadecation"`), &tier); err == nil {
		t.Error("expected error decoding unknown tier name")
	}
}

func TestTier_InvalidIntegerRejected(t *testing.T) {
	var tier Tier = 99
	if _, err := json.Marshal(tier); err == nil {
		t.Error("expected marshal error for invalid tier value")
	}
}
