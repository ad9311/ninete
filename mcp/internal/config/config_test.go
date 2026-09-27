package config_test

import (
	"testing"
	"time"

	"github.com/ad9311/ninete-mcp/internal/config"
	"github.com/stretchr/testify/require"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoad(t *testing.T) {
	const token = "nin_abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_load_a_valid_configuration",
			fn: func(t *testing.T) {
				cfg, err := config.Load(env(map[string]string{
					config.EnvURL:   "https://ninete.example.com/",
					config.EnvToken: token,
					config.EnvTZ:    "America/Bogota",
				}))
				require.NoError(t, err)
				require.Equal(t, "https://ninete.example.com", cfg.BaseURL)
				require.Equal(t, token, cfg.Token)
				require.Equal(t, "America/Bogota", cfg.Location.String())
			},
		},
		{
			name: "should_default_the_zone_to_local",
			fn: func(t *testing.T) {
				cfg, err := config.Load(env(map[string]string{
					config.EnvURL:   "https://ninete.example.com",
					config.EnvToken: token,
				}))
				require.NoError(t, err)
				require.Equal(t, time.Local, cfg.Location)
			},
		},
		{
			name: "should_allow_plain_http_only_on_loopback",
			fn: func(t *testing.T) {
				for _, url := range []string{"http://localhost:8080", "http://127.0.0.1:8080", "http://[::1]:8080"} {
					_, err := config.Load(env(map[string]string{config.EnvURL: url, config.EnvToken: token}))
					require.NoError(t, err, url)
				}

				_, err := config.Load(env(map[string]string{
					config.EnvURL: "http://ninete.example.com", config.EnvToken: token,
				}))
				require.ErrorIs(t, err, config.ErrInsecureURL)
			},
		},
		{
			name: "should_reject_malformed_urls",
			fn: func(t *testing.T) {
				for _, url := range []string{
					"ninete.example.com",
					"ftp://ninete.example.com",
					"https://ninete.example.com/api",
					"https://user:pass@ninete.example.com",
					"https://ninete.example.com?x=1",
				} {
					_, err := config.Load(env(map[string]string{config.EnvURL: url, config.EnvToken: token}))
					require.ErrorIs(t, err, config.ErrInvalidURL, url)
				}

				_, err := config.Load(env(map[string]string{config.EnvToken: token}))
				require.ErrorIs(t, err, config.ErrMissingURL)
			},
		},
		{
			name: "should_reject_missing_or_foreign_tokens",
			fn: func(t *testing.T) {
				_, err := config.Load(env(map[string]string{config.EnvURL: "https://ninete.example.com"}))
				require.ErrorIs(t, err, config.ErrMissingToken)

				_, err = config.Load(env(map[string]string{
					config.EnvURL: "https://ninete.example.com", config.EnvToken: "ghp_something",
				}))
				require.ErrorIs(t, err, config.ErrInvalidToken)
			},
		},
		{
			name: "should_reject_an_unknown_zone",
			fn: func(t *testing.T) {
				_, err := config.Load(env(map[string]string{
					config.EnvURL: "https://ninete.example.com", config.EnvToken: token, config.EnvTZ: "Mars/Olympus",
				}))
				require.ErrorIs(t, err, config.ErrInvalidTZ)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
