package types

import (
	tmmath "github.com/Finschia/ostracon/libs/math"
	"github.com/Finschia/ostracon/light"
)

// DefaultTrustLevel is the tendermint light client default trust level
var DefaultTrustLevel = NewFractionFromTm(light.DefaultTrustLevel)

// NewFractionFromTm returns a new Fraction instance from a tmmath.Fraction
func NewFractionFromTm(f tmmath.Fraction) Fraction {
	return Fraction{
		Numerator:   f.Numerator,
		Denominator: f.Denominator,
	}
}

// ToTendermint converts Fraction to tmmath.Fraction
func (f Fraction) ToTendermint() tmmath.Fraction {
	return tmmath.Fraction{
		Numerator:   f.Numerator,
		Denominator: f.Denominator,
	}
}
