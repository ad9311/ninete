package serve

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

const (
	assetManifestPath = "./web/static/js/build/manifest.json"
	staticBuildPath   = "/static/js/build/"
)

// assetManifest maps a web/build.ts entry name to the content-hashed filename
// it produced, so the templates never hardcode a bundle name that a deploy can
// leave stale under /static/*'s cache window (docs/spa-migration.md §3.8).
type assetManifest map[string]string

// LoadAssetManifest reads the manifest web/build.ts writes beside the bundles.
// It runs at startup, like LoadTemplates, and again from the development
// template-reload hook (serve.go) because a rebuild changes the hashed
// filenames.
//
// Call it after LoadTemplates: the missing-file branch reads s.templates to
// tell the two cases apart.
//
// A missing file is not an error, the same way parseTemplates (template.go)
// treats an empty views glob: internal/logic and internal/repo build
// *Server through spec.New too, from a working directory with no web/ tree at
// all, and never render a page or ask for a bundle path. Only a manifest that
// exists but fails to parse — the shape a real build-static-js failure would
// take at the repo root every other package runs from — is fatal.
func (s *Server) LoadAssetManifest() error {
	data, err := os.ReadFile(assetManifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Logged rather than swallowed whenever templates did load, which
			// is what separates "this process has no web/ tree at all" from
			// "the bundles were never built". In the second case this line is
			// the only signal that every rendered page is about to ship with an
			// empty <script src> and no JS behind it.
			if len(s.templates) > 0 {
				s.app.Logger.Errorf(
					"asset manifest not found at %s, run make build-static-js",
					assetManifestPath,
				)
			}

			return nil
		}

		return err
	}

	var manifest assetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	s.assetManifest = manifest

	return nil
}

// bundlePath resolves a manifest entry to the public /static/* path. An
// unknown entry returns "", which surfaces as a broken <script src> in the
// rendered page rather than a panic — the same failure shape as a missing
// TemplateName.
func (s *Server) bundlePath(entry string) string {
	name, ok := s.assetManifest[entry]
	if !ok {
		return ""
	}

	return staticBuildPath + name
}
