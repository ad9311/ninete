package serve

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveHost(t *testing.T) {
	cases := []struct {
		name     string
		set      bool
		host     string
		expected string
	}{
		{
			name:     "should_default_to_loopback_when_unset",
			set:      false,
			expected: defaultHost,
		},
		{
			// Genuine reproduction: reading HOST with os.LookupEnv accepted a
			// bare "HOST=" line as an explicit empty value, which
			// net.JoinHostPort turns into ":8080" — every interface, which is
			// exactly what the loopback default exists to prevent.
			name:     "should_default_to_loopback_when_set_but_empty",
			set:      true,
			host:     "",
			expected: defaultHost,
		},
		{
			name:     "should_honor_an_explicit_host",
			set:      true,
			host:     "0.0.0.0",
			expected: "0.0.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Set first either way so t.Setenv registers the cleanup that
			// restores whatever the ambient environment had.
			t.Setenv("HOST", tc.host)
			if !tc.set {
				require.NoError(t, os.Unsetenv("HOST"))
			}

			require.Equal(t, tc.expected, resolveHost())
		})
	}
}
