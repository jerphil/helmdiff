package renderer

import "github.com/jerphil/helmdiff/internal/diff"

type Renderer interface {
	Render(report *diff.DiffReport) error
}
