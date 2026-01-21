module github.com/MadAppGang/httplog/v2/examples/alice

go 1.21

replace github.com/MadAppGang/httplog/v2 => ../..

require (
	github.com/MadAppGang/httplog/v2 v2.0.0
	github.com/justinas/alice v1.2.0
	github.com/justinas/nosurf v1.1.1
)

require (
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/sys v0.5.0 // indirect
)
