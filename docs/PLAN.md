---
PLAN: "feat: reservation module — calendar filter + targethour list"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 853638464225669385
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Part of `RESERVATION_MODULE_MASTER_PLAN.md` (monorepo root), **Phase B**.
>
> ⚠️ **DO NOT DISPATCH until Phase A is published**: this plan imports
> `github.com/tinywasm/components/targethour`, which does not exist until the
> `components` plan
> (`https://github.com/tinywasm/components/blob/main/docs/PLAN.md`) merges and
> `components` is re-tagged. After that, bump `components` in `go.mod` and
> dispatch this.
>
> **DRAFT** — the maintainer will refine it.

# Plan — `app-demo` module `reservation` ("Reserva Hora")

A fourth demo module, built on `crudview` exactly like
`modules/medicalhistory`, with two differences:

1. **The filter control is a month calendar** (`calendarslider.CalendarSlider`)
   instead of a search bar. Picking a day filters the list to that day's
   reservations.
2. **The aside list is `targethour.TargetHour`** (hour lead + status tint)
   instead of `targetdate` / `targetlist`.

The footer action buttons are **unchanged** — `crudview` already provides
create (`+`) / `↺`, `🗑`, `✏`, capability-gated and collapsing to what is
actionable (`layout@v0.2.6`).

Reference (legacy Pa100T v3.0.1): `http://192.168.122.10:1100/#reservation`,
login `12345678-9`.

---

## Ecosystem rules

`app-demo` is a **demo application**, not a published library:

- **UI text (labels, titles, toasts) in Spanish** — match `devices` /
  `medicalhistory`. Code, identifiers, comments, error messages **in English**.
- **No Go stdlib in files that reach WASM** (`model.go`, `reservation.go`,
  `svg.go` is `!wasm` and MAY use stdlib): use `github.com/tinywasm/fmt`.
  `_test.go` files may use stdlib.
- `store.go` is backend-only demo wiring and legitimately uses `orm`,
  `storage`, `storage/mem` — do NOT "fix" those imports.
- **Embed `dom.Element` by value.** **No `map[...]`** in WASM-reachable files.
- **SSR split**: the module icon geometry goes in `svg.go` under
  `//go:build !wasm`, method `IconSvg` on `*Module`.
- `docs/PLAN.md` stays at the app root (next to `go.mod`).
- **Tests**: `gotest`, never `go test`.

---

## Stage 0 — verify the toolchain

`go.mod` is already on the published deps you need — `components v0.6.9`
(carries `targethour` and `calendarslider` as `widget.Filterable`),
`layout v0.2.9`, `view v0.2.3` — with **no `replace` directives** for them.
Do NOT add any `replace`. Run `go build ./...` and `gotest`; both must be
clean before writing any module code. If a dep needs bumping, `go get
<mod>@<tag>` + `go mod tidy` — never a local `replace`.

---

## Stage 1 — `modules/reservation/model.go`

Mirror `modules/medicalhistory/model.go` (a `model.Definition` + a plain
struct implementing `model.Model` + `view.Itemizer`). Fields, from the
reference form:

| Field name | Type | NotNull | Notes |
|---|---|---|---|
| `id` | `input.Text()` | yes | PK (`DB: &model.FieldDB{PK: true}`) |
| `patient_run` | `input.Text()` | yes | RUN/DNI |
| `patient_name` | `input.Text()` | yes | |
| `patient_birthday` | `input.Text()` | no | "YYYY-MM-DD" (age is derived, not stored) |
| `patient_contact` | `input.Text()` | no | free text |
| `day` | `input.Text()` | yes | "YYYY-MM-DD" — the reservation's date (comes from the calendar) |
| `hour` | `input.Text()` | yes | "HH:MM" |
| `detail` | `input.Text()` | no | observación |
| `status` | `input.Text()` | no | `""` \| `"confirmed"` \| `"attended"` |

> The maintainer may swap some `input.Text()` for richer widgets
> (`input.Date()`, `input.Time()`) later — ship `input.Text()` for all so the
> demo compiles against the current `tinywasm/input`.

Struct:

```go
type Reservation struct {
	Id, PatientRun, PatientName, PatientBirthday, PatientContact string
	Day, Hour, Detail, Status                                    string
}
```

