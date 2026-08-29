// Package medicalhistory is a demo module: the same crudview.New pattern.
// platformd/modules/devices uses for a full CRUD view, with one addition —
// a selectsearch.SelectSearch fills crudview.Config.Filter instead of the
// default searchbar.SearchBar. SelectSearch satisfies widget.Filterable
// (github.com/tinywasm/components v0.5.0) the same generic contract a
// SearchBar fills, so picking a patient from today's agenda narrows the
// visit list to that patient's fichas — no bespoke wiring in crudview or
// rightpanel.
package medicalhistory

import (
	"github.com/tinywasm/components/selectsearch"
	"github.com/tinywasm/components/targethour"
	"github.com/tinywasm/layout/crudview"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/model"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/unixid"
	"github.com/tinywasm/view"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

// requirePatient wraps the generic Presenter view.New builds so Filter(term)
// means "which patient", not a free-text search: this module has nothing
// worth browsing until a patient is picked (no term → no fichas, module
// logic, not something crudview or view should decide generically), and a
// picked patient's fichas are found by an exact match against visitDB
// directly — core.Filter's own Label+Description substring search can't do
// it anymore now that Description carries the attending doctor, not the
// patient (see Visit.Item).
type requirePatient struct {
	view.Presenter
}

func (p requirePatient) Filter(term string) []view.Item {
	if term == "" {
		return nil
	}
	var rows []*Visit
	_ = visitDB.Query(&Visit{}).Where("patient").Eq(term).ReadAll(
		func() model.Model { return &Visit{} },
		func(m model.Model) { rows = append(rows, m.(*Visit)) },
	)
	items := make([]view.Item, len(rows))
	for i, v := range rows {
		// Newest first, same convention memCaller's visit_list uses.
		items[len(rows)-1-i] = v.Item()
	}
	return items
}

// Save/Delete forward to the embedded Presenter explicitly: embedding an
// INTERFACE only promotes the methods view.Presenter itself declares, never
// extra ones a concrete value happens to also satisfy — view.New's return
// value here always implements both (Config always sets WithSaveOp and
// WithDeleteOp below), so the fallback errors are unreachable in practice,
// not a real degraded mode.
func (p requirePatient) Save(payload model.Model) error {
	if s, ok := p.Presenter.(view.Saver); ok {
		return s.Save(payload)
	}
	return Errf("requirePatient: underlying presenter cannot save")
}

func (p requirePatient) Delete(id string) error {
	if d, ok := p.Presenter.(view.Deleter); ok {
		return d.Delete(id)
	}
	return Errf("requirePatient: underlying presenter cannot delete")
}

var _ view.Presenter = requirePatient{}
var _ view.Saver = requirePatient{}
var _ view.Deleter = requirePatient{}

const Icon = svg.Icon("mod-medicalhistory")

// Module is the platformd.UIModule. It carries the platform because, like
// devices, it notifies save/delete results in the chassis message bar.
type Module struct {
	p *platformd.Platform
}

func New(p *platformd.Platform) *Module { return &Module{p: p} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "medicalhistory" }
func (m *Module) Label() string     { return "Historial Médico" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	pres := requirePatient{view.New(&memCaller{db: visitDB}, &Visit{}, "visit_list",
		func() model.ModelSlice { return &visitList{} },
		view.WithTitle("Ficha Paciente"),
		view.WithSaveOp("visit_save"),
		view.WithDeleteOp("visit_delete"),
	)}

	// The picker's ID IS the patient's name, not todayAgenda's opaque "p1"
	// key: OnFilterChange's term reaches requirePatient.Filter(term) above,
	// which matches it against Visit.Patient directly — so the term has to
	// be that same name.
	picker := &selectsearch.SelectSearch{Placeholder: "Seleccione un paciente..."}
	opts := make([]selectsearch.SsOption, len(todayAgenda))
	for i, p := range todayAgenda {
		opts[i] = selectsearch.SsOption{
			ID:          p.Name,
			Label:       p.Name + ", " + p.Age + " años",
			Sublabel:    p.Run,
			Description: p.Time,
		}
	}
	picker.Options = opts

	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	cv, err := crudview.New(crudview.Config{
		ParentID:  m.ModelName(),
		Presenter: pres,
		IDs:       ids,
		Filter:    picker,
		// The fichas list leads with the visit's date (see Visit.Item's
		// LeadTop/Main/Bottom) instead of a plain label — targethour reads
		// that badge, targetlist (crudview's default) would just ignore it.
		List: func(selected *SignalString, onSelect func(view.Item), onDelete func(string)) crudview.ListView {
			return &targethour.TargetHour{Selected: selected, OnSelect: onSelect, OnDelete: onDelete}
		},
	})
	if err != nil {
		panic(err)
	}
	cv.OnNew = func() { m.p.Notify(Msg.Info, "Nueva ficha", platformd.Auto()) }
	cv.OnSaved = func(err error) {
		if err == nil {
			m.p.Notify(Msg.Success, "Guardado", platformd.Auto())
		}
	}
	cv.OnDeleted = func(id string, err error) {
		if err == nil {
			m.p.Notify(Msg.Success, "Eliminado "+id, platformd.Auto())
		}
	}
	cv.OnCancel = func() { m.p.Notify(Msg.Info, "Cancelado", platformd.Auto()) }
	return cv
}
