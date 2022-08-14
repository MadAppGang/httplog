module github.com/MadAppGang/httplog/examples/alice

go 1.19

replace github.com/MadAppGang/httplog => ../..

require (
	github.com/MadAppGang/httplog v0.0.0-00010101000000-000000000000
	github.com/justinas/alice v1.2.0
	github.com/justinas/nosurf v1.1.1
)

require (
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)
