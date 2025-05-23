# Building examples

Masc examples can be built with the Go 1.14+ WebAssembly compilation target.

## Building for WebAssembly with Go 1.14+

**Ensure you are running Go 1.14 or higher.** Masc requires Go 1.14+ as it makes use of improvements to the `syscall/js` package which are not present in earlier versions of Go.

### Running examples

The easiest way to run the examples as WebAssembly is via [`wasmserve`](https://github.com/hajimehoshi/wasmserve).

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
