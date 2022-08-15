module github.com/MadAppGang/httplog/examples/zap

go 1.19

replace github.com/MadAppGang/httplog => ../..

replace github.com/MadAppGang/httplog/zap => ../../zap

require (
	github.com/MadAppGang/httplog v0.0.0-00010101000000-000000000000
	github.com/MadAppGang/httplog/zap v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.22.0
)

require (
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)
