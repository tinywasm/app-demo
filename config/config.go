// Package config is the demo's composition root. Today it only carries the
// visual theme (see css.go); the shared server/client symbols the framework
// layout expects would land here too if the demo grew a real backend.
package config

// The two brand colors the primary surface is built from:
//
//   - WASMViolet: the official WebAssembly logo purple.
//   - GoCyan:     the Go gopher mascot's light blue.
//
// They are the endpoints of the primary gradient defined in css.go.
const (
	WASMViolet = "#654FF0"
	GoCyan     = "#00ADD8"
)
