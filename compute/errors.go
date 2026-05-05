package compute

import (
	"errors"
	"fmt"

	"github.com/JamesPagetButler/wyrd/model"
)

// ErrTierMixed is returned when an operation requires both operands at
// the same tier but receives a mismatch.
var ErrTierMixed = errors.New("wyrd: tier mismatch")

// ErrTierUnsupported is returned when an operation is requested at a
// tier whose backend is not yet implemented.
var ErrTierUnsupported = errors.New("wyrd: tier not yet supported")

// ErrInvalidTier is returned when a Tier value is not one of the four
// canonical constants.
var ErrInvalidTier = errors.New("wyrd: invalid tier")

func errMixedTiers(a, b model.Tier) error {
	return fmt.Errorf("%w: %s vs %s", ErrTierMixed, a, b)
}

func errTierUnsupported(t model.Tier) error {
	return fmt.Errorf("%w: %s", ErrTierUnsupported, t)
}

func errInvalidTier(t model.Tier) error {
	return fmt.Errorf("%w: %v", ErrInvalidTier, t)
}
