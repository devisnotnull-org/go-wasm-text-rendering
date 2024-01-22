GOOS=js GOARCH=wasm go build -o ./build.wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./
