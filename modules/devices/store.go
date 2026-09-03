package devices

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/storage/mem"

	. "github.com/tinywasm/fmt"
)

// deviceDB is a real (in-memory) CRUD backend for the demo, so the UI's actual
// behavior — save-on-blur, the new item landing in the list, delete removing
// it — can be exercised for real instead of against a static 3-item fixture.
// Package-level and built once: it must survive across View() calls (switching
// tabs and back) so the demo doesn't lose its data every time the module is
// re-rendered. It DOES reset on a full page reload/process restart — that is
// the expected trade-off of an in-memory store, not a bug.
var deviceDB = newSeededDeviceDB()

func newSeededDeviceDB() *orm.DB {
	db := orm.New(mem.New())
	for _, d := range []*Device{
		{Id: "10", Name: "Pc Administracion", Ip: "192.168.122.10"},
		{Id: "11", Name: "Pc Ventas", Ip: "192.168.122.11"},
		{Id: "12", Name: "Servidor Web", Ip: "192.168.122.20"},
		{Id: "13", Name: "Pc Soporte", Ip: "192.168.122.13"},
		{Id: "14", Name: "Pc Bodega", Ip: "192.168.122.14"},
		{Id: "15", Name: "Servidor Backup", Ip: "192.168.122.21"},
		{Id: "16", Name: "Pc Recepcion", Ip: "192.168.122.16"},
		{Id: "17", Name: "Pc Gerencia", Ip: "192.168.122.17"},
		{Id: "18", Name: "Servidor Impresion", Ip: "192.168.122.22"},
		{Id: "19", Name: "Pc Contabilidad", Ip: "192.168.122.19"},
		{Id: "20", Name: "Pc Taller", Ip: "192.168.122.30"},
		{Id: "21", Name: "Servidor Archivos", Ip: "192.168.122.23"},
		{Id: "22", Name: "Pc Despacho", Ip: "192.168.122.31"},
		{Id: "23", Name: "Pc Calidad", Ip: "192.168.122.32"},
		{Id: "24", Name: "Servidor Monitoreo", Ip: "192.168.122.24"},
	} {
		_ = db.Create(d)
	}
	return db
}

// memCaller adapts deviceDB (an *orm.DB over storage/mem) to router.Caller —
// the seam view.Presenter drives. Device-specific rather than generic: this is
// app/demo wiring, not a shared library.
type memCaller struct{ db *orm.DB }

func (c *memCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	var err error
	switch op {
	case "device_list":
		if dl, ok := into.(*deviceList); ok {
			var rows []*Device
			err = c.db.Query(&Device{}).ReadAll(
				func() model.Model { return &Device{} },
				func(m model.Model) { rows = append(rows, m.(*Device)) },
			)
			// Newest first: mem has no timestamp/sequence column, so this just
			// reverses creation order — the same order Create() appended in.
			for i := len(rows) - 1; i >= 0 && err == nil; i-- {
				dst := dl.Append().(*Device)
				*dst = *rows[i]
			}
		}
	case "device_save":
		// Plural contract: view ships saveArgs{records}, N=1 for the form's
		// single save. deviceDB is unexported by design, so the wire is read
		// through its encoding (see readArgs) — never asserted as *Device.
		recs := readArgs(args).records
		if len(recs) == 0 {
			err = Errf("memCaller: device_save: empty records")
			break
		}
		for _, d := range recs {
			if err != nil {
				break
			}
			if findErr := c.db.Query(&Device{}).Where("id").Eq(d.Id).ReadOne(); findErr != nil {
				err = c.db.Create(d) // no existing row with this id — new record
			} else {
				err = c.db.Update(d, storage.Eq("id", d.Id))
			}
		}
	case "device_update":
		// Bulk field patch: only the named columns, across every id, one
		// statement — the "ids + delta" contract. N=1 (a single-record bulk
		// edit) arrives in the same shape as N=100.
		a := readArgs(args)
		switch {
		case len(a.ids) == 0:
			err = Errf("memCaller: device_update: empty ids")
		case len(a.fields) == 0:
			err = Errf("memCaller: device_update: empty fields")
		case a.record == nil:
			err = Errf("memCaller: device_update: missing record")
		default:
			anyIDs := make([]any, len(a.ids))
			for i, id := range a.ids {
				anyIDs[i] = id
			}
			err = c.db.UpdateFields(a.record, a.fields, storage.In("id", anyIDs))
		}
	case "device_delete":
		ids := readArgs(args).ids
		if len(ids) == 0 {
			err = Errf("memCaller: device_delete: empty ids")
			break
		}
		anyIDs := make([]any, len(ids))
		for i, id := range ids {
			anyIDs[i] = id
		}
		// One statement for the whole batch: atomic by construction, no loop
		// that could leave a half-applied delete behind.
		err = c.db.Delete(&Device{}, storage.In("id", anyIDs))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}

// readArgs walks a view payload's encoding. view wraps every write in an
// unexported struct — saveArgs{records}, updateArgs{ids,fields,record},
// deleteArgs{ids} — so the store cannot type-assert the payload; it reads the
// wire instead. The record objects ARE this module's own *Device, passed
// through by view untouched, so those are asserted (the store owns the type).
func readArgs(args model.Encodable) *deviceArgs {
	w := &deviceArgs{}
	args.EncodeFields(w)
	return w
}

type deviceArgs struct {
	ids     []string  // update, delete
	fields  []string  // update: which columns to write
	records []*Device // save: N whole records
	record  *Device   // update: the values carrier
}

func (*deviceArgs) String(string, string) {}
func (*deviceArgs) Int(string, int64)     {}
func (*deviceArgs) Float(string, float64) {}
func (*deviceArgs) Bool(string, bool)     {}
func (*deviceArgs) Bytes(string, []byte)  {}
func (*deviceArgs) Null(string)           {}
func (*deviceArgs) Raw(string, string)    {}

func (w *deviceArgs) Object(name string, val model.Encodable) {
	if name == "record" {
		if d, ok := val.(*Device); ok {
			w.record = d
		}
	}
}

func (w *deviceArgs) Array(name string, _ int) model.ArrayWriter {
	switch name {
	case "ids":
		return (*stringSink)(&w.ids)
	case "fields":
		return (*stringSink)(&w.fields)
	case "records":
		return &deviceSink{w}
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

type deviceSink struct{ w *deviceArgs }

func (deviceSink) String(string) {}
func (deviceSink) Int(int64)     {}
func (deviceSink) Float(float64) {}
func (deviceSink) Bool(bool)     {}
func (deviceSink) Bytes([]byte)  {}
func (s deviceSink) Object(val model.Encodable) {
	if d, ok := val.(*Device); ok {
		s.w.records = append(s.w.records, d)
	}
}
func (deviceSink) Close() {}

type discardSink struct{}

func (discardSink) String(string)          {}
func (discardSink) Int(int64)              {}
func (discardSink) Float(float64)          {}
func (discardSink) Bool(bool)              {}
func (discardSink) Bytes([]byte)           {}
func (discardSink) Object(model.Encodable) {}
func (discardSink) Close()                 {}
