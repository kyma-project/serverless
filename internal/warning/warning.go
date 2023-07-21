package warning

import (
	"fmt"
	"strings"
)

type Builder struct {
	warnings []string
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (w *Builder) With(warning string) *Builder {
	w.warnings = append(w.warnings, warning)
	return w
}

func (w *Builder) Build() string {
	msg := ""
	if len(w.warnings) > 0 {
		msg = fmt.Sprintf("Warning: %s", strings.Join(w.warnings, "; "))
	}
	return msg
}
