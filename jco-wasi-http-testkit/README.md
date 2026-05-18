# jco-wasi-http-testkit

**HTTP リクエストハンドラを wasm component として書き、テストでは preview2-shim の `HTTPServer` で wrap してテスト用 loopback に listen** させるパターン。Go の `httptest` の精神 (本物の経路を通る) を継承しつつ、テスト対象側 (= サーバ実装) を wasm に逃がす。

## 設計

```
   ┌─────────────────────────────┐
   │  Rust component (wasm)      │  wasi:http/proxy world
   │   incoming-handler.handle   │  ← production と完全同一の handler
   └──────────────┬──────────────┘
                  │  jco transpile (ESM)
   ┌──────────────▼──────────────┐
   │  preview2-shim/http         │
   │   HTTPServer(incomingHandler) → 127.0.0.1:<port>
   └──────────────┬──────────────┘
                  │  loopback TCP
   ┌──────────────▼──────────────┐
   │  test.js                    │  await fetch(`http://127.0.0.1:${port}/...`)
   │   node:test + assert        │
   └─────────────────────────────┘
```

## なぜこの形か

| 観点 | 値 |
|---|---|
| Guest 改変 | **不要** ── `wasmtime serve` / Spin / wasm-cloud に deploy する component と **バイト単位で同一** |
| ネットワーク経路 | 本物 (Node の loopback) ── routing / header / status / body が production と同じパスを通る |
| デッドロック | 無し ── `fetch()` は async、preview2-shim の HTTPServer は別 worker thread |
| 並行性 | host runtime の責務 ── Promise.all で 10 並行 fetch しても wasm handler が順次 invoke される |
| 状態 | Guest はステートレス想定。ステートを持たせたい場合は **component instance を分ける** (一つの testkit 内では共有可) |

これは「**listen 側を wasm に逃がす**」設計の典型形。同じ component を:

- **production**: `wasmtime serve handler.wasm` で本物の HTTP サーバ
- **test**: `HTTPServer(incomingHandler)` で loopback の HTTP サーバ

の両方で動かせる。テストコードは普通の fetch クライアントなので、Rust / Node / Go / 何で書いてもいい。

## 構成

```
jco-wasi-http-testkit/
├── guest/                # Rust component (wasi:http/proxy)
│   ├── wit/world.wit     # include wasi:http/proxy@0.2.0
│   ├── Cargo.toml
│   └── src/lib.rs        # 単純ルーティング: /hello /empty /echo-path/* /teapot /unknown
├── host/
│   ├── package.json      # @bytecodealliance/preview2-shim, jco
│   └── test.js           # node:test で fetch() 経由の 6 ケース
└── Makefile
```

## ビルド & 実行

```sh
make run
```

期待出力:

```
ok 1 - GET /hello → 200 'Hello, world!'
ok 2 - GET /empty → 200 空ボディ
ok 3 - GET /echo-path/foo/bar → 200 (path をボディに echo)
ok 4 - GET /teapot → 418
ok 5 - GET /unknown → 404
ok 6 - 並行リクエスト: wasm handler が正しく interleave される
# tests 6
# pass 6
# fail 0
```

## 検証できる性質

- ステータスコード (200 / 418 / 404)
- レスポンスボディ (空ボディ / UTF-8 含む / path echo)
- ルーティング条件 (`startsWith` パターンなど)
- **並行性**: 10 リクエスト並行で wasm handler が順次正しく処理 (host event loop が driving)

## 関連: 並行性が wasm runtime のもの、という設計

wasm guest は単一 linear memory・単一実行コンテキスト、`handle` 自体は同期関数。ところが上記の並行 fetch がきちんと結果を返すのは、**ホスト (Node + preview2-shim) のイベントループと synckit worker が複数 invocation を順次 driving しているから**。WASI Preview 2 の poll-based / handler-based 設計が、**guest 並列なし + host 並行制御** という分担を可能にしている。`wasmtime serve` / Spin / wasm-cloud などのプロダクション環境はこの同じ仕組みを使って多数のリクエストを捌く。

## 既知の制限

- guest 側で設定した `content-type` レスポンスヘッダが preview2-shim 経由で抜け落ちる (テストでは検証しない)。ヘッダ系の網羅は別 issue。
- POST/PUT 等の **リクエストボディ** は本サンプルでは扱わない (handler 側で `req.consume()` → IncomingBody の async read 経路が必要)。次の発展テーマ。
- WASI Preview 3 の async handler シグネチャは未対応 (cargo-component 0.21.1 の async code-gen と runtime crate に API ドリフトあり、`wasi-preview3-probe` 参照)。
