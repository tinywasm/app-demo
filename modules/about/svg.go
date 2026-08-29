//go:build !wasm

package about

import "github.com/tinywasm/svg/sprite"

// IconSvg registra el glifo del módulo. tinywasm/ssr fusiona el resultado de cada
// IconSvg() del grafo en un único sprite inyectado en <body>.
//
// El receptor se instancia como valor cero (`&about.Module{}`), así que este
// método no puede tocar ningún campo: solo devuelve definiciones.
func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 16 16", sprite.Path("m7 11h2v2h-2zm4-7c0.552 0 1 0.448 1 1v3l-3 2h-2v-1l3-2v-1h-5v-2h6zm-3-2.5c-1.736 0-3.369 0.676-4.596 1.904s-1.904 2.86-1.904 4.596 0.676 3.369 1.904 4.596 2.86 1.904 4.596 1.904 3.369-0.676 4.596-1.904 1.904-2.86 1.904-4.596-0.676-3.369-1.904-4.596-2.86-1.904-4.596-1.904zm0-1.5c4.418 0 8 3.582 8 8s-3.582 8-8 8-8-3.582-8-8 3.582-8 8-8z")),
	)
}
