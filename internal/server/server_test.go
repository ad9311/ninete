package server_test

import (
	"os"
	"testing"

	"github.com/ad9311/go-api-base/internal/service"
)

func TestMain(m *testing.M) {
	os.Exit(service.RunTestsWithCleanUp(m))
}
