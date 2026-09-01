//go:build !wasm

package config

import "github.com/tinywasm/css"

// goToken exposes GoCyan as a design token so SetGradient can reference it —
// SetGradient takes Tokens, not raw hex. It is gradient-only: never emitted as
// its own custom property, so its var() fallback (GoCyan) is what resolves.
var goToken = css.Token{Name: "--color-go", Dark: GoCyan}

// Theme is the visual composition root. sitec discovers RootCSS() while walking
// the project at build time and emits it into web/public/style.css.
type Theme struct{}

// RootCSS turns the primary surface into a 135° gradient that runs from the
// WebAssembly logo violet (top-left) into the Go mascot's cyan (bottom-right).
// Every primary-filled surface picks it up at run time: the chassis header, the
// "Ficha Paciente" banner, the "+" action button, the field-label chips.
//
// WASMViolet also stays as the solid ColorPrimary underneath — that is what
// ColorOnPrimary text and the hover/focus/press states still derive from,
// since color-mix() has no gradient form.
func (Theme) RootCSS() *css.Stylesheet {
	return css.Theme(
		css.Set(css.ColorPrimary, WASMViolet),
		css.SetGradient(css.ColorPrimary, "135deg", css.ColorPrimary, goToken),
	)
}
