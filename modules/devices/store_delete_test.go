package devices

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

// The plural contract, through the real presenter: Delete ships ids (N=1 and
// N>1 in the same shape) and the store removes exactly those rows in one
// statement. Regression net for the silent no-op, where the store asserted
// the old singular *Device payload and every delete errored unheard.
func TestMemCallerBulkDeleteRemovesOnlyMarked(t *testing.T) {
	db := newSeededDeviceDB()
	pres := view.New(&deviceStore{db: db}, &Device{}, view.WithTitle("t"))
	deleter, ok := pres.(view.Deleter)
	if !ok {
		t.Fatal("presenter must implement view.Deleter")
	}
	if err := pres.Reload(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	if err := deleter.Delete("10", "11"); err != nil {
		t.Fatalf("bulk delete failed: %v", err)
	}

	var ids []string
	err := db.Query(&Device{}).ReadAll(
		func() model.Model { return &Device{} },
		func(m model.Model) { ids = append(ids, m.(*Device).Id) },
	)
	if err != nil {
		t.Fatalf("re-read failed: %v", err)
	}
	if len(ids) != 13 {
		t.Errorf("expected 13 rows left after deleting 2 of 15, got %d", len(ids))
	}
	for _, id := range ids {
		if id == "10" || id == "11" {
			t.Errorf("deleted id %q is still in the store", id)
		}
	}

	// N=1 ships the same shape: a single delete is a batch of one.
	if err := deleter.Delete("12"); err != nil {
		t.Fatalf("single delete failed: %v", err)
	}
	if findErr := db.Query(&Device{}).Where("id").Eq("12").ReadOne(); findErr == nil {
		t.Error("id 12 must be gone after the single delete")
	}
}
