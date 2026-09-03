module github.com/tinywasm/app-demo

go 1.25.2

require (
	github.com/tinywasm/components v0.6.8
	github.com/tinywasm/css v0.4.20
	github.com/tinywasm/dom v0.13.8
	github.com/tinywasm/fmt v0.25.7
	github.com/tinywasm/html v0.0.19
	github.com/tinywasm/input v0.0.3
	github.com/tinywasm/layout v0.2.8
	github.com/tinywasm/model v0.1.7
	github.com/tinywasm/orm v0.12.0
	github.com/tinywasm/storage v0.0.6
	github.com/tinywasm/svg v0.3.3
	github.com/tinywasm/time v0.5.4
	github.com/tinywasm/unixid v0.2.26
	github.com/tinywasm/view v0.2.3
)

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/form v0.4.0 // indirect
	github.com/tinywasm/icons v0.0.2 // indirect
	github.com/tinywasm/json v0.5.23 // indirect
	github.com/tinywasm/router v0.1.30 // indirect
	github.com/tinywasm/widget v0.6.21 // indirect
)

// Local replaces for unreleased work: the daemon serves these live, so no
// publish is needed to verify in the running demo. Drop each line once its
// repo is published past the change.

replace github.com/tinywasm/css => ../css
