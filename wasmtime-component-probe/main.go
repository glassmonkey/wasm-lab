// wasmtime-go v44 が WASM Component Model モジュールを Go API から
// 読み込めるかを実機で確認するプローブ。
//
// 仮説: wasmtime-go v44 は Config.SetWasmComponentModel で機能フラグを
// 立てられるが、Component を「ロードする」Go API（NewComponent / ComponentLinker
// 相当）は未実装。NewModule はコア wasm 専用の C API を呼ぶため、
// コンポーネント binary を渡すとエラーになるはず。
//
// 検証方法:
//  1. 最小のコンポーネント preamble (magic + version 0x0d + layer 0x01)
//  2. これを wasmtime.NewModule に渡してエラー文字列を観察する
//  3. コア wasm preamble (version 0x01 + layer 0x00) を渡した場合のエラーと比較する
package main

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go/v44"
)

var (
	// Component Model の preamble: \0asm version=0x0d layer=0x01
	componentPreamble = []byte{
		0x00, 0x61, 0x73, 0x6d, // \0asm
		0x0d, 0x00, 0x01, 0x00, // version=13 (component), layer=1
	}
	// Core wasm の preamble: \0asm version=0x01 layer=0x00
	corePreamble = []byte{
		0x00, 0x61, 0x73, 0x6d,
		0x01, 0x00, 0x00, 0x00,
	}
)

func main() {
	fmt.Println("== probe: wasmtime-go v44 Component Model support ==")

	cfg := wasmtime.NewConfig()
	cfg.SetWasmComponentModel(true)
	engine := wasmtime.NewEngineWithConfig(cfg)

	probe(engine, "core wasm preamble", corePreamble)
	probe(engine, "component preamble", componentPreamble)
}

func probe(engine *wasmtime.Engine, label string, bytes []byte) {
	fmt.Printf("\n--- %s ---\n", label)
	fmt.Printf("first 8 bytes: % x\n", bytes)

	_, err := wasmtime.NewModule(engine, bytes)
	if err == nil {
		fmt.Println("=> NewModule accepted (unexpected for this minimal preamble)")
		return
	}
	fmt.Printf("=> NewModule error: %v\n", err)
}
