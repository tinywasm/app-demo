---
PLAN: "refactor!: app-demo stores implement the typed view.Backend instead of faking router.Caller"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 15409905331598733492
PR: https://github.com/tinywasm/app-demo/pull/2
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Do NOT run `gopush` or `codejob`.
>
> **This is a BREAKING-change follow-up (Fase B).** `github.com/tinywasm/view`
> v0.3.0 replaced its transport seam (`router.Caller` + op strings) with a
> domain seam. This repo is migrated to it. **Do not touch any other repo from
> here.**

# PLAN — `app-demo`: `memCaller` → typed stores (`view.Backend`)

## Why (read this before writing code)

`view.New` no longer takes a `router.Caller`. Its signature is now:

```go
func New(b Backend, record model.Model, opts ...Option) Presenter
```

with the domain seam (from `view/backend.go`, v0.3.0):

```go
type Backend interface {
	List() ([]model.Model, error)
}
type BackendSaver interface {
	Save(recs []model.Model) error
}
type BackendUpdater interface {
	Update(ids []string, rec model.Model, fields []string) error
}
type BackendDeleter interface {
	Delete(ids []string) error
}
```

Only `WithTitle` and `WithSearchPlaceholder` options still exist. The
presenter's capabilities mirror the backend's: it is a `view.Saver` iff the
backend implements `BackendSaver`, and so on — no strings involved.

Each CRUD module in this repo currently fakes a transport: `store.go` holds a
`memCaller` with `Call(op, args, into, done)` + a `switch op` + ~80 lines of
hand-written `model.FieldWriter` sinks (`readArgs`, `deviceArgs`,
`stringSink`, `deviceSink`, `discardSink`) that unpack view's unexported wire
envelopes. All of that dies. What survives is the real logic: the seeded
`orm.DB`, the upsert loop, the `UpdateFields` patch, the single-statement
`Delete`, and the newest-first `List` order.

## Repo rules

- Demo code, not a public library: **Spanish** comments/UI text are fine (the
  existing modules do it); identifiers in English.
- Familiar Go only: **no generics, no reflection, no `any`** except where a
  library signature forces it. Plain structs, explicit fields.
- `gotest`, never `go test`. First prerequisite on a fresh machine:
  `go install github.com/tinywasm/devflow/cmd/gotest@latest`.
- `go.mod` already requires `github.com/tinywasm/view v0.3.0`. If it does not
  resolve, run `go get github.com/tinywasm/view@v0.3.0` first — then migrate.
- Do NOT bump `github.com/tinywasm/layout`: its public `crudview` API is
  unchanged (its own migration is tests-only and lands separately).
- `AGENTS.md` (repo root) documents the module shape — Stage 4 updates it.

---

## Stage 1 — `modules/devices` (the worked example; copy its shape twice)

### 1a. `modules/devices/store.go`: `memCaller` → `deviceStore`

Keep `deviceDB` and `newSeededDeviceDB` byte-for-byte. Delete everything from
`// memCaller adapts` (line 45) to the end of the file (line 202): the
`memCaller` type, its `Call`, its `Dispatch`, `readArgs`, `deviceArgs` + all
its methods, `stringSink`, `deviceSink`, `discardSink`. In their place write:

```go
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

func (s *deviceStore) Save(recs []model.Model) error {
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

func (s *deviceStore) Delete(ids []string) error {
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
```

Notes: the old code returned the `orm` errors from `ReadAll` directly
(`err = c.db.Query...ReadAll(...)`); the new `List` does the same. The old
`device_list` case silently returned zero rows when `into` was not a
`*deviceList`; the new `List` has no such hole. The old error texts were
`"memCaller: device_save: ..."`; they are renamed to the `"deviceStore: ..."`
shape above (no test asserts the old strings — verified by grep). Imports of
`store.go` become `model`, `orm`, `storage`, `storage/mem`, dot-`fmt` (the
`router` import, if present, leaves).

### 1b. `modules/devices/devices.go`: `View()` over the store

```go
pres := view.New(&deviceStore{db: deviceDB}, &Device{}, view.WithTitle("Computadores"))
```

Delete the `"device_list"` string, the `func() model.ModelSlice` factory, and
the three `WithXOp` options. The `"github.com/tinywasm/model"` import of
`devices.go` becomes unused (its only use was the factory) — remove that
import line too.

### 1c. `modules/devices/model.go`: delete `deviceList`

Delete the `deviceList` struct and all its methods plus the
`var _ model.ModelSlice = (*deviceList)(nil)` line. (`model.go` keeps using
`model` everywhere else, so its imports stay.)

### 1d. `modules/devices/store_save_test.go`, `store_update_test.go` (two
`view.New` sites: lines ~16 and ~63), `store_delete_test.go`: construction only

Replace each

```go
pres := view.New(&memCaller{db: db}, &Device{}, "device_list",
	func() model.ModelSlice { return &deviceList{} },
	view.WithSaveOp("device_save"),
)
```

with

```go
pres := view.New(&deviceStore{db: db}, &Device{}, view.WithTitle("t"))
```

