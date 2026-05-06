package main

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"

	"github.com/bytecodealliance/wasmtime-go/v44"
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

	engine := wasmtime.NewEngine()
	store := wasmtime.NewStore(engine)

	// Reactor wasm でも Go runtime の init で WASI 呼び出しが要るので、
	// WASI を有効化したうえで linker に流し込む。
	wasiCfg := wasmtime.NewWasiConfig()
	store.SetWasi(wasiCfg)

	linker := wasmtime.NewLinker(engine)
	if err := linker.DefineWasi(); err != nil {
		fail(err)
	}

	module, err := wasmtime.NewModule(engine, guestWasm)
	if err != nil {
		fail(err)
	}

	instance, err := linker.Instantiate(store, module)
	if err != nil {
		fail(err)
	}

	// Reactor: 関数を呼ぶ前に _initialize を必ず実行する。
	if init := instance.GetExport(store, "_initialize"); init != nil {
		if _, err := init.Func().Call(store); err != nil {
			fail(fmt.Errorf("_initialize: %w", err))
		}
	}

	export := instance.GetExport(store, op)
	if export == nil {
		fail(fmt.Errorf("unknown function: %q (available: add, mul)", op))
	}

	res, err := export.Func().Call(store, int32(a), int32(b))
	if err != nil {
		fail(err)
	}

	fmt.Println(res.(int32))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: wasmtime-cli <add|mul> <a> <b>")
	os.Exit(2)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
