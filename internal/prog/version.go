package prog

import "fmt"

// Build identity, injected at link time with `-X`. `scripts/build.sh` sets them
// for production builds and the Makefile sets them for local ones; both derive
// the values from git.
//
// The defaults are what an unstamped build reports — `go test`, `go run`, an
// editor's build, a `go build` typed by hand. Nothing may depend on the flags
// being present, so every value stays a plain string with a usable fallback.
// nolint on each: -X can only write to a package-level string variable, so these
// cannot be constants or live behind a function.
var (
	// Version is `git describe --tags --always --dirty`: the nearest tag, the
	// number of commits since it, the abbreviated commit, and a `-dirty` suffix
	// when the checkout had uncommitted changes.
	Version = "dev" //nolint:gochecknoglobals // link-time build stamp

	// Commit is the abbreviated commit hash on its own. Version already carries
	// it unless the build sat exactly on a tag, which is the case where knowing
	// the commit matters most — a tag can be moved, a commit cannot.
	Commit = "unknown" //nolint:gochecknoglobals // link-time build stamp

	// BuildTime is when the binary was linked, RFC 3339 in UTC.
	BuildTime = "unknown" //nolint:gochecknoglobals // link-time build stamp
)

// VersionString renders the build identity as one line, for a boot log or a
// `version` command.
func VersionString() string {
	return fmt.Sprintf("version=%s commit=%s built=%s", Version, Commit, BuildTime)
}
