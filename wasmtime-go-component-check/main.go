// wasmtime-go-component-check — develop ブランチの wasmtime-go が
// WIT 由来の component を実機で扱えるかを検証するスモークテスト。
//
// 検証内容:
//   1. cargo-component が出力した .component.wasm を NewComponent でロード
//   2. ComponentLinker で Instantiate
//   3. ComponentType / ComponentItem / ComponentValType で型階層を探索
//      (slice 2: types)
//   4. ComponentInstance.GetFunc → ComponentFunc.Call で primitive 引数の
//      引き渡し (slice 2: values)
//
// 依存ブランチ: glassmonkey/wasmtime-go の develop (origin/develop と同期)。
// go.mod で replace ディレクティブによりローカルチェックアウトを参照。
package main

import (
	_ "embed"
	"fmt"
	"math"
	"os"

	"github.com/bytecodealliance/wasmtime-go/v44"
)

//go:embed guest/target/wasm32-wasip2/release/calculator.wasm
var calculatorWasm []byte

func main() {
	cfg := wasmtime.NewConfig()
	cfg.SetWasmComponentModel(true)
	engine := wasmtime.NewEngineWithConfig(cfg)
	store := wasmtime.NewStore(engine)

	component, err := wasmtime.NewComponent(engine, calculatorWasm)
	if err != nil {
		fail(fmt.Errorf("NewComponent: %w", err))
	}
	defer component.Close()

	fmt.Println("=== Component type introspection ===")
	if err := dumpComponentType(component); err != nil {
		fail(err)
	}

	linker := wasmtime.NewComponentLinker(engine)
	defer linker.Close()

	instance, err := linker.Instantiate(store, component)
	if err != nil {
		fail(fmt.Errorf("Instantiate: %w", err))
	}

	fmt.Println("\n=== Function calls ===")
	cases := []struct {
		name string
		args []interface{}
		want interface{}
	}{
		{"add-s32", []interface{}{int32(3), int32(5)}, int32(8)},
		{"add-s32", []interface{}{int32(-10), int32(4)}, int32(-6)},
		{"mul-s64", []interface{}{int64(1_000_000), int64(2_000_000)}, int64(2_000_000_000_000)},
		{"div-f64", []interface{}{float64(355), float64(113)}, float64(355.0 / 113.0)},
		{"negate-s32", []interface{}{int32(42)}, int32(-42)},
		{"sum-u32", []interface{}{uint32(4_000_000_000), uint32(294_967_295)}, uint32(4_294_967_295)},
		{"is-positive", []interface{}{int32(1)}, true},
		{"is-positive", []interface{}{int32(-1)}, false},
		{"is-positive", []interface{}{int32(0)}, false},
	}

	failed := 0
	for _, tc := range cases {
		fn := instance.GetFunc(store, tc.name)
		if fn == nil {
			fmt.Printf("  ✗ %s: GetFunc returned nil\n", tc.name)
			failed++
			continue
		}
		got, err := fn.Call(store, tc.args...)
		if err != nil {
			fmt.Printf("  ✗ %s%v: error %v\n", tc.name, tc.args, err)
			failed++
			continue
		}
		if !equal(got, tc.want) {
			fmt.Printf("  ✗ %s%v = %v, want %v\n", tc.name, tc.args, got, tc.want)
			failed++
			continue
		}
		fmt.Printf("  ✓ %s%v = %v\n", tc.name, tc.args, got)
	}

	if failed > 0 {
		fail(fmt.Errorf("%d case(s) failed", failed))
	}
	fmt.Printf("\nAll %d cases passed.\n", len(cases))
}

func dumpComponentType(component *wasmtime.Component) error {
	ct := component.Type()
	defer ct.Close()
	importN := ct.ImportCount()
	exportN := ct.ExportCount()
	fmt.Printf("  imports: %d\n  exports: %d\n", importN, exportN)
	for i := 0; i < exportN; i++ {
		name, item := ct.ExportNth(i)
		fmt.Printf("    [%d] %s  kind=%v\n", i, name, item.Kind())
		if vt := item.TypeAlias(); vt != nil {
			defer vt.Close()
			fmt.Printf("        valtype kind=%v\n", vt.Kind())
		}
		item.Close()
	}
	return nil
}

func equal(a, b interface{}) bool {
	switch av := a.(type) {
	case float64:
		bv, ok := b.(float64)
		if !ok {
			return false
		}
		// 浮動小数の桁あふれを避けて相対誤差比較
		return math.Abs(av-bv) < 1e-12*math.Max(math.Abs(av), 1)
	default:
		return a == b
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
