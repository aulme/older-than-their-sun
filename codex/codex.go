// Package codex holds the hand-written fiction of the galaxy for whoever
// narrates it: one markdown file per system, an index, the rendering
// grammar of the tellings and the principles of what is knowable. It is
// embedded so that a run directory carries its own copy. The files are
// written at the codex step (specs/plan.md step 8); until then the
// index alone stands here.
package codex

import "embed"

// FS is every file of the codex.
//
//go:embed *.md
var FS embed.FS
