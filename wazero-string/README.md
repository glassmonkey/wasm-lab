# wazero-string

ホスト⇄ゲスト間で **文字列を linear memory 越しにやり取りする** サンプル。`wazero-cli` の発展版で、メモリ管理の作法を学ぶのが主目的。

## 構成

```
wazero-string/
├── guest/          # WASM 側
│   ├── main.go     # alloc / free / greet を export
│   └── go.mod
├── main.go         # ホスト CLI（メモリ書き込み・読み出し付き）
├── go.mod
└── Makefile
```

## ビルド & 実行

```sh
make build
./wazero-string greet world      # => Hello, world!
./wazero-string greet ダーリン   # => Hello, ダーリン!  (UTF-8 OK)
```

## ABI

ゲストの公開関数:

| 関数 | シグネチャ | 役割 |
|---|---|---|
| `alloc` | `(size i32) -> ptr i32` | ゲストの linear memory にバッファ確保 |
| `free`  | `(ptr i32)` | バッファ解放 |
| `greet` | `(in_ptr i32, in_len i32) -> packed i64` | 文字列処理。戻り値は `(out_ptr<<32) \| out_len` |

## 設計ポイント

### なぜ packed i64 なのか
Go 1.24+ の `//go:wasmexport` は **戻り値を 1 つしか許可しない**。wasm spec も wazero も multi-value 対応してるが、Go ツールチェイン側の制限。なので `(ptr, len)` を 32bit ずつ packing して i64 にして返している。TinyGo を使えば multi-value も可能。

### バッファの寿命管理
ゲスト側で `alloc` した `[]byte` は `keepAlive` グローバルマップに保持して GC されないようにする。理由: ホストはゲスト linear memory のアドレスを直接読み書きするので、対応するスライスが GC で回収されると問題になる（実際は wasip1 の Go GC は非移動なのでアドレスは固定だが、領域自体が解放されないことを保証する必要がある）。

### ホスト側の使い方
1. `alloc(N)` でゲストに入力用領域確保
2. `mod.Memory().Write(ptr, bytes)` で書き込み
3. `greet(in_ptr, in_len)` 呼び出し → packed i64 → unpack
4. `mod.Memory().Read(out_ptr, out_len)` で結果取得
5. `free(in_ptr)` / `free(out_ptr)` で解放

呼び出し側が **alloc/free のライフサイクル管理に責任を持つ** のが要点。