Implement `ModelName`, `Schema`, `Pointers`, `IsNil`, `EncodeFields`,
`DecodeFields` — copy the shape from `medicalhistory/model.go`'s `Visit`.

`view.Itemizer` — this is the ONLY view-specific method on the model:

```go
func (r *Reservation) Item() view.Item {
	return view.Item{
		ID:       r.Id,
		LeadMain: r.Hour,       // targethour renders this as the prominent hour
		Label:    r.PatientName,
		Description: statusLabel(r.Status), // "" | "Confirmada" | "Atendida"
	}
}

func statusLabel(s string) string {
	switch s {
	case "confirmed":
		return "Confirmada"
	case "attended":
		return "Atendida"
	}
	return ""
}
```

Also a `reservationList` implementing `model.ModelSlice` — copy
`medicalhistory/model.go`'s `visitList`.

---

## Stage 2 — `modules/reservation/store.go`

**Copy `modules/devices/store.go` verbatim** (it is the current, correct
reference — it already reads the `view` plural wire through the `readArgs`
`model.FieldWriter` walk, and its `device_save` case reads `saveArgs{records}`
rather than asserting the payload type). Then:

- Rename `Device`→`Reservation`, `deviceDB`→`reservationDB`,
  `deviceList`→`reservationList`, `device_*`→`reservation_*` ops, the walker
  types (`deviceArgs`/`deviceSink`/`stringSink`/`discardSink`) →
  `reservationArgs`/`reservationSink`/... (keep them local to the package —
  the demo tier keeps one `memCaller` per module, not shared).
- Seed `reservationDB` with ~8 reservations across 2–3 days near "today", so
  picking a day in the calendar shows a non-empty list. Vary `Status`
  (`""`, `"confirmed"`, `"attended"`) so the tint is visible.
- The `reservation_list` op returns ALL rows (newest first, like
  `device_list`). Day filtering is done in the presenter's `Filter` (Stage 3),
  NOT in the store.

Anti-footgun: do NOT assert `args.(*Reservation)` anywhere — always go through
the `readArgs` walk. The `*_save` case iterates `readArgs(args).records`.

---

## Stage 3 — `modules/reservation/reservation.go`

Mirror `modules/medicalhistory/medicalhistory.go`. Key pieces:

### The presenter wrapper — day filter

`medicalhistory` wraps `view.New`'s presenter in `requirePatient` to make
`Filter(term)` mean "which patient". Here it means **"which day"**:

```go
// byDay wraps the generic Presenter so Filter(term) means "the reservations
// of the day term" (term is "YYYY-MM-DD", emitted by the calendar via
// widget.Filterable). No term -> no rows (nothing is shown until a day is
// picked), same "module logic, not view's" decision as medicalhistory.
type byDay struct {
	view.Presenter
}

func (p byDay) Filter(term string) []view.Item {
	if term == "" {
		return nil
	}
	var rows []*Reservation
	_ = reservationDB.Query(&Reservation{}).Where("day").Eq(term).ReadAll(
		func() model.Model { return &Reservation{} },
		func(m model.Model) { rows = append(rows, m.(*Reservation)) },
	)
	items := make([]view.Item, len(rows))
	for i, r := range rows {
		items[i] = r.Item()
	}
	return items
}
```

Forward `Save`, `Update`, `Delete` explicitly (embedding an interface does NOT
promote extra methods) with `var _ view.Saver/Updater/Deleter = byDay{}`
compile proofs — copy the exact pattern from `medicalhistory.go`'s
`requirePatient`.

### Construction

