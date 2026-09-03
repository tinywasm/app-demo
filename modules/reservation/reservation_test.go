//go:build !wasm

package reservation

import (
	"testing"

	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

func TestByDayFilter(t *testing.T) {
	pres := byDay{view.New(&memCaller{db: reservationDB}, &Reservation{}, "reservation_list",
		func() model.ModelSlice { return &reservationList{} },
	)}

	if items := pres.Filter(""); items != nil {
		t.Fatalf("expected nil items for empty filter term, got %d items", len(items))
	}

	items := pres.Filter("2026-09-10")
	if len(items) != 4 {
		t.Fatalf("expected 4 items for 2026-09-10, got %d", len(items))
	}

	foundConfirmed := false
	foundAttended := false
	for _, it := range items {
		if it.LeadMain == "" {
			t.Errorf("expected LeadMain (hour) to be set on view.Item, got empty")
		}
		if it.Description == "Confirmada" {
			foundConfirmed = true
		}
		if it.Description == "Atendida" {
			foundAttended = true
		}
	}

	if !foundConfirmed {
		t.Errorf("expected at least one item with description 'Confirmada'")
	}
	if !foundAttended {
		t.Errorf("expected at least one item with description 'Atendida'")
	}

	itemsOther := pres.Filter("2026-09-12")
	if len(itemsOther) != 1 {
		t.Fatalf("expected 1 item for 2026-09-12, got %d", len(itemsOther))
	}
	if itemsOther[0].Label != "Diego Castro" {
		t.Errorf("expected label 'Diego Castro', got %q", itemsOther[0].Label)
	}
}
