package serve

import (
	"encoding/json"
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
// It runs once at startup, like LoadTemplates: every command that starts this
// server rebuilds the JS first (make dev, make test), so the file is always
// fresh when this is called.
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
		if os.IsNotExist(err) {
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
