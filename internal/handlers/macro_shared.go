package handlers

import "math"

const foodBaseAmountG = 100.0

func roundMacro(v float64) float64 {
	return math.Round(v*100) / 100
}
