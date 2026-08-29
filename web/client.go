//go:build wasm

package main

import (
	"github.com/tinywasm/components/themetoggle"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"

	// Global form skin: one import at the composition root makes EVERY form in the
	// app render as labeled fieldset boxes (CSS-only, collected via SSR).
	"github.com/tinywasm/app-demo/modules/about"
	"github.com/tinywasm/app-demo/modules/devices"
	"github.com/tinywasm/app-demo/modules/medicalhistory"
	_ "github.com/tinywasm/components/fieldset"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/svg"
)

// hiddenModule exists solo para probar CanView: se registra y el chasis no debe
// pintarlo ni en el rail ni en el escenario. No tiene paquete propio porque no es
// un módulo demo — es un caso de permisos.
type hiddenModule struct{}

func (hiddenModule) ModelName() string { return "hidden" }
func (hiddenModule) Label() string     { return "Oculto" }
func (hiddenModule) Icon() svg.Icon    { return about.Icon }
func (hiddenModule) View() Component   { return nil }

// demoBrand mocks the platform's own identity — what the shell is called and
// marked with. In a real application this comes from the deployment, not from
// a login. The mark must be NON-empty: demoIdentity's empty avatar already
// proves the fallback path, and this stub exists to prove the non-empty one —
// the two render different shapes in the header.
type demoBrand struct{}

func (demoBrand) BrandName() string { return "Demo Platform" }
func (demoBrand) BrandMark() string {
	// An inline SVG data URI, so the demo needs no asset server. The
	// percent-encoding of # is load-bearing: a raw # would read as a URL
	// fragment and the image would 404.
	return "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512'%3E%3Cpath d='M272.6 17.9L323.8 158.5l143.7 12.3c27.5 2.4 38.6 37.6 17.4 55.2l-109.5 96.5 32 142.6c6.1 27.2-23.7 48.5-47.9 34.2L256 403.9 152.9 499.3c-24.2 14.3-54-7-47.9-34.2l32-142.6-109.5-96.5C7.9 208.4 19 173.2 46.5 170.8l143.7-12.3L241.4 17.9c9.5-23.9 42.6-23.9 52.2 0z' fill='%23ffffff'/%3E%3C/svg%3E"
}

// demoIdentity mocks the login for the demo. In a real application this is
// whatever github.com/tinywasm/user hands back for the current session — see
// that repository's docs/PLAN.md for the work that makes it satisfy this
// contract directly.
type demoIdentity struct{}

func (demoIdentity) UserName() string   { return "Thor Odinson" }
func (demoIdentity) UserAvatar() string { return "" } // sin imagen: cae al glifo
func (demoIdentity) UserRoles() []string {
	// Tres, para que el caso N se vea en el demo.
	return []string{"Administrador", "Operaciones", "Soporte"}
}

func main() {
	p := &platformd.Platform{
		AppName:   "Demo Platform",
		Brand:     demoBrand{},
		DefaultID: "medicalhistory",
		User:      demoIdentity{},
		// A factory, not an instance: the shell renders one menu for the header
		// and one for the drawer, and a single component in both places would
		// put one id on two nodes — the second copy inert.
		UserActions: func() Component {
			return &themetoggle.ThemeToggle{}
		},
		CanView: func(id string) bool {
			return id != "hidden"
		},
	}

	p.Modules = []platformd.UIModule{
		devices.New(p),
		medicalhistory.New(p),
		about.New(),
		hiddenModule{},
	}

	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", platformd.Auto())

	select {}
}
