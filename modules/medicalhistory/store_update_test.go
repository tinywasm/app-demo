package medicalhistory

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/view"
)

// Bulk patch through the real presenter, wrapped in requirePatient so the
// wrapper's explicit Update forward is exercised: only the named columns move,
// across exactly the given ids; N=1 and N>1 are the same shape.
func TestMemCallerBulkUpdatePatchesOnlyNamedFields(t *testing.T) {
	db := newSeededVisitDB()
	pres := requirePatient{view.New(&visitStore{db: db}, &Visit{}, view.WithTitle("t"))}

	var updater view.Updater = pres // requirePatient satisfies it via the forward
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// Patch "reason" on two visits — Diagnosis (a different column) must survive.
	if err := updater.Update([]string{"v1", "v2"}, &Visit{Reason: "Revisión"}, []string{"reason"}); err != nil {
		t.Fatalf("bulk update failed: %v", err)
	}
	got := readAll(t, db)
	for _, id := range []string{"v1", "v2"} {
		if got[id].Reason != "Revisión" {
			t.Errorf("id %s: reason not patched, got %q", id, got[id].Reason)
		}
		if got[id].Diagnosis == "" || got[id].Diagnosis == "Revisión" {
			t.Errorf("id %s: diagnosis must be untouched, got %q", id, got[id].Diagnosis)
		}
	}
	if got["v3"].Reason == "Revisión" {
		t.Error("v3 was not in the id set and must be unchanged")
	}

	// N=1.
	if err := updater.Update([]string{"v3"}, &Visit{Doctor: "dr. House"}, []string{"doctor"}); err != nil {
		t.Fatalf("single update failed: %v", err)
	}
	got = readAll(t, db)
	if got["v3"].Doctor != "dr. House" {
		t.Errorf("v3: doctor not patched, got %q", got["v3"].Doctor)
	}
	if got["v3"].Patient == "" {
		t.Errorf("v3: patient must be untouched, got %q", got["v3"].Patient)
	}
}

func readAll(t *testing.T, db *orm.DB) map[string]*Visit {
	t.Helper()
	out := map[string]*Visit{}
	err := db.Query(&Visit{}).ReadAll(
		func() model.Model { return &Visit{} },
		func(m model.Model) { v := m.(*Visit); out[v.Id] = v },
	)
	if err != nil {
		t.Fatalf("re-read failed: %v", err)
	}
	return out
}
