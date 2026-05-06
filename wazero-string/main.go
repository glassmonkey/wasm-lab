package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed guest.wasm
var guestWasm []byte

func main() {
	if len(os.Args) != 3 || os.Args[1] != "greet" {
		fmt.Fprintln(os.Stderr, "usage: wazero-string greet <name>")
		os.Exit(2)
	}
	input := []byte(os.Args[2])

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	cfg := wazero.NewModuleConfig().WithStartFunctions("_initialize")
	mod, err := r.InstantiateWithConfig(ctx, guestWasm, cfg)
	if err != nil {
		fail(err)
	}

	allocFn := mod.ExportedFunction("alloc")
	freeFn := mod.ExportedFunction("free")
	greetFn := mod.ExportedFunction("greet")

	// 1. ゲスト linear memory に入力用バッファを確保
	res, err := allocFn.Call(ctx, uint64(len(input)))
	if err != nil {
		fail(err)
	}
	inPtr := uint32(res[0])
	defer freeFn.Call(ctx, uint64(inPtr))

	// 2. 入力バイト列をゲストメモリに書き込む
	if !mod.Memory().Write(inPtr, input) {
		fail(fmt.Errorf("write input: out of bounds"))
	}

	// 3. greet 呼び出し。戻り値は packed i64: (out_ptr<<32) | out_len
	res, err = greetFn.Call(ctx, uint64(inPtr), uint64(len(input)))
	if err != nil {
		fail(err)
	}
	packed := res[0]
	outPtr := uint32(packed >> 32)
	outLen := uint32(packed)
	defer freeFn.Call(ctx, uint64(outPtr))

	// 4. ゲストメモリから結果を読み出す
	out, ok := mod.Memory().Read(outPtr, outLen)
	if !ok {
		fail(fmt.Errorf("read output: out of bounds"))
	}

	fmt.Println(string(out))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
