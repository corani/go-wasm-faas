# Go WASM FaaS

Toy project to create a FaaS service in Go that allows creating functions as WASM modules.

This is based on the excellent article [FAAS in Go with WASM, WASI and Rust](https://eli.thegreenplace.net/2023/faas-in-go-with-wasm-wasi-and-rust/) by Eli Bendersky.

## Running

Use `./build.sh` to build the server and the example WASM modules (note that the WASM module requires Go 1.21 or later!)

Run the server using `./bin/server`.

To register the example function, navigate to http://localhost:8080/ and enter:
- Name: `add`
- File: upload `target/add.wasm`

To execute the function, navigate to http://localhost:8080/run/add?a=34&b=35
