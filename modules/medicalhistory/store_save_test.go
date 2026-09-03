package medicalhistory

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

// Regression net for the silent save breakage (see devices' equivalent): the
// store asserted *Visit but view ships saveArgs{records}. Now the wire is read.
func TestMemCallerSaveThroughThePresenter(t *testing.T) {
	db := newSeededVisitDB()
	pres := requirePatient{view.New(&memCaller{db: db}, &Visit{}, "visit_list",
		func() model.ModelSlice { return &visitList{} },
		view.WithSaveOp("visit_save"),
	)}

	var saver view.Saver = pres
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if err := saver.Save(&Visit{Id: "v9", Patient: "Ana Lima", Doctor: "dr. X", Date: "2026-09-01", Reason: "Control", Diagnosis: ""}); err != nil {
		t.Fatalf("create-via-save failed: %v", err)
	}
	got := readAll(t, db)
	if got["v9"] == nil || got["v9"].Patient != "Ana Lima" {
		t.Errorf("new record not persisted: %+v", got["v9"])
	}

	if err := saver.Save(&Visit{Id: "v1", Patient: "Juan Pérez", Doctor: "dra. Nueva", Date: "2026-07-20", Reason: "Control", Diagnosis: "Actualizado"}); err != nil {
		t.Fatalf("update-via-save failed: %v", err)
	}
	got = readAll(t, db)
	if got["v1"].Doctor != "dra. Nueva" || got["v1"].Diagnosis != "Actualizado" {
		t.Errorf("existing record not replaced: %+v", got["v1"])
	}
}
