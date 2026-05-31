package output

import (
	"encoding/json"
	"io"

	"github.com/samaasi/kubectl-plan/internal/analysis"
)

type Renderer struct {
	format string
	writer io.Writer
	ascii  bool
}

func NewRenderer(format string, writer io.Writer, ascii bool) *Renderer {
	return &Renderer{
		format: format,
		writer: writer,
		ascii:  ascii,
	}
}

func (r *Renderer) Render(result *analysis.AnalysisResult) error {
	if r.format == "json" {
		encoder := json.NewEncoder(r.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	return r.RenderTerminal(result)
}
