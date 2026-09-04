# Agent Guide — `tinywasm/app-demo`

Read this before touching any file here.

`app-demo` is not a throwaway. It is the **worked example** of how you build an
application with the TinyWasm framework: someone evaluating the framework opens
`modules/devices/` and reads it top to bottom to learn the pattern. So the code
here is held to a stricter *readability* bar than a normal app — the code **is**
the documentation.

The one rule everything below serves: **a Go developer who has never seen
TinyWasm must be able to read a module and understand it without a guide.**

---

## What a module looks like — always these files, always short

Every module is the same four flat files (never a subfolder inside a module):

| File | One responsibility | Rough ceiling |
|---|---|---|
| `model.go` | the record struct, its field schema, the in-memory seed data | ~120 lines |
| `<module>.go` | the wiring: `View()` builds the presenter + `crudview` and hooks its callbacks | ~100 lines |
| `store.go` | the in-memory backend: one `switch op` over the CRUD operations | ~120 lines |
| `svg.go` | `//go:build !wasm`, the module's nav glyph | ~20 lines |

`about/` is the minimal case (a static module: `about.go` + `svg.go`). `devices/`
is the canonical CRUD case — **copy it to make a new module**, rename the type,
swap the model. That symmetry is the lesson; do not let a module drift into its
own shape.

**If a file blows past its ceiling, stop.** A long `store.go` full of adapter
boilerplate is not "just how it is" — it means a piece of wiring that every
module repeats belongs *upstream*, in the library that owns that seam. File it
as a plan against that library (`view`, `orm`, `layout`, …). Do not paste the
boilerplate a fourth time. This is the framework's own rule:
`tinywasm/app-releases/docs/CONSTRUCTION_HARNESS.md` — *"A missing contract at a
boundary is a defect in the library, not in the consumer."*

---

## Familiar Go only

The demo must look like ordinary, boring Go. Concretely, in module code:

- **No generics, no reflection, no `any`** except where a library signature
  forces it at the I/O edge (`model.Encodable` in `store.go`'s `Call`).
- **Plain structs with explicit fields.** No builder chains of your own, no
  clever embedding to save a line.
- **A `switch op string`** is how `store.go` dispatches — one `case` per
  operation, each a handful of lines. If a `case` needs a helper, the helper is
  a named function right below, not an inline closure three levels deep.
- **Names say what they are.** `deviceDB`, `newSeededDeviceDB`, `memCaller` —
  not `db`, `mk`, `c2`.
- **Comments explain the framework touch-point**, not the Go. Assume the reader
  knows Go and does not know TinyWasm.

---

## SRP — one file, one job

- `model.go` never imports `dom` and never builds UI. It is the shape + the
  seed, nothing else.
- `<module>.go` never touches storage. It builds `view.New(...)`,
  `crudview.New(...)`, and wires `OnSaved`/`OnDeleted`/`OnUpdated` to
  `p.Notify(...)`. That is all `View()` does.
- `store.go` never builds a `dom.Element`. It adapts the in-memory `orm.DB` to
  the `router.Caller` seam the presenter drives.
- `web/client.go` is the composition root: it constructs `platformd.Platform`
  and appends every module. Adding a module is **one line** here.

---

## DRY across modules = a library gap, never a local copy

If `devices`, `medicalhistory` and `reservation` all contain the *same* helper,
that helper is missing from a library. The fix is upstream, not a shared file in
this repo (`app-demo` is a leaf; it has nothing to share *to*).

> **Known gap — the `store.go` payload walk.** `view` ships every write wrapped
> in an unexported struct (`saveArgs{records}`, `updateArgs{ids,fields,record}`,
> `deleteArgs{ids}`), so `store.go` cannot type-assert the payload — it walks
> the `model.FieldWriter` encoding by hand (`readArgs` + the little sink types).
> That block is currently near-identical in all three CRUD modules. It is
> tracked for a `view` plan (expose the walk, or a typed accessor, in the
> library). **Until that lands: copy the existing `readArgs` block verbatim
> from `devices/store.go`. Do not invent a variant.**

---

## The demo mirrors a real deployment

- **No `if dev` / no test-only code paths / no fixtures a real backend would not
  have.** The in-memory store is a *real* `orm.DB` over `storage/mem`, seeded
  with realistic rows, exercising real save-on-blur / delete / bulk-patch. It
  resets on a full page reload — that is the honest trade-off of an in-memory
  store, documented in `store.go`, not a bug to paper over.
- **Placeholder data only.** Never a real client's names, IPs, or records — the
  seeds are invented (`Pc Administracion`, `Juan Pérez`, …).

---

## Language

`app-demo` is a demo, not a public library, so it follows the app convention:

- **UI text in Spanish** — module labels, titles, toast messages, field
  placeholders (`"Buscar..."`, `"Guardado"`, `"Reserva Hora"`).
- **Identifiers, types, operation strings in English** — `Device`, `deviceDB`,
  `"device_save"`.
- **Comments in Spanish** is fine here (the existing modules do it); keep them
  about the framework touch-point.

---

## The seam — every module implements `platformd.UIModule`

```go
type Module struct{ p *platformd.Platform }

func New(p *platformd.Platform) *Module { return &Module{p: p} }

func (m *Module) ModelName() string { return "devices" }
func (m *Module) Label() string     { return "Computadores" }
func (m *Module) Icon() svg.Icon    { return Icon }
func (m *Module) View() Component    { /* build crudview */ }

var _ platformd.UIModule = (*Module)(nil)   // the compile-time proof, always present
```

A module that needs more (e.g. `CanView` for visibility gating) adds the extra
interface method and its own `var _ platformd.Xxx = (*Module)(nil)` line. The
chassis discovers capabilities by assertion — the module never registers itself.

---

## Running it

```
tinywasm          # from this repo — dev server :8080, MCP :6060, hot reload
```

Hot reload picks up every `.go` / `css.go` change automatically. Do **not** run
`go build` to "apply" a change; just edit and look at the running app. A one-off
`GOOS=js GOARCH=wasm go build -o /dev/null ./web/` is only for a compile check
before publishing, plus the sprite-leak check:

```
GOOS=js GOARCH=wasm go list -deps ./web/ | grep tinywasm/svg/sprite   # must be empty
```

---

## Tests

```
go install github.com/tinywasm/devflow/cmd/gotest@latest
gotest
```

`gotest`, never `go test`. Stdlib `testing` only. Each CRUD module carries
`store_save_test.go` / `store_update_test.go` / `store_delete_test.go` driving
the presenter through the real `orm.DB` — the "consumer-shaped test" the harness
requires. Publish with `gopush 'message'`.

---

## Documentation

`README.md` indexes the modules and how the demo is wired. Update it when you
add or remove a module. This file (`AGENTS.md`) is the construction standard —
update it when the module shape itself changes.
