package task

import (
	"encoding/csv"
	"os"
)

// openCSVFile opens a CSV file at the given path and returns its records.
func openCSVFile(path string) ([][]string, error) {
	var records [][]string
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return records, err
	}

	reader := csv.NewReader(f)

	records, err = reader.ReadAll()
	if err != nil {
		return records, err
	}

	err = f.Close()
	if err != nil {
		return [][]string{}, err
	}

	return records, err
}
