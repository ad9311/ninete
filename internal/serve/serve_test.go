package serve_test

import (
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/testhelper"
)

func TestMain(m *testing.M) {
	if code := testhelper.SetUpPackageTest("serve_test"); code > 0 {
		os.Exit(code)
	}

	os.Exit(m.Run())
}
