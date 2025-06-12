# Running and Building Examples

## Quick Start with masc CLI

The easiest way to run masc examples is with the `masc serve` command:

```bash
# Install the masc CLI (if not already installed)
go install github.com/octoberswimmer/masc/cmd/masc@latest

# Run the hello world example
cd example/hellomasc/
masc serve

# Or serve from the project root
masc serve ./example/hellomasc/

# Serve with custom port
masc serve -p 3000
```

The `masc serve` command automatically:
- Builds your Go application to WebAssembly
- Serves it on a local development server (default port 8000)
- Watches for file changes and rebuilds automatically
- Opens your browser to the running application

## Manual Building

Masc examples can also be built manually with the Go 1.14+ WebAssembly compilation target.

## Building for WebAssembly with Go 1.14+

**Ensure you are running Go 1.14 or higher.** Masc requires Go 1.14+ as it makes use of improvements to the `syscall/js` package which are not present in earlier versions of Go.

### Running examples with wasmserve (alternative method)

Examples can also be run using [`wasmserve`](https://github.com/hajimehoshi/wasmserve).

Install it (**using Go 1.14+**):

```bash
go install github.com/hajimehoshi/wasmserve@latest
```

Then run an example:

```bash
cd example/markdown/
wasmserve
```

Then navigate to http://localhost:8080/


## Building with other Go compilers

Other compilers such as [GopherJS](https://github.com/gopherjs) may work so long as they are compliant with the official Go 1.14+ compiler (support modules, the `syscall/js` interface, reflection, etc.)

Masc currently can only be built to run in web browsers.

## Testing the HelloMasc example

To run the example test, change into the `hellomasc` directory and run:

    cd example/hellomasc
    go test ./...
