# wasi-preview3-probe

WASI Preview 3 (`future<T>` / `stream<T>` / `error-context`) を **2026-05 時点** の Rust + cargo-component ツールチェインでどこまで動かせるか実機で探ったログ。動いたものではなく **「壁の正確な位置」** を記録するサンプル。

## 環境スナップショット

| ツール | バージョン |
|---|---|
| Rust | 1.92.0 |
| Rust target `wasm32-wasip3` | **存在せず** |
| `cargo-component` | 0.21.1 (現時点最新) |
| `wit-bindgen-rt` | 0.39.0 (cargo-component 既定) / 0.44.0 (試験的) |

## 試した 3 つのパターン

```
probes/
├── future-simple/    export double-async: func(x: u32) -> future<u32>
├── error-context/    export try-divide: func(a: s32, b: s32) -> result<s32, error-context>
└── stream-bytes/     export read-bytes: func(n: u32) -> stream<u8>
```

それぞれ `make probe-future` / `probe-error` / `probe-stream` で再現可能。`make all` で全部一気に。

## 結果サマリ

| Probe | WIT 受容 | Rust コンパイル | wasm リンク | Wall |
|---|---|---|---|---|
| `future<u32>` | ✓ | ✕ | – | **A** |
| `error-context` | ✓ | ✓ | **✕** | **B** |
| `stream<u8>` | ✓ | ✕ | – | **A** |

### Wall A: cargo-component ↔ wit-bindgen-rt の async API ドリフト

cargo-component 0.21.1 が `bindings.rs` を生成するときに `wit_bindgen_rt::async_support::FutureVtable<T>` / `StreamVtable<T>` のフィールドや `future_new` / `stream_new` のシグネチャを **古い形** で参照する。一方 `wit-bindgen-rt` 0.39 / 0.44 のリリース版はもう構造が変わっていて噛み合わない。

具体エラー:

```
error[E0560]: struct `FutureVtable<u32>` has no field named `new`
error[E0061]: this function takes 2 arguments but 1 argument was supplied
              wit_bindgen_rt::async_support::future_new::<T>(/* fn() -> T */, T::VTABLE)
error[E0308]: mismatched types
              expected fn pointer `unsafe extern "C" fn() -> u64`
              found fn item       `unsafe extern "C" fn() -> u32`
```

`wit-bindgen-rt = "0.39"` (cargo-component 既定) でも `"0.44"` でも同じ系の不一致。**現時点の cargo-component に同梱される code-gen と公開リリースの runtime crate に互換ペアが存在しない**。GitHub から両方を Git 直 dep で揃えるか、cargo-component / wit-bindgen 双方の master を使えば動くかもしれないが、ユーザビリティとして "リリース済みものだけで future が動く" 状態には至っていない。

### Wall B: `wasm-component-ld` adapter に preview3 canonical builtins が無い

`error-context` は async path に乗らないので Rust コンパイルは通る (`async` feature を付ければ `error_context_new` も使える)。けれど `wasm-component-ld` が wasm を component に encode する段階で:

```
error: failed to encode component
  Caused by:
    0: failed to decode world from module
    1: module was not valid
    2: failed to resolve import `$root::[error-context-new;encoding=utf8]`
    3: no top-level imported function `[error-context-new;encoding=utf8]` specified
```

`error-context-new` は preview3 で導入される **canonical built-in 関数** で、ホスト (= wasmtime) または adapter が提供する想定。cargo-component 同梱の adapter (`wasi_snapshot_preview1.reactor.wasm` 系) はまだ提供してない。

## まとめ

- **WIT 構文レベル**: `future` / `stream` / `error-context` ともに 2026-05 の cargo-component で受け付けられる
- **コード生成**: ✓ bindings.rs まで生成される (生成物の API は不整合)
- **Rust コンパイル**: ✕ async 系は API ドリフトで詰まる、`error-context` だけ通る
- **wasm component リンク**: ✕ canonical built-in (`error-context-new` 等) を adapter が解決できない
- **wasmtime-go (libwasmtime v44) で実行**: ✕ そもそも作れないし、val kind enum に future/stream/error-context が無いので呼び戻しもできない

**結論**: Preview 3 は **WIT としては書ける**、**ツールチェイン全体では実用前段階**。動作確認したいなら 2026 年後半以降の wit-bindgen / cargo-component / wasmtime CLI を待つのが現実的。

## 関連リンク

- WIT 言語仕様 (preview3 syntax): <https://github.com/WebAssembly/component-model>
- wit-bindgen-rt 0.39: <https://docs.rs/wit-bindgen-rt/0.39.0/wit_bindgen_rt/async_support/>
- cargo-component 0.21.1: <https://github.com/bytecodealliance/cargo-component>

## このサンプルの意図

「**WASI Preview 3 を試したい**」と言ったときに、何が動いて何が動かないのか・どこで壁にぶつかるのかを **empirical に明示** する記録。`probes/*/wit/world.wit` と `Cargo.toml` だけ残せば、半年後に再ビルドして「壁が動いたか」を 1 コマンドで再評価できる仕組み。
