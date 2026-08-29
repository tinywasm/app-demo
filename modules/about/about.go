// Package about es el módulo demo de una pantalla de información estática.
package about

import (
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/svg"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

const Icon = svg.Icon("mod-about")

type Module struct{}

func New() *Module { return &Module{} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "about" }
func (m *Module) Label() string     { return "Acerca de" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	return &rightpanel.RightPanel{
		Title:   m.Label(),
		Article: Div().Text("Contenido de " + m.Label()),
	}
}
