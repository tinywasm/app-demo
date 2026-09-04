package reservation

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/storage/mem"

	. "github.com/tinywasm/fmt"
)

var reservationDB = newSeededReservationDB()

func newSeededReservationDB() *orm.DB {
	db := orm.New(mem.New())
	for _, r := range []*Reservation{
		{Id: "100", PatientRun: "12345678-9", PatientName: "María Gonzalez", PatientBirthday: "1985-04-12", PatientContact: "+56912345678", Day: "2026-09-10", Hour: "09:00", Detail: "Consulta general", Status: "confirmed"},
		{Id: "101", PatientRun: "98765432-1", PatientName: "Juan Pérez", PatientBirthday: "1990-08-23", PatientContact: "juan@example.com", Day: "2026-09-10", Hour: "10:30", Detail: "Revisión exámenes", Status: "attended"},
		{Id: "102", PatientRun: "11223344-5", PatientName: "Ana Silva", PatientBirthday: "1978-11-05", PatientContact: "+56987654321", Day: "2026-09-10", Hour: "11:15", Detail: "Control rutinario", Status: ""},
		{Id: "103", PatientRun: "15556677-8", PatientName: "Carlos Tapia", PatientBirthday: "2001-02-17", PatientContact: "+56911223344", Day: "2026-09-10", Hour: "14:00", Detail: "Dolor lumbar", Status: "confirmed"},
		{Id: "104", PatientRun: "16778899-0", PatientName: "Lucía Morales", PatientBirthday: "1995-06-30", PatientContact: "lucia@example.com", Day: "2026-09-11", Hour: "09:30", Detail: "Especialista cardiología", Status: "confirmed"},
		{Id: "105", PatientRun: "17889900-1", PatientName: "Pedro Soto", PatientBirthday: "1982-12-14", PatientContact: "+56955443322", Day: "2026-09-11", Hour: "11:00", Detail: "Seguimiento tratamiento", Status: ""},
		{Id: "106", PatientRun: "18990011-2", PatientName: "Sofia Rojas", PatientBirthday: "1998-03-22", PatientContact: "+56966778899", Day: "2026-09-11", Hour: "15:30", Detail: "Consulta inicial", Status: "attended"},
		{Id: "107", PatientRun: "19001122-3", PatientName: "Diego Castro", PatientBirthday: "1975-09-08", PatientContact: "diego@example.com", Day: "2026-09-12", Hour: "10:00", Detail: "Chequeo anual", Status: "confirmed"},
	} {
		_ = db.Create(r)
	}
	return db
}

type reservationStore struct{ db *orm.DB }

func (s *reservationStore) List() ([]model.Model, error) {
	var rows []*Reservation
	err := s.db.Query(&Reservation{}).ReadAll(
		func() model.Model { return &Reservation{} },
		func(m model.Model) { rows = append(rows, m.(*Reservation)) },
	)
	if err != nil {
		return nil, err
	}
	out := make([]model.Model, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		out = append(out, rows[i])
	}
	return out, nil
}

func (s *reservationStore) Save(recs ...model.Model) error {
	if len(recs) == 0 {
		return Errf("reservationStore: save: empty records")
	}
	for _, m := range recs {
		r := m.(*Reservation)
		if findErr := s.db.Query(&Reservation{}).Where("id").Eq(r.Id).ReadOne(); findErr != nil {
			if err := s.db.Create(r); err != nil {
				return err
			}
		} else {
			if err := s.db.Update(r, storage.Eq("id", r.Id)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *reservationStore) Update(ids []string, rec model.Model, fields []string) error {
	switch {
	case len(ids) == 0:
		return Errf("reservationStore: update: empty ids")
	case len(fields) == 0:
		return Errf("reservationStore: update: empty fields")
	case model.IsNil(rec):
		return Errf("reservationStore: update: missing record")
	default:
		return s.db.UpdateFields(rec, fields, storage.In("id", anyIDs(ids)))
	}
}

func (s *reservationStore) Delete(ids ...string) error {
	if len(ids) == 0 {
		return Errf("reservationStore: delete: empty ids")
	}
	return s.db.Delete(&Reservation{}, storage.In("id", anyIDs(ids)))
}

func anyIDs(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}
