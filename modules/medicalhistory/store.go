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
		// Plural contract, mirrors devices' device_save: view ships
		// saveArgs{records}, N=1 for the form's single save. The wire is read
		// through its encoding (see readArgs), never asserted as *Visit.
		recs := readArgs(args).records
		if len(recs) == 0 {
			err = Errf("memCaller: visit_save: empty records")
			break
		}
		for _, v := range recs {
			if err != nil {
				break
			}
			if findErr := c.db.Query(&Visit{}).Where("id").Eq(v.Id).ReadOne(); findErr != nil {
				err = c.db.Create(v) // no existing row with this id — new record
			} else {
				err = c.db.Update(v, storage.Eq("id", v.Id))
			}
		}
	case "visit_update":
		// Bulk field patch, mirrors devices' device_update: only the named
		// columns, across every id, one statement. N=1 arrives in the same
		// shape as N=100.
		a := readArgs(args)
		switch {
		case len(a.ids) == 0:
			err = Errf("memCaller: visit_update: empty ids")
		case len(a.fields) == 0:
			err = Errf("memCaller: visit_update: empty fields")
		case a.record == nil:
			err = Errf("memCaller: visit_update: missing record")
		default:
			anyIDs := make([]any, len(a.ids))
			for i, id := range a.ids {
				anyIDs[i] = id
			}
			err = c.db.UpdateFields(a.record, a.fields, storage.In("id", anyIDs))
		}
	case "visit_delete":
		ids := readArgs(args).ids
		if len(ids) == 0 {
			err = Errf("memCaller: visit_delete: empty ids")
			break
		}
		anyIDs := make([]any, len(ids))
		for i, id := range ids {
			anyIDs[i] = id
		}
		err = c.db.Delete(&Visit{}, storage.In("id", anyIDs))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}

// readArgs walks a view payload's encoding. Mirrors devices' own — the demo
// tier keeps one memCaller per module instead of sharing, so the walk lives
// beside each use. See devices/store.go for why the wire is read, not asserted.
func readArgs(args model.Encodable) *visitArgs {
	w := &visitArgs{}
	args.EncodeFields(w)
	return w
}

type visitArgs struct {
	ids     []string // update, delete
	fields  []string // update: which columns to write
	records []*Visit // save: N whole records
	record  *Visit   // update: the values carrier
}

func (*visitArgs) String(string, string) {}
func (*visitArgs) Int(string, int64)     {}
func (*visitArgs) Float(string, float64) {}
func (*visitArgs) Bool(string, bool)     {}
func (*visitArgs) Bytes(string, []byte)  {}
func (*visitArgs) Null(string)           {}
func (*visitArgs) Raw(string, string)    {}

func (w *visitArgs) Object(name string, val model.Encodable) {
	if name == "record" {
		if v, ok := val.(*Visit); ok {
			w.record = v
		}
	}
}

func (w *visitArgs) Array(name string, _ int) model.ArrayWriter {
	switch name {
	case "ids":
		return (*stringSink)(&w.ids)
	case "fields":
		return (*stringSink)(&w.fields)
	case "records":
		return &visitSink{w}
	}
	return discardSink{}
}

type stringSink []string

func (s *stringSink) String(v string)      { *s = append(*s, v) }
func (*stringSink) Int(int64)              {}
func (*stringSink) Float(float64)          {}
func (*stringSink) Bool(bool)              {}
func (*stringSink) Bytes([]byte)           {}
func (*stringSink) Object(model.Encodable) {}
func (*stringSink) Close()                 {}

type visitSink struct{ w *visitArgs }

func (visitSink) String(string) {}
func (visitSink) Int(int64)     {}
func (visitSink) Float(float64) {}
func (visitSink) Bool(bool)     {}
func (visitSink) Bytes([]byte)  {}
func (s visitSink) Object(val model.Encodable) {
	if v, ok := val.(*Visit); ok {
		s.w.records = append(s.w.records, v)
	}
}
func (visitSink) Close() {}

type discardSink struct{}

func (discardSink) String(string)          {}
func (discardSink) Int(int64)              {}
func (discardSink) Bool(bool)              {}
func (discardSink) Float(float64)          {}
func (discardSink) Bytes([]byte)           {}
func (discardSink) Object(model.Encodable) {}
func (discardSink) Close()                 {}
