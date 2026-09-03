package reservation

import (
	"github.com/tinywasm/components/calendarslider"
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

type byDay struct {
	view.Presenter
}

func (p byDay) Filter(term string) []view.Item {
	if term == "" {
		return nil
	}
	var rows []*Reservation
	_ = reservationDB.Query(&Reservation{}).Where("day").Eq(term).ReadAll(
		func() model.Model { return &Reservation{} },
		func(m model.Model) { rows = append(rows, m.(*Reservation)) },
	)
	items := make([]view.Item, len(rows))
	for i, r := range rows {
		items[i] = r.Item()
	}
	return items
}

func (p byDay) Save(recs ...model.Model) error {
	if s, ok := p.Presenter.(view.Saver); ok {
		return s.Save(recs...)
	}
	return Errf("byDay: underlying presenter cannot save")
}

func (p byDay) Update(ids []string, rec model.Model, fields []string) error {
	if u, ok := p.Presenter.(view.Updater); ok {
		return u.Update(ids, rec, fields)
	}
	return Errf("byDay: underlying presenter cannot update")
}

func (p byDay) Delete(ids ...string) error {
	if d, ok := p.Presenter.(view.Deleter); ok {
		return d.Delete(ids...)
	}
	return Errf("byDay: underlying presenter cannot delete")
}

var _ view.Presenter = byDay{}
var _ view.Saver = byDay{}
var _ view.Updater = byDay{}
var _ view.Deleter = byDay{}

const Icon = svg.Icon("mod-reservation")

type Module struct {
	p *platformd.Platform
}

func New(p *platformd.Platform) *Module { return &Module{p: p} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "reservation" }
func (m *Module) Label() string     { return "Reserva Hora" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	pres := byDay{view.New(&memCaller{db: reservationDB}, &Reservation{}, "reservation_list",
		func() model.ModelSlice { return &reservationList{} },
		view.WithTitle("Reserva Hora"),
		view.WithSaveOp("reservation_save"),
		view.WithUpdateOp("reservation_update"),
		view.WithDeleteOp("reservation_delete"),
	)}

	cal := &calendarslider.CalendarSlider{}

	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}

	cv, err := crudview.New(crudview.Config{
		ParentID:  m.ModelName(),
		Presenter: pres,
		IDs:       ids,
		Filter:    cal,
		List: func(selected *SignalString, onSelect func(view.Item)) crudview.ListView {
			return &targethour.TargetHour{
				Selected: selected,
				OnSelect: onSelect,
				StatusOf: func(it view.Item) targethour.Status {
					switch it.Description {
					case "Confirmada":
						return targethour.StatusConfirmed
					case "Atendida":
						return targethour.StatusAttended
					}
					return targethour.StatusPending
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	cv.OnNew = func() { m.p.Notify(Msg.Info, "Nueva reserva", platformd.Auto()) }
	cv.OnSaved = func(err error) {
		if err == nil {
			m.p.Notify(Msg.Success, "Guardado", platformd.Auto())
		}
	}
	cv.OnDeleted = func(ids []string, err error) {
		if err != nil || len(ids) == 0 {
			return
		}
		msg := "Eliminado " + ids[0]
		if len(ids) != 1 {
			msg = Sprintf("%d registros eliminados", len(ids))
		}
		m.p.Notify(Msg.Success, msg, platformd.Auto())
	}
	cv.OnUpdated = func(ids []string, err error) {
		if err != nil || len(ids) == 0 {
			return
		}
		msg := "Actualizado " + ids[0]
		if len(ids) != 1 {
			msg = Sprintf("%d registros actualizados", len(ids))
		}
		m.p.Notify(Msg.Success, msg, platformd.Auto())
	}
	return cv
}
