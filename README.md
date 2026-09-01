# app-demo
<img src="docs/img/badges.svg">

TinyWasm demo app — the `platformd` shell wired with three CRUD modules,
served from this dedicated repo so the demo never weighs on library `go.mod`s.

The reference demo used by [github.com/tinywasm/layout](https://github.com/tinywasm/layout)
(`platformd`): a playground to import ANY tinywasm package freely and test it
in a running app.

## Modules

- `modules/devices` — full CRUD (form + list) over an in-memory `storage/mem` store via `layout/crudview`.
- `modules/medicalhistory` — CRUD with a `selectsearch`/`targetdate` list.
- `modules/about` — static module.
- `web/client.go` — the composition root: `platformd.Platform` + `hiddenModule`
  (exercises `CanView`), mock brand/identity, `themetoggle` user action, and
  the `fieldset` global form skin.

## Run

    tinywasm            # from this repo; dev server :8080, MCP :6060

## Layout dependency

The demo consumes the shell via a local replace while layout is in monorepo
development:

    replace github.com/tinywasm/layout => ../layout

With no replace, resolution falls back to the published `github.com/tinywasm/layout`
module. (A clone outside this workspace needs either the published version or
the `../layout` checkout next to it.)