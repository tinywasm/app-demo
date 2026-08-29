package medicalhistory

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/storage/mem"

	. "github.com/tinywasm/fmt"
)

// Patient is a LOCAL FAKE type for this demo only — today's agenda. A real
// deployment replaces todayAgenda with a router.Caller call into
// appointment_booking (today's reservations for a staff member); this repo
// never imports it. Every patient/visit record here is fake, package-level,
// in-memory data, same tier as devices' deviceDB below.
type Patient struct {
	ID   string
	Name string
	Time string
	Age  string
	Run  string
}

var todayAgenda = []Patient{
	{ID: "p1", Name: "Juan Pérez", Time: "09:00", Age: "34", Run: "12.345.678-9"},
	{ID: "p2", Name: "María Soto", Time: "09:30", Age: "27", Run: "15.987.654-3"},
	{ID: "p3", Name: "Diego Rojas", Time: "10:15", Age: "58", Run: "9.876.543-2"},
	{ID: "p4", Name: "Camila Vidal", Time: "11:00", Age: "41", Run: "11.222.333-4"},
}

// visitDB is a real (in-memory) CRUD backend for the demo — see devices'
// deviceDB for why: package-level so it survives across View() calls, and
// resets on a full reload, which is the expected trade-off.
var visitDB = newSeededVisitDB()

func newSeededVisitDB() *orm.DB {
	db := orm.New(mem.New())
	for _, v := range []*Visit{
		{Id: "v1", Patient: "Juan Pérez", Doctor: "dr. Tony Stark", Date: "2026-07-20", Reason: "Control", Diagnosis: "Sin hallazgos"},
		{Id: "v2", Patient: "Juan Pérez", Doctor: "dra. Natasha Romanoff", Date: "2026-03-11", Reason: "Dolor abdominal", Diagnosis: "Gastritis"},
		{Id: "v3", Patient: "María Soto", Doctor: "dra. Natasha Romanoff", Date: "2026-06-02", Reason: "Chequeo anual", Diagnosis: "Saludable"},
		{Id: "v4", Patient: "Diego Rojas", Doctor: "dr. Tony Stark", Date: "2026-01-15", Reason: "Fractura", Diagnosis: "Fractura de radio"},
	} {
		_ = db.Create(v)
	}
	return db
}

// memCaller adapts visitDB to router.Caller — the seam view.Presenter
// drives. Mirrors devices' memCaller exactly.
type memCaller struct{ db *orm.DB }

func (c *memCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	var err error
	switch op {
	case "visit_list":
		if vl, ok := into.(*visitList); ok {
			var rows []*Visit
			err = c.db.Query(&Visit{}).ReadAll(
				func() model.Model { return &Visit{} },
				func(m model.Model) { rows = append(rows, m.(*Visit)) },
			)
			for i := len(rows) - 1; i >= 0 && err == nil; i-- {
				dst := vl.Append().(*Visit)
				*dst = *rows[i]
			}
		}
	case "visit_save":
		v, ok := args.(*Visit)
		if !ok {
			err = Errf("memCaller: visit_save: unexpected payload type")
			break
		}
		if findErr := c.db.Query(&Visit{}).Where("id").Eq(v.Id).ReadOne(); findErr != nil {
			err = c.db.Create(v) // no existing row with this id — new record
		} else {
			err = c.db.Update(v, storage.Eq("id", v.Id))
		}
	case "visit_delete":
		v, ok := args.(*Visit)
		if !ok {
			err = Errf("memCaller: visit_delete: unexpected payload type")
			break
		}
		err = c.db.Delete(v, storage.Eq("id", v.Id))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}
