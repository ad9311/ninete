package handlers_test

import (
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/spec"
)

func TestMain(m *testing.M) {
	root, ok := prog.FindRoot()
	if !ok {
		os.Exit(1)
	}

	if err := os.Chdir(root); err != nil {
		os.Exit(1)
	}

	if code := spec.SetupPackageTest("handlers_test.db"); code > 0 {
		os.Exit(code)
	}

	os.Exit(m.Run())
}
