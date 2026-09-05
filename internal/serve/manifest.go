package serve

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

const assetManifestPath = "./web/static/manifest.json"

// assetManifest maps a web/build.ts entry name to the public /static/* path of
// the content-hashed file it produced, so the templates never hardcode a name
// that a deploy can leave stale under /static/*'s cache window
// (docs/spa-migration.md §3.8).
//
// The value is the whole path, not just the filename: the JS bundle and the
// stylesheet build into different directories, and web/build.ts is the only
// place that decides which. Reconstructing it here would be a second copy of
// that layout, in the language that cannot see the build script.
type assetManifest map[string]string

// LoadAssetManifest reads the manifest web/build.ts writes beside the build output.
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
// exists but fails to parse — the shape a real build-static failure would
// take at the repo root every other package runs from — is fatal.
func (s *Server) LoadAssetManifest() error {
	data, err := os.ReadFile(assetManifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Logged rather than swallowed whenever templates did load, which
			// is what separates "this process has no web/ tree at all" from
			// "the assets were never built". In the second case this line is
			// the only signal that every rendered page is about to ship with an
			// empty <script src> and an empty <link href> behind it.
			if len(s.templates) > 0 {
				s.app.Logger.Errorf(
					"asset manifest not found at %s, run make build-static",
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

// assetPath resolves a manifest entry to the public /static/* path web/build.ts
// recorded for it. An unknown entry returns "", which surfaces as a broken
// <script src> or <link href> in the rendered page rather than a panic — the
// same failure shape as a missing TemplateName.
func (s *Server) assetPath(entry string) string {
	return s.assetManifest[entry]
}
