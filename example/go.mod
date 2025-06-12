module github.com/octoberswimmer/masc/example

go 1.23.6

toolchain go1.24.2

replace github.com/octoberswimmer/masc => ../

require (
	github.com/gost-dom/browser v0.5.7
	github.com/octoberswimmer/masc v0.0.0-00010101000000-000000000000
	github.com/yuin/goldmark v1.5.2
)

require (
	github.com/gost-dom/css v0.1.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
)
