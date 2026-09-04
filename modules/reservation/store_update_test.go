package reservation

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/view"
)

func TestMemCallerBulkUpdatePatchesOnlyNamedFields(t *testing.T) {
	db := newSeededReservationDB()
	pres := view.New(&reservationStore{db: db}, &Reservation{}, view.WithTitle("t"))
	updater, ok := pres.(view.Updater)
	if !ok {
		t.Fatal("presenter must implement view.Updater")
	}
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if err := updater.Update([]string{"100", "101"}, &Reservation{Status: "attended"}, []string{"status"}); err != nil {
		t.Fatalf("bulk update failed: %v", err)
	}

	got := readAllReservations(t, db)
	for _, id := range []string{"100", "101"} {
		if got[id].Status != "attended" {
			t.Errorf("id %s: status not patched, got %q", id, got[id].Status)
		}
		if got[id].PatientName == "" || got[id].PatientName == "attended" {
			t.Errorf("id %s: PatientName must be untouched, got %q", id, got[id].PatientName)
		}
	}
	if got["102"].Status == "attended" {
		t.Error("id 102 was not in the id set and must be unchanged")
	}

	if err := updater.Update([]string{"102"}, &Reservation{Hour: "12:00"}, []string{"hour"}); err != nil {
		t.Fatalf("single update failed: %v", err)
	}
	got = readAllReservations(t, db)
	if got["102"].Hour != "12:00" {
		t.Errorf("id 102: hour not patched, got %q", got["102"].Hour)
	}
	if got["102"].PatientName == "" || got["102"].PatientName == "12:00" {
		t.Errorf("id 102: PatientName must be untouched, got %q", got["102"].PatientName)
	}
}

func TestMemCallerBulkUpdateRejectsEmptyFields(t *testing.T) {
	db := newSeededReservationDB()
	pres := view.New(&reservationStore{db: db}, &Reservation{}, view.WithTitle("t"))
	updater := pres.(view.Updater)
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if err := updater.Update([]string{"100"}, &Reservation{Status: "attended"}, nil); err == nil {
		t.Error("expected an error for an empty field set")
	}
}

func readAllReservations(t *testing.T, db *orm.DB) map[string]*Reservation {
	t.Helper()
	out := map[string]*Reservation{}
	err := db.Query(&Reservation{}).ReadAll(
		func() model.Model { return &Reservation{} },
		func(m model.Model) { r := m.(*Reservation); out[r.Id] = r },
	)
	if err != nil {
		t.Fatalf("re-read failed: %v", err)
	}
	return out
}
