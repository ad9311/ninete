package contracttest

import "errors"

// ErrNoSource means the runtime could not report this file's path, which Load
// needs to find the repository root.
var ErrNoSource = errors.New("contracttest: cannot locate source file")
