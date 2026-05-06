package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strconv"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed guest.wasm
var guestWasm []byte

func main() {
	if len(os.Args) != 4 {
		usage()
	}
	op := os.Args[1]
	a, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fail(fmt.Errorf("parse a: %w", err))
	}
	b, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		fail(fmt.Errorf("parse b: %w", err))
	}

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	cfg := wazero.NewModuleConfig().WithStartFunctions("_initialize")
	mod, err := r.InstantiateWithConfig(ctx, guestWasm, cfg)
	if err != nil {
		fail(err)
	}

	fn := mod.ExportedFunction(op)
	if fn == nil {
		fail(fmt.Errorf("unknown function: %q (available: add, mul)", op))
	}

	res, err := fn.Call(ctx, api.EncodeI32(int32(a)), api.EncodeI32(int32(b)))
	if err != nil {
		fail(err)
	}

	fmt.Println(api.DecodeI32(res[0]))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: wazero-cli <add|mul> <a> <b>")
	os.Exit(2)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
