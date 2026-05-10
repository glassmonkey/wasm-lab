module github.com/glassmonkey/wasm-lab/wasmtime-go-component-check

go 1.26.1

require github.com/bytecodealliance/wasmtime-go/v44 v44.0.0

// 注意: 親ディレクトリに glassmonkey/wasmtime-go のチェックアウトがあり、
// その develop ブランチが checkout されていることを前提にしている。
// 詳細は README.md を参照。
replace github.com/bytecodealliance/wasmtime-go/v44 => ../../wasmtime-go
