package prog_test

import (
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/stretchr/testify/require"
)

func TestVersionString(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "reports every build variable",
			fn: func(t *testing.T) {
				out := prog.VersionString()

				require.Contains(t, out, "version="+prog.Version)
				require.Contains(t, out, "commit="+prog.Commit)
				require.Contains(t, out, "built="+prog.BuildTime)
			},
		},
		{
			// The -X flags are set by scripts/build.sh and the Makefile, never by
			// `go test`, so an unstamped build must still produce a usable line
			// rather than empty fields.
			name: "falls back to placeholders when unstamped",
			fn: func(t *testing.T) {
				out := prog.VersionString()

				require.NotEmpty(t, prog.Version)
				require.NotEmpty(t, prog.Commit)
				require.NotEmpty(t, prog.BuildTime)
				require.False(t, strings.Contains(out, "= "))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