keeping each site's original capability set implied (the store implements all
four, so every assertion still holds). **Do not touch any other line of the
tests** — they drive the presenter through the real `orm.DB` and are the spec
that the store rewrite preserved the behavior — except one import:
`store_save_test.go` uses `model` only in the deleted factory line, so remove
its `"github.com/tinywasm/model"` import line (`store_update_test.go` and
`store_delete_test.go` keep theirs: their `readAll` helpers still use it).

---

## Stage 2 — `modules/medicalhistory` (same shape, visit vocabulary)

- `store.go`: keep `Patient`, `todayAgenda`, `visitDB`, `newSeededVisitDB`
  byte-for-byte. Replace `memCaller` + `readArgs` + sinks with `visitStore`
  (`List`/`Save`/`Update`/`Delete` + an `anyIDs` helper), copying Stage 1a and
  renaming `Device→Visit`, `deviceStore→visitStore`, `deviceDB→visitDB`. The
  per-case `orm` logic (upsert loop, `UpdateFields`, single-statement `Delete`,
  newest-first order) is the code already in this file's `visit_*` cases —
  preserve it line-for-line, only changing where the inputs come from (typed
  parameters instead of `readArgs(args)`).
- `medicalhistory.go`: `pres := requirePatient{view.New(&visitStore{db: visitDB}, &Visit{}, view.WithTitle("Ficha Paciente"))}`.
  `requirePatient` (embedding + `Save`/`Update`/`Delete` forwards + the four
  `var _` guards) compiles unchanged — keep it. Reword its `Filter` comment
  that says "same convention memCaller's visit_list uses" to "same convention
  visitStore.List uses".
- `model.go`: delete `visitList` + methods + its `var _ model.ModelSlice` line.
- `store_save_test.go`, `store_update_test.go`: construction lines only, same
  replacement as Stage 1d (store `visitStore{db: db}`, record `&Visit{}`).
  `store_save_test.go` also loses its `"github.com/tinywasm/model"` import
  (used only by the deleted factory); `store_update_test.go` keeps its import.

## Stage 3 — `modules/reservation` (same shape, reservation vocabulary)

- `store.go`: keep `reservationDB` + seed byte-for-byte. `memCaller` →
  `reservationStore` + `anyIDs`, preserving this file's `reservation_*` orm
  logic line-for-line.
- `reservation.go`: `pres := byDay{view.New(&reservationStore{db: reservationDB}, &Reservation{}, view.WithTitle("Reserva Hora"))}`.
  `byDay` compiles unchanged — keep it.
- `model.go`: delete `reservationList` + methods + its `var _` line.
- `store_save_test.go`, `store_update_test.go` (two `view.New` sites),
  `reservation_test.go`: construction lines only. `store_save_test.go` and
  `reservation_test.go` also lose their `"github.com/tinywasm/model"` imports
  (used only by the deleted factories); `store_update_test.go` keeps its
  import.

## Stage 4 — docs + green

- `AGENTS.md`: the module-shape table says `store.go` is "one `switch op` over
  the CRUD operations" (~120 lines) — rewrite that row: `store.go` is the
  in-memory backend implementing `view.Backend` (`List` + the `Save`/`Update`/
  `Delete` capabilities it supports). Delete the "Known gap — the `store.go`
  payload walk" section (the gap this plan closes). Update the SRP row for
  `store.go` ("adapts the in-memory `orm.DB` to the `router.Caller` seam")
  to the `view.Backend` seam, and the `<module>.go` row's `view.New(...)`
  mention to `view.New(store, record, ...)`. Keep everything else.
- Run `gotest ./...` (whole repo) until green.
- Compile check for the demo web target:
  `GOOS=js GOARCH=wasm go build -o /dev/null ./web/` must succeed.

## Acceptance

- `grep -rn "memCaller\|readArgs\|deviceArgs\|visitArgs\|reservationArgs\|stringSink\|deviceSink\|discardSink" --include="*.go" modules/` → empty.
- `grep -rn "WithSaveOp\|WithUpdateOp\|WithDeleteOp\|WithArgs\|device_list\|visit_list\|reservation_list\|deviceList\|visitList\|reservationList" --include="*.go" modules/` → empty.
- `grep -rn "router\." --include="*.go" modules/` → empty.
- `modules/devices/store.go` ≤ ~110 lines; the other two stores shrink by the
  same ~80-line deletion.
- `gotest ./...` green. `GOOS=js GOARCH=wasm go build -o /dev/null ./web/`
  succeeds.

## Stages

| # | Scope | Files |
|---|---|---|
| 1 | devices store + wiring + tests | `modules/devices/store.go`, `devices.go`, `model.go`, `store_save_test.go`, `store_update_test.go`, `store_delete_test.go` |
| 2 | medicalhistory | `modules/medicalhistory/store.go`, `medicalhistory.go`, `model.go`, `store_save_test.go`, `store_update_test.go` |
| 3 | reservation | `modules/reservation/store.go`, `reservation.go`, `model.go`, `store_save_test.go`, `store_update_test.go`, `reservation_test.go` |
| 4 | docs + green | `AGENTS.md`, full `gotest`, WASM build |

---

## NOT in this plan

`layout/crudview` (its migration is tests-only and has its own plan),
`auth` (already migrated), anything outside this repo.
