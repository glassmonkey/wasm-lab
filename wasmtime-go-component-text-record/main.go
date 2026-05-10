// wasmtime-go-component-text-record — develop ブランチに乗せた
// string + list + record の marshaling が実機で動くかを WIT ベースで検証する。
//
// 検証関数:
//   テキスト系   greet(string)→string, upper(string)→string, words(string)→list<string>
//   構造体系     translate(record, f64, f64)→record, distance(record, record)→f64,
//                format-person(record)→string, make-person(...)→record
//
// record は Go 側で map[string]interface{} として渡す/受ける。
// list は []interface{}。
package main

import (
	_ "embed"
	"fmt"
	"math"
	"os"
	"reflect"

	"github.com/bytecodealliance/wasmtime-go/v44"
)

//go:embed guest/target/wasm32-wasip2/release/demo.wasm
var demoWasm []byte

func main() {
	cfg := wasmtime.NewConfig()
	cfg.SetWasmComponentModel(true)
	engine := wasmtime.NewEngineWithConfig(cfg)
	store := wasmtime.NewStore(engine)

	component, err := wasmtime.NewComponent(engine, demoWasm)
	if err != nil {
		fail(fmt.Errorf("NewComponent: %w", err))
	}
	defer component.Close()

	linker := wasmtime.NewComponentLinker(engine)
	defer linker.Close()

	// この component は (Rust ランタイム経由で) wasi:io/error を import
	// するが、本サンプル中の関数群では実際に呼ばれない。develop にはまだ
	// DefineWasi (wasip2 一括 link) が無いので、未解決 import を trap に
	// 落としておく。
	if err := linker.DefineUnknownImportsAsTraps(component); err != nil {
		fail(fmt.Errorf("DefineUnknownImportsAsTraps: %w", err))
	}

	instance, err := linker.Instantiate(store, component)
	if err != nil {
		fail(fmt.Errorf("Instantiate: %w", err))
	}

	cases := []struct {
		name string
		args []interface{}
		want interface{}
	}{
		// --- text patterns ---
		{
			"greet",
			[]interface{}{"world"},
			"Hello, world!",
		},
		{
			"greet",
			[]interface{}{"ダーリン"},
			"Hello, ダーリン!",
		},
		{
			"upper",
			[]interface{}{"hello world"},
			"HELLO WORLD",
		},
		{
			"words",
			[]interface{}{"  the quick   brown fox  "},
			[]interface{}{"the", "quick", "brown", "fox"},
		},

		// --- struct patterns ---
		{
			"translate",
			[]interface{}{
				map[string]interface{}{"x": float64(1), "y": float64(2)},
				float64(10),
				float64(20),
			},
			map[string]interface{}{"x": float64(11), "y": float64(22)},
		},
		{
			"distance",
			[]interface{}{
				map[string]interface{}{"x": float64(0), "y": float64(0)},
				map[string]interface{}{"x": float64(3), "y": float64(4)},
			},
			float64(5),
		},
		{
			"format-person",
			[]interface{}{
				map[string]interface{}{
					"name": "Alice",
					"age":  uint32(30),
					"tags": []interface{}{"engineer", "wasm-fan"},
				},
			},
			"Alice (age 30) [engineer, wasm-fan]",
		},
		{
			"make-person",
			[]interface{}{
				"Bob",
				uint32(42),
				[]interface{}{"reviewer", "skeptic"},
			},
			map[string]interface{}{
				"name": "Bob",
				"age":  uint32(42),
				"tags": []interface{}{"reviewer", "skeptic"},
			},
		},
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
			fmt.Printf("  ✗ %s: error %v\n", tc.name, err)
			failed++
			continue
		}
		if !equal(got, tc.want) {
			fmt.Printf("  ✗ %s\n      got:  %v\n      want: %v\n", tc.name, got, tc.want)
			failed++
			continue
		}
		fmt.Printf("  ✓ %s = %v\n", tc.name, summarize(got))
	}

	if failed > 0 {
		fail(fmt.Errorf("%d case(s) failed", failed))
	}
	fmt.Printf("\nAll %d cases passed.\n", len(cases))
}

func equal(a, b interface{}) bool {
	if af, ok := a.(float64); ok {
		bf, ok := b.(float64)
		if !ok {
			return false
		}
		return math.Abs(af-bf) < 1e-12*math.Max(math.Abs(af), 1)
	}
	return reflect.DeepEqual(a, b)
}

func summarize(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
