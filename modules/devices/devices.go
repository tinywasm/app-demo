// Package devices es el módulo demo de un CRUD completo: modelo real, backend en
// memoria y crudview montado sobre él. Es el patrón a copiar para un módulo que
// administra registros.
package devices

import (
	"github.com/tinywasm/components/searchbar"
	"github.com/tinywasm/layout/crudview"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/model"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/unixid"
	"github.com/tinywasm/view"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

// Icon es la referencia compartida al glifo del módulo. El dibujo vive en svg.go,
// que es backend puro. El prefijo mod- evita chocar con los ids del sprite del
// chasis dentro del sprite de página, que es la fusión de todos ellos.
const Icon = svg.Icon("mod-devices")

// Module es el módulo de equipos. Implementa platformd.UIModule.
type Module struct {
	p *platformd.Platform
}

// New construye el módulo. Recibe la plataforma porque este módulo notifica el
// resultado de sus mutaciones en la barra de mensajes del chasis.
func New(p *platformd.Platform) *Module { return &Module{p: p} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "devices" }
func (m *Module) Label() string     { return "Computadores" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	pres := view.New(&memCaller{db: deviceDB}, &Device{}, "device_list",
		func() model.ModelSlice { return &deviceList{} },
		view.WithTitle("Computadores"),
		view.WithSaveOp("device_save"),
		view.WithDeleteOp("device_delete"),
	)
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	cv, err := crudview.New(crudview.Config{
		ParentID:  m.ModelName(),
		Presenter: pres,
		IDs:       ids,
		// El filtro lo elige la aplicación, no el layout. Esta es la línea que un
		// despliegue cambia por un calendario o un select de categoría; ni crudview
		// ni rightpanel se enteran.
		Filter: &searchbar.SearchBar{Placeholder: "Buscar..."},
	})
	if err != nil {
		panic(err)
	}
	// Sin notificación al seleccionar: elegir una fila es navegar, no actuar — un
	// aviso ahí no aporta nada y en móvil, donde el módulo llena la pantalla, el
	// mensaje se superpone a la vista entera. Guardar y borrar SÍ notifican, porque
	// confirman una mutación real.
	cv.OnNew = func() { m.p.Notify(Msg.Info, "Nuevo", platformd.Auto()) }
	cv.OnSaved = func(err error) {
		if err == nil {
			m.p.Notify(Msg.Success, "Guardado", platformd.Auto())
		}
	}
	cv.OnDeleted = func(id string, err error) {
		if err == nil {
			// Éxito, no Error: la mutación se aplicó. El tipo informa de la
			// severidad, no del verbo.
			m.p.Notify(Msg.Success, "Eliminado "+id, platformd.Auto())
		}
	}
	cv.OnCancel = func() { m.p.Notify(Msg.Info, "Cancelado", platformd.Auto()) }
	return cv
}
