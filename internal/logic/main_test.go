package logic_test

import (
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/spec"
)

func TestMain(m *testing.M) {
	if code := spec.SetupPackageTest("logic_test.db"); code > 0 {
		os.Exit(code)
	}

	os.Exit(m.Run())
}