```go
pres := byDay{view.New(&memCaller{db: reservationDB}, &Reservation{}, "reservation_list",
	func() model.ModelSlice { return &reservationList{} },
	view.WithTitle("Reserva Hora"),
	view.WithSaveOp("reservation_save"),
	view.WithUpdateOp("reservation_update"),
	view.WithDeleteOp("reservation_delete"),
)}

cal := &calendarslider.CalendarSlider{}   // satisfies widget.Filterable after Phase A

cv, err := crudview.New(crudview.Config{
	ParentID:  m.ModelName(),
	Presenter: pres,
	IDs:       ids,
	Filter:    cal,   // the calendar IS the filter — crudview auto-wires
	                  // widget.Filterable in Init, term = "YYYY-MM-DD"
	List: func(selected *SignalString, onSelect func(view.Item)) crudview.ListView {
		return &targethour.TargetHour{
			Selected: selected,
			OnSelect: onSelect,
			StatusOf: func(it view.Item) targethour.Status {
				switch it.Description {
				case "Confirmada":
					return targethour.StatusConfirmed
				case "Atendida":
					return targethour.StatusAttended
				}
				return targethour.StatusPending
			},
		}
	},
})
```

### Toasts

`cv.OnNew`, `cv.OnSaved`, `cv.OnDeleted`, `cv.OnUpdated` — copy the shape from
`devices.go` (Spanish strings: "Nueva reserva", "Guardado", "Eliminado …",
"Actualizado …").

### `Module` + `platformd.UIModule`

`ModelName() "reservation"`, `Label() "Reserva Hora"`, `Icon()` →
`svg.Icon("mod-reservation")`. `New(p *platformd.Platform) *Module`.

---

## Stage 4 — `modules/reservation/svg.go` (`//go:build !wasm`)

```go
//go:build !wasm

package reservation

import "github.com/tinywasm/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 16 16", sprite.Path("<a clock / calendar-clock path — maintainer picks the glyph>")),
	)
}
```

Use a clock or calendar-with-clock glyph, viewBox `0 0 16 16`, single
`fill="currentColor"` path (that is what `sprite.Path` emits). If unsure, reuse
FontAwesome "clock" solid scaled to a 16-box.

---

## Stage 5 — register the module

**`web/client.go`**: add the import
`"github.com/tinywasm/app-demo/modules/reservation"` and add
`reservation.New(p),` to the `p.Modules = []platformd.UIModule{...}` slice
(after `medicalhistory.New(p)`).

---

## Stage 6 — tests

- `modules/reservation/store_save_test.go` and
  `modules/reservation/store_update_test.go` — copy `modules/devices/`'s
  equivalents, rename to `Reservation`. Assert PATCH semantics for update
  (only named columns move) and that save persists through the plural wire.
- `modules/reservation/reservation_test.go` (`//go:build !wasm`,
  `package reservation`): build `byDay{view.New(...)}`, seed 2 reservations on
  the same day + 1 on another, assert `Filter("2026-09-10")` returns exactly
  the 2, `Filter("")` returns none, and each returned `view.Item` has
  `LeadMain` set to the hour and `Description` set to the localized status.

---

## Acceptance criteria

```bash
go build ./...                                             # clean
GOOS=js GOARCH=wasm go build -o /dev/null ./web/           # wasm client builds
GOOS=js GOARCH=wasm go list -deps ./web/ | grep tinywasm/svg/sprite   # prints NOTHING
gotest                                                     # all green
grep -n "reservation.New(p)" web/client.go                 # module registered
grep -rn "args.(\*Reservation)" modules/reservation/       # prints NOTHING (never assert the payload)
```

Manual (dev server, hot reload): open `#reservation` → the aside shows a month
calendar (no search bar); pick a day with seeded reservations → the list fills
with `targethour` rows (hour lead, status tint on confirmed/attended); the
footer shows only `+` until a day with rows is picked, then `+ 🗑 ✏`; creating
a reservation fires "Guardado".

## Stages table

| # | File(s) | Action |
|---|---|---|
| 0 | `go.mod` | bump `components` to the `targethour` tag; `go mod tidy` |
| 1 | `modules/reservation/model.go` | `Reservation` model + `reservationList` + `Item()` |
| 2 | `modules/reservation/store.go` | copy `devices/store.go`, rename, seed ~8 rows over 2–3 days |
| 3 | `modules/reservation/reservation.go` | `byDay` day-filter wrapper + `crudview.New` with calendar Filter + `targethour` List |
| 4 | `modules/reservation/svg.go` | module icon (clock glyph) |
| 5 | `web/client.go` | register `reservation.New(p)` |
| 6 | `modules/reservation/*_test.go` | store save/update tests + day-filter test |
