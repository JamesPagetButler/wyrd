package compute

import (
	"errors"
	"testing"

	"github.com/JamesPagetButler/wyrd/model"
)

func TestCanSynthesize_DownwardOK(t *testing.T) {
	cases := []struct {
		caller, target model.Tier
	}{
		{model.TierQuaternion, model.TierComplex},
		{model.TierOctonion, model.TierQuaternion},
		{model.TierSedenion, model.TierOctonion},
		{model.TierSedenion, model.TierComplex},
		{model.TierComplex, model.TierComplex},
	}
	for _, c := range cases {
		if err := CanSynthesize(c.caller, c.target); err != nil {
			t.Errorf("CanSynthesize(%s, %s): unexpected error %v", c.caller, c.target, err)
		}
	}
}

func TestCanSynthesize_UpwardBlocked(t *testing.T) {
	cases := []struct {
		caller, target model.Tier
	}{
		{model.TierComplex, model.TierQuaternion},  // T2.1.a — ℂ ↛ ℍ
		{model.TierQuaternion, model.TierOctonion}, // T2.1.b — ℍ ↛ 𝕆
		{model.TierOctonion, model.TierSedenion},   // T2.1.c — 𝕆 ↛ 𝕊
		{model.TierComplex, model.TierSedenion},    // transitive
	}
	for _, c := range cases {
		err := CanSynthesize(c.caller, c.target)
		if err == nil {
			t.Errorf("CanSynthesize(%s → %s): expected privilege violation", c.caller, c.target)
			continue
		}
		if !errors.Is(err, ErrPrivilegeViolation) {
			t.Errorf("CanSynthesize(%s → %s): error %v not ErrPrivilegeViolation", c.caller, c.target, err)
		}
	}
}
