package service

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(RunWithIsolatedSchema(m, "service"))
}
