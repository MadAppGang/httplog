module github.com/MadAppGang/httplog/examples/chi

go 1.19

replace github.com/MadAppGang/httplog => ../..

require (
	github.com/MadAppGang/httplog v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.0.7
)

require (
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)
