package handlers

func macroMealTypeOptions() map[string]string {
	return map[string]string{
		"breakfast": "Breakfast",
		"lunch":     "Lunch",
		"dinner":    "Dinner",
		"snack":     "Snack",
		"other":     "Other",
	}
}

func macroAmountUnitOptions() map[string]string {
	return map[string]string{
		"g":    "Grams",
		"ml":   "Milliliters",
		"unit": "Unit (piece)",
		"oz":   "Ounces",
	}
}
