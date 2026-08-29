//go:build !wasm

package medicalhistory

import "github.com/tinywasm/svg/sprite"

// IconSvg registers the module's glyph — the original medicalhistory icon: a
// clipboard (patient chart) with a checkmark badge and two list lines.
// tinywasm/ssr fuses every IconSvg() in the graph into one sprite injected
// into <body>.
//
// The receiver is instantiated as a zero value (&medicalhistory.Module{}),
// so this method must not touch any field — it only returns definitions.
func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 16 16",
			sprite.Path("m13.69 2.75h-3.938c0-0.9669-0.7831-1.75-1.75-1.75s-1.75 0.7831-1.75 1.75h-3.938c-0.2415 0-0.4375 0.196-0.4375 0.4375v11.38c0 0.2415 0.196 0.4375 0.4375 0.4375h11.38c0.2415 0 0.4375-0.196 0.4375-0.4375v-11.38c0-0.2415-0.196-0.4375-0.4375-0.4375zm-5.688-0.875c0.483 0 0.875 0.392 0.875 0.875s-0.392 0.875-0.875 0.875-0.875-0.392-0.875-0.875 0.392-0.875 0.875-0.875zm5.25 12.25h-10.5v-10.5h1.75v1.312c0 0.2415 0.196 0.4375 0.4375 0.4375h6.125c0.2415 0 0.4375-0.196 0.4375-0.4375v-1.312h1.75z"),
			sprite.Path("m10.65 5.892c-0.5089 0-0.922 0.4131-0.922 0.922 0 0.5089 0.4131 0.922 0.922 0.922 0.5089 0 0.922-0.4131 0.922-0.922 0-0.5089-0.4131-0.922-0.922-0.922zm0 1.844h-0.6147c-0.5071 0-0.922 0.2766-0.922 0.6147v0.6147h3.073v-0.6147c0-0.3381-0.4149-0.6147-0.922-0.6147z"),
			sprite.Path("m4.251 11.72h6.125v0.875h-6.125z"),
			sprite.Path("m4.251 9.975h6.125v0.875h-6.125z"),
		),
	)
}
