# wasmtime-cli

`wazero-cli` と同じ題材（Go 製 reactor wasm の `add` / `mul` を呼ぶ CLI）を、ホストランタイムだけ **wasmtime-go** に差し替えたもの。wazero 版との比較用サンプル。

## 構成

```
wasmtime-cli/
├── guest/          # WASM 側（wazero-cli の guest と同一）
│   ├── main.go     # //go:wasmexport で add / mul を公開
│   └── go.mod
├── main.go         # ホスト CLI（wasmtime-go 使用、CGO 必須）
├── go.mod
└── Makefile
```

## 前提

- Go 1.24+
- **CGO 有効** (`CGO_ENABLED=1`、デフォルト ON)
- C コンパイラ（macOS は Xcode CLT、Linux は gcc/clang）
- `github.com/bytecodealliance/wasmtime-go/v44` が prebuild の libwasmtime を同梱

## ビルド & 実行

```sh
make build
./wasmtime-cli add 3 5    # => 8
./wasmtime-cli mul 6 7    # => 42
```

## wazero との違い

| 観点 | wazero | wasmtime-go |
|---|---|---|
| 実装 | Pure Go | CGO 経由で Rust 製 wasmtime を呼ぶ |
| バイナリサイズ | ~9MB | **~24MB**（libwasmtime 同梱） |
| クロスコンパイル | `GOOS=...` だけで OK | 各 OS/arch 用 native lib が要る |
| reactor 起動 | `WithStartFunctions("_initialize")` で linker に任せる | **`_initialize` を明示的に Call する** |
| 関数呼び出し | `[]uint64` を介して `api.EncodeI32` 等で型変換 | `interface{}` を `int32` 等のネイティブ型で渡す |
| WASI Preview 1 | 概ね対応（sockets除く） | 完全対応 |
| WASI Preview 2 (sockets/http) | ✕ | wasmtime 本体 ◎、Go bindings は要検証 |

## ABI のポイント

reactor wasm を扱う際の作法は wazero 版と同じだが、**API スタイルが違う**:

```go
// reactor の初期化を明示的に呼ぶ必要がある
if init := instance.GetExport(store, "_initialize"); init != nil {
    init.Func().Call(store)
}

// 引数は interface{} だが内部で wasm 型にマップされる
res, _ := export.Func().Call(store, int32(a), int32(b))
fmt.Println(res.(int32))
```

wazero の `WithStartFunctions("_initialize")` のような「設定で済ませる」スタイルではなく、wasmtime-go では呼び出し側が手続き的に書く。
