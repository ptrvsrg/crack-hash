package helper

import (
	"errors"
	"math"
	"math/big"
)

var (
	ErrDivisionByZero = errors.New("division by zero: common ratio r cannot be 1 in this case")
	ErrIntLimits      = errors.New("result exceeds the limits of int")
)

func SumOfGeomSeries(a, r, n int) (int, error) {
	// Convert inputs to big.Int for handling large numbers
	bigA := big.NewInt(int64(a))
	bigR := big.NewInt(int64(r))
	bigN := big.NewInt(int64(n))

	one := big.NewInt(1)

	// Check if the common ratio r is equal to 1
	if r == 1 {
		sum := big.NewInt(0).Mul(bigA, bigN)
		return convertBigToInt(sum)
	}

	// Calculate r^n
	rPowerN := new(big.Int).Exp(bigR, bigN, nil)

	// Calculate numerator: 1 - r^n
	numerator := new(big.Int).Sub(one, rPowerN)

	// Calculate denominator: 1 - r
	denominator := new(big.Int).Sub(one, bigR)

	// Check for division by zero
	if denominator.Cmp(big.NewInt(0)) == 0 {
		return 0, ErrDivisionByZero
	}

	// Calculate quotient: (1 - r^n) / (1 - r)
	quotient := new(big.Int).Div(numerator, denominator)

	// Multiply by the first term a
	result := new(big.Int).Mul(bigA, quotient)

	return convertBigToInt(result)
}

func convertBigToInt(num *big.Int) (int, error) {
	if !isWithinIntLimits(num) {
		return 0, ErrIntLimits
	}
	return int(num.Int64()), nil
}

func isWithinIntLimits(num *big.Int) bool {
	minInt := big.NewInt(math.MinInt)
	maxInt := big.NewInt(math.MaxInt)
	return num.Cmp(minInt) >= 0 && num.Cmp(maxInt) <= 0
}
