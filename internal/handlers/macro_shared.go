package handlers

import "math"

const foodBaseAmountG = 100.0

type selectOption struct {
	Key   string
	Label string
}

func roundMacro(v float64) float64 {
	return math.Round(v*100) / 100
}

func macroAmountUnitOptions() []selectOption {
	return []selectOption{
		{Key: "g", Label: "Grams"},
		{Key: "ml", Label: "Milliliters"},
		{Key: "unit", Label: "Unit (piece)"},
		{Key: "oz", Label: "Ounces"},
	}
}
