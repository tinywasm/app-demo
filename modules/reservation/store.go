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

type memCaller struct{ db *orm.DB }

func (c *memCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	var err error
	switch op {
	case "reservation_list":
		if rl, ok := into.(*reservationList); ok {
			var rows []*Reservation
			err = c.db.Query(&Reservation{}).ReadAll(
				func() model.Model { return &Reservation{} },
				func(m model.Model) { rows = append(rows, m.(*Reservation)) },
			)
			for i := len(rows) - 1; i >= 0 && err == nil; i-- {
				dst := rl.Append().(*Reservation)
				*dst = *rows[i]
			}
		}
	case "reservation_save":
		recs := readArgs(args).records
		if len(recs) == 0 {
			err = Errf("memCaller: reservation_save: empty records")
			break
		}
		for _, r := range recs {
			if err != nil {
				break
			}
			if findErr := c.db.Query(&Reservation{}).Where("id").Eq(r.Id).ReadOne(); findErr != nil {
				err = c.db.Create(r)
			} else {
				err = c.db.Update(r, storage.Eq("id", r.Id))
			}
		}
	case "reservation_update":
		a := readArgs(args)
		switch {
		case len(a.ids) == 0:
			err = Errf("memCaller: reservation_update: empty ids")
		case len(a.fields) == 0:
			err = Errf("memCaller: reservation_update: empty fields")
		case a.record == nil:
			err = Errf("memCaller: reservation_update: missing record")
		default:
			anyIDs := make([]any, len(a.ids))
			for i, id := range a.ids {
				anyIDs[i] = id
			}
			err = c.db.UpdateFields(a.record, a.fields, storage.In("id", anyIDs))
		}
	case "reservation_delete":
		ids := readArgs(args).ids
		if len(ids) == 0 {
			err = Errf("memCaller: reservation_delete: empty ids")
			break
		}
		anyIDs := make([]any, len(ids))
		for i, id := range ids {
			anyIDs[i] = id
		}
		err = c.db.Delete(&Reservation{}, storage.In("id", anyIDs))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}

func readArgs(args model.Encodable) *reservationArgs {
	w := &reservationArgs{}
	args.EncodeFields(w)
	return w
}

type reservationArgs struct {
	ids     []string
	fields  []string
	records []*Reservation
	record  *Reservation
}

func (*reservationArgs) String(string, string) {}
func (*reservationArgs) Int(string, int64)     {}
func (*reservationArgs) Float(string, float64) {}
func (*reservationArgs) Bool(string, bool)     {}
func (*reservationArgs) Bytes(string, []byte)  {}
func (*reservationArgs) Null(string)           {}
func (*reservationArgs) Raw(string, string)    {}

func (w *reservationArgs) Object(name string, val model.Encodable) {
	if name == "record" {
		if r, ok := val.(*Reservation); ok {
			w.record = r
		}
	}
}

func (w *reservationArgs) Array(name string, _ int) model.ArrayWriter {
	switch name {
	case "ids":
		return (*stringSink)(&w.ids)
	case "fields":
		return (*stringSink)(&w.fields)
	case "records":
		return &reservationSink{w}
	}
	return discardSink{}
}

type stringSink []string

func (s *stringSink) String(v string)       { *s = append(*s, v) }
func (*stringSink) Int(int64)               {}
func (*stringSink) Float(float64)           {}
func (*stringSink) Bool(bool)               {}
func (*stringSink) Bytes([]byte)            {}
func (*stringSink) Object(model.Encodable)  {}
func (*stringSink) Close()                  {}

type reservationSink struct{ w *reservationArgs }

func (reservationSink) String(string) {}
func (reservationSink) Int(int64)     {}
func (reservationSink) Float(float64) {}
func (reservationSink) Bool(bool)     {}
func (reservationSink) Bytes([]byte)  {}
func (s reservationSink) Object(val model.Encodable) {
	if r, ok := val.(*Reservation); ok {
		s.w.records = append(s.w.records, r)
	}
}
func (reservationSink) Close() {}

type discardSink struct{}

func (discardSink) String(string)          {}
func (discardSink) Int(int64)              {}
func (discardSink) Float(float64)          {}
func (discardSink) Bool(bool)              {}
func (discardSink) Bytes([]byte)           {}
func (discardSink) Object(model.Encodable) {}
func (discardSink) Close()                 {}
