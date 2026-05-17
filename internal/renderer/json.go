package renderer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jerphil/helmdiff/internal/diff"
)

type JSONRenderer struct{}

func (j *JSONRenderer) Render(report *diff.DiffReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
