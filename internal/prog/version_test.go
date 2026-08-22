package prog_test

import (
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
			// Invariant guard, not a bug reproduction. `go test` passes no
			// -ldflags (see the test target in the Makefile), so these variables
			// hold their declared defaults here and the exact fallbacks can be
			// asserted. Nothing may depend on the stamp being present, so the
			// fallbacks must stay non-empty and must not drift.
			name: "falls back to placeholders when unstamped",
			fn: func(t *testing.T) {
				require.Equal(t, "dev", prog.Version)
				require.Equal(t, "unknown", prog.Commit)
				require.Equal(t, "unknown", prog.BuildTime)

				require.Equal(
					t,
					"version=dev commit=unknown built=unknown",
					prog.VersionString(),
				)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
