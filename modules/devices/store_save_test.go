package devices

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

// Regression net for the silent save breakage: the store used to assert the
// singular *Device payload, but view ships saveArgs{records}, so every save
// errored unheard and the "Guardado" toast never fired. Now the wire is read.
func TestMemCallerSaveThroughThePresenter(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&memCaller{db: db}, &Device{}, "device_list",
		func() model.ModelSlice { return &deviceList{} },
		view.WithSaveOp("device_save"),
	)
	saver, ok := pres.(view.Saver)
	if !ok {
		t.Fatal("presenter must implement view.Saver")
	}
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// A brand-new record lands.
	if err := saver.Save(&Device{Id: "99", Name: "Nuevo", Ip: "10.0.0.99"}); err != nil {
		t.Fatalf("create-via-save failed: %v", err)
	}
	got := readAll(t, db)
	if got["99"] == nil || got["99"].Name != "Nuevo" {
		t.Errorf("new record not persisted: %+v", got["99"])
	}

	// Saving an existing id replaces the whole record.
	if err := saver.Save(&Device{Id: "10", Name: "Renombrado", Ip: "10.0.0.10"}); err != nil {
		t.Fatalf("update-via-save failed: %v", err)
	}
	got = readAll(t, db)
	if got["10"].Name != "Renombrado" || got["10"].Ip != "10.0.0.10" {
		t.Errorf("existing record not replaced: %+v", got["10"])
	}
}
