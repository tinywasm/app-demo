//go:build !wasm

package reservation

import "github.com/tinywasm/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(
			Icon,
			"0 0 16 16",
			sprite.Path("M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 14.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13zM7.25 4v4.5l3.75 2.25.75-1.23-3-1.77V4h-1.5z"),
		),
	)
}
