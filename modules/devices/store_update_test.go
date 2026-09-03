package devices

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/view"
)

// The bulk-patch contract, through the real presenter: Update ships
// {ids, fields, record} and the store writes ONLY the named columns, across
// exactly those ids, in one statement. N=1 and N>1 arrive in the same shape.
func TestMemCallerBulkUpdatePatchesOnlyNamedFields(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&memCaller{db: db}, &Device{}, "device_list",
		func() model.ModelSlice { return &deviceList{} },
		view.WithUpdateOp("device_update"),
	)
	updater, ok := pres.(view.Updater)
	if !ok {
		t.Fatal("presenter must implement view.Updater")
	}
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// Only "name" is in the field set — Ip must survive on every matched row.
	if err := updater.Update([]string{"10", "11"}, &Device{Name: "PATCHED"}, []string{"name"}); err != nil {
		t.Fatalf("bulk update failed: %v", err)
	}

	got := readAll(t, db)
	for _, id := range []string{"10", "11"} {
		if got[id].Name != "PATCHED" {
			t.Errorf("id %s: name not patched, got %q", id, got[id].Name)
		}
		if got[id].Ip == "" || got[id].Ip == "PATCHED" {
			t.Errorf("id %s: Ip must be untouched, got %q", id, got[id].Ip)
		}
	}
	if got["12"].Name == "PATCHED" {
		t.Error("id 12 was not in the id set and must be unchanged")
	}

	// N=1 ships the same shape: a single-record bulk edit is a batch of one.
	if err := updater.Update([]string{"12"}, &Device{Ip: "10.0.0.9"}, []string{"ip"}); err != nil {
		t.Fatalf("single update failed: %v", err)
	}
	got = readAll(t, db)
	if got["12"].Ip != "10.0.0.9" {
		t.Errorf("id 12: ip not patched, got %q", got["12"].Ip)
	}
	if got["12"].Name == "" || got["12"].Name == "10.0.0.9" {
		t.Errorf("id 12: name must be untouched, got %q", got["12"].Name)
	}
}

// An empty field set is an error, not a silent no-op (the commit button is
// disabled until there is something to apply, so this is a belt).
func TestMemCallerBulkUpdateRejectsEmptyFields(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&memCaller{db: db}, &Device{}, "device_list",
		func() model.ModelSlice { return &deviceList{} },
		view.WithUpdateOp("device_update"),
	)
	updater := pres.(view.Updater)
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if err := updater.Update([]string{"10"}, &Device{Name: "X"}, nil); err == nil {
		t.Error("expected an error for an empty field set")
	}
}

func readAll(t *testing.T, db *orm.DB) map[string]*Device {
	t.Helper()
	out := map[string]*Device{}
	err := db.Query(&Device{}).ReadAll(
		func() model.Model { return &Device{} },
		func(m model.Model) { d := m.(*Device); out[d.Id] = d },
	)
	if err != nil {
		t.Fatalf("re-read failed: %v", err)
	}
	return out
}
