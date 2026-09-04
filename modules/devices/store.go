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

// deviceStore is the backend of the demo: an orm.DB over storage/mem,
// speaking view's domain seam directly. No operation names, no envelopes.
type deviceStore struct{ db *orm.DB }

func (s *deviceStore) List() ([]model.Model, error) {
	var rows []*Device
	err := s.db.Query(&Device{}).ReadAll(
		func() model.Model { return &Device{} },
		func(m model.Model) { rows = append(rows, m.(*Device)) },
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

func (s *deviceStore) Save(recs ...model.Model) error {
	if len(recs) == 0 {
		return Errf("deviceStore: save: empty records")
	}
	for _, m := range recs {
		d := m.(*Device)
		if findErr := s.db.Query(&Device{}).Where("id").Eq(d.Id).ReadOne(); findErr != nil {
			if err := s.db.Create(d); err != nil { // no existing row with this id — new record
				return err
			}
		} else {
			if err := s.db.Update(d, storage.Eq("id", d.Id)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *deviceStore) Update(ids []string, rec model.Model, fields []string) error {
	switch {
	case len(ids) == 0:
		return Errf("deviceStore: update: empty ids")
	case len(fields) == 0:
		return Errf("deviceStore: update: empty fields")
	case model.IsNil(rec):
		return Errf("deviceStore: update: missing record")
	default:
		return s.db.UpdateFields(rec, fields, storage.In("id", anyIDs(ids)))
	}
}

func (s *deviceStore) Delete(ids ...string) error {
	if len(ids) == 0 {
		return Errf("deviceStore: delete: empty ids")
	}
	// One statement for the whole batch: atomic by construction, no loop
	// that could leave a half-applied delete behind.
	return s.db.Delete(&Device{}, storage.In("id", anyIDs(ids)))
}

func anyIDs(ids []string) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}
