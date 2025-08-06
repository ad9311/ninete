package app

// SetDefaultString returns the default string d if v is empty; otherwise, it returns v.
func SetDefaultString(v, d string) string {
	if v == "" {
		return d
	}

	return v
}
