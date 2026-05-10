# wasmtime-go-component-check

`glassmonkey/wasmtime-go` の **`develop` ブランチ** が WIT 由来の component を実機で扱えるかを確認するスモークテスト。

`develop` には Bytecode Alliance 本体に upstream マージされた基本 Component bindings (PR #281 相当) に加えて、fork 内 PR の **slice 2 (型階層 + 値マーシャリング)** がマージ済み。本サンプルはその両方を踏むコードを書いて empirical に動作確認する。

## 検証スコープ

| API | 確認内容 |
|---|---|
| `wasmtime.NewComponent` | cargo-component の出力 (`.component.wasm`) を読み込めるか |
| `Component.Type()` + `ComponentType.ImportCount/ExportCount/ExportNth` | 型階層 introspection (slice 2a) |
| `ComponentItem.Kind` / `TypeAlias()` | exports の種類が取れるか |
| `ComponentLinker.Instantiate` | 純計算 component の instantiate |
| `ComponentInstance.GetFunc` | export を引ける |
| `ComponentFunc.Call(args ...interface{})` | s32 / s64 / u32 / f64 / bool の双方向 marshaling (slice 2b) |

**未対応 (テスト対象外)**: 文字列 / list / record / tuple / variant / enum / option / result / flags / map / resource、wasi / wasi-http link。develop に着くまで本サンプルの guest は **プリミティブのみ** を使う設計。

## 構成

```
wasmtime-go-component-check/
├── guest/               # Rust component (cargo-component で WIT → wasm)
│   ├── wit/world.wit    # primitive only の calculator world
│   ├── Cargo.toml
│   └── src/lib.rs       # add-s32, mul-s64, div-f64, negate-s32, sum-u32, is-positive
├── main.go              # ホスト: introspect → call → 期待値比較
├── go.mod               # replace で ../../wasmtime-go (develop) を参照
└── Makefile
```

## 前提

1. **`glassmonkey/wasmtime-go` を兄弟ディレクトリに clone**
   ```sh
   # ~/src/github.com/glassmonkey/ 配下に両リポジトリを並べる
   gh repo clone glassmonkey/wasmtime-go
   cd wasmtime-go && git checkout develop
   ```

2. ツール:
   - Rust toolchain + `rustup target add wasm32-wasip2`
   - `cargo install cargo-component`
   - `cargo install wkg` (WIT deps fetch 用)
   - Go 1.24+

`go.mod` の `replace ... => ../../wasmtime-go` がローカルの develop チェックアウトを直接読みに行く。go module 経由ではなく path 直参照で、fork のままビルドできる。

## 実行

```sh
make run
```

期待出力:

```
=== Component type introspection ===
  imports: 0
  exports: 6
    [0] add-s32       kind=3
    [1] mul-s64       kind=3
    [2] div-f64       kind=3
    [3] negate-s32    kind=3
    [4] sum-u32       kind=3
    [5] is-positive   kind=3

=== Function calls ===
  ✓ add-s32[3 5] = 8
  ✓ add-s32[-10 4] = -6
  ✓ mul-s64[1000000 2000000] = 2000000000000
  ✓ div-f64[355 113] = 3.1415929203539825
  ✓ negate-s32[42] = -42
  ✓ sum-u32[4000000000 294967295] = 4294967295
  ✓ is-positive[1] = true
  ✓ is-positive[-1] = false
  ✓ is-positive[0] = false

All 9 cases passed.
```

## なぜ replace 直参照なのか

fork (`glassmonkey/wasmtime-go`) の `go.mod` は upstream のまま `module github.com/bytecodealliance/wasmtime-go/v44` を declare している。go module 経由で fork を取り込もうとすると、

```
declares its path as github.com/bytecodealliance/wasmtime-go/v44 ...
```

の不一致で fail する。fork の go.mod を書き換えるのは upstream PR を出す上で不便なので、本サンプルは **クロスリポジトリ開発の常套手段である path 直 replace** を採用した。

将来 develop が upstream に取り込まれて新しいタグ (例: v45.0.0) で配布されたら、`replace` を消して通常の `require` に戻せる。
