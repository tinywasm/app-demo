module github.com/tinywasm/app-demo

go 1.25.2

require (
	github.com/tinywasm/components v0.6.0
	github.com/tinywasm/css v0.4.19
	github.com/tinywasm/dom v0.13.7
	github.com/tinywasm/fmt v0.25.7
	github.com/tinywasm/html v0.0.17
	github.com/tinywasm/input v0.0.3
	github.com/tinywasm/layout v0.2.1
	github.com/tinywasm/model v0.1.7
	github.com/tinywasm/orm v0.12.0
	github.com/tinywasm/storage v0.0.6
	github.com/tinywasm/svg v0.3.0
	github.com/tinywasm/time v0.5.4
	github.com/tinywasm/unixid v0.2.26
	github.com/tinywasm/view v0.2.1
)

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/form v0.4.0 // indirect
	github.com/tinywasm/json v0.5.23 // indirect
	github.com/tinywasm/router v0.1.30 // indirect
	github.com/tinywasm/widget v0.6.19 // indirect
)

replace github.com/tinywasm/layout => ../layout

replace github.com/tinywasm/components => ../components

replace github.com/tinywasm/css => ../css
