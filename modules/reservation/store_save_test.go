package reservation

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

func TestMemCallerSaveThroughThePresenter(t *testing.T) {
	db := newSeededReservationDB()
	pres := view.New(&memCaller{db: db}, &Reservation{}, "reservation_list",
		func() model.ModelSlice { return &reservationList{} },
		view.WithSaveOp("reservation_save"),
	)
	saver, ok := pres.(view.Saver)
	if !ok {
		t.Fatal("presenter must implement view.Saver")
	}
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if err := saver.Save(&Reservation{
		Id:          "200",
		PatientRun:  "12345678-0",
		PatientName: "Nuevo Paciente",
		Day:         "2026-09-15",
		Hour:        "12:00",
		Status:      "confirmed",
	}); err != nil {
		t.Fatalf("create-via-save failed: %v", err)
	}
	got := readAllReservations(t, db)
	if got["200"] == nil || got["200"].PatientName != "Nuevo Paciente" {
		t.Errorf("new record not persisted: %+v", got["200"])
	}

	if err := saver.Save(&Reservation{
		Id:          "100",
		PatientRun:  "12345678-9",
		PatientName: "María Gonzalez Renombrada",
		Day:         "2026-09-10",
		Hour:        "09:30",
		Status:      "attended",
	}); err != nil {
		t.Fatalf("update-via-save failed: %v", err)
	}
	got = readAllReservations(t, db)
	if got["100"].PatientName != "María Gonzalez Renombrada" || got["100"].Hour != "09:30" {
		t.Errorf("existing record not replaced: %+v", got["100"])
	}
}
