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
// the seam view.Presenter drives. Device-specific (type-asserts *Device)
// rather than generic: this is app/demo wiring, not a shared library.
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
		d, ok := args.(*Device)
		if !ok {
			err = Errf("memCaller: device_save: unexpected payload type")
			break
		}
		if findErr := c.db.Query(&Device{}).Where("id").Eq(d.Id).ReadOne(); findErr != nil {
			err = c.db.Create(d) // no existing row with this id — new record
		} else {
			err = c.db.Update(d, storage.Eq("id", d.Id))
		}
	case "device_delete":
		d, ok := args.(*Device)
		if !ok {
			err = Errf("memCaller: device_delete: unexpected payload type")
			break
		}
		err = c.db.Delete(d, storage.Eq("id", d.Id))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}
