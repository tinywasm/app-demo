package medicalhistory

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/storage/mem"

	. "github.com/tinywasm/fmt"
)

// Patient is a LOCAL FAKE type for this demo only — today's agenda. A real
// deployment replaces todayAgenda with a backend call into
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

// visitStore is the backend of the demo: an orm.DB over storage/mem,
// speaking view's domain seam directly. No operation names, no envelopes.
type visitStore struct{ db *orm.DB }

func (s *visitStore) List() ([]model.Model, error) {
	var rows []*Visit
	err := s.db.Query(&Visit{}).ReadAll(
		func() model.Model { return &Visit{} },
		func(m model.Model) { rows = append(rows, m.(*Visit)) },
	)
	if err != nil {
		return nil, err
	}
	// Newest first: mem has no timestamp/sequence column, so this just
	// reverses creation order — the same order Create() appended in.
	out := make([]model.Model, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		out = append(out, rows[i])
	}
	return out, nil
}

func (s *visitStore) Save(recs []model.Model) error {
	if len(recs) == 0 {
		return Errf("visitStore: save: empty records")
	}
	for _, m := range recs {
		v := m.(*Visit)
		if findErr := s.db.Query(&Visit{}).Where("id").Eq(v.Id).ReadOne(); findErr != nil {
			if err := s.db.Create(v); err != nil { // no existing row with this id — new record
				return err
			}
		} else {
			if err := s.db.Update(v, storage.Eq("id", v.Id)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *visitStore) Update(ids []string, rec model.Model, fields []string) error {
	switch {
	case len(ids) == 0:
		return Errf("visitStore: update: empty ids")
	case len(fields) == 0:
		return Errf("visitStore: update: empty fields")
	case model.IsNil(rec):
		return Errf("visitStore: update: missing record")
	default:
		return s.db.UpdateFields(rec, fields, storage.In("id", anyIDs(ids)))
	}
}

func (s *visitStore) Delete(ids []string) error {
	if len(ids) == 0 {
		return Errf("visitStore: delete: empty ids")
	}
	// One statement for the whole batch: atomic by construction, no loop
	// that could leave a half-applied delete behind.
	return s.db.Delete(&Visit{}, storage.In("id", anyIDs(ids)))
}

func anyIDs(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}
