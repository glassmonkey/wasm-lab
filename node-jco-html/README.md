# node-jco-html

**WASI Preview 2 (Component Model + wasi-http) で example.com を fetch して HTML をパース** するサンプル。`wasmtime-go` では実装不能だった経路 ([wasmtime-component-probe](../wasmtime-component-probe/) 参照) を、Bytecode Alliance 公式の **`jco`** で JS 化して Node.js 上で動かす。

## 構成

```
node-jco-html/
├── guest/                # Rust component
│   ├── wit/world.wit     # fetcher world: extract(url) を export、wasi-http を import
│   ├── Cargo.toml
│   └── src/lib.rs        # outgoing-handler で fetch + scraper でパース
├── host/
│   ├── package.json      # @bytecodealliance/preview2-shim, jco
│   └── main.js           # transpile 済み ESM を import して extract を呼ぶ
└── Makefile
```

## 前提

- Rust toolchain (`rustc` / `cargo`)
- `rustup target add wasm32-wasip2`
- `cargo install cargo-component`
- Node.js 20+
- `npm`

## ビルド & 実行

```sh
make build               # cargo component build → jco transpile → npm install
make run                 # node main.js https://example.com
```

実行例:

```sh
$ make run
{
  "title": "Example Domain",
  "links": [
    "https://www.iana.org/domains/example"
  ]
}
```

## なぜ jco + Node なのか

`wasi:http/outgoing-handler` を使うには **WASI Preview 2 の Component Model に対応したホストランタイム** が必要。比較表:

| ホスト | Component | wasi-http | 備考 |
|---|---|---|---|
| **wasmtime CLI** | ◎ | ◎ | Rust 製、リファレンス実装 |
| **wasmtime-go** | ✕ | ✕ | C ヘッダにあるが Go binding が wrap してない |
| **wazero** | ✕ | ✕ | Preview 2 対応の予定なし |
| **jco + Node** | ◎ | ◎ | コンポーネントを ES module に transpile して Node 上で実行。`preview2-shim` が `wasi-http` を Node `fetch` にバインド |

つまり「Go ホスト固定で wasi-http」は今日時点では塞がっており、**Node を選ぶことで初めて手元で wasi-http が走る** のが本サンプルの狙い。

## パイプライン詳細

1. **`cargo component build --target wasm32-wasip2`**
   - Rust コードを直接 Component にビルド
   - `wit/world.wit` から `src/bindings.rs` を自動生成
   - 出力: `target/wasm32-wasip2/release/html_fetcher.wasm` (純粋なコンポーネント)
2. **`jco transpile *.wasm -o dist/`**
   - コンポーネントを ESM (`fetcher.js` + 補助ファイル) に変換
   - `@bytecodealliance/preview2-shim` を import するコードが含まれる
3. **`node main.js`**
   - ESM を import して `extract(url)` を直接呼ぶ
   - `preview2-shim` が `wasi:http/outgoing-handler` を Node の `fetch` にディスパッチ
   - 本物の HTTPS リクエストが example.com に飛ぶ

## 設計メモ

- **HTML パースは `scraper`**: CSS セレクタが使える。`html5ever` 直よりずっと書きやすい代わりに依存ツリーは大きい (ビルド時間そこそこ)
- **`url` クレート**: URL パース用。host:port や query を切り出すのに使う
- **`http_get` 内のブロッキング読み**: `OutgoingFuture::subscribe().block()` で同期的に待ち、`InputStream::blocking_read` でボディを最後まで読む。非同期版もあるが学習目的なら sync が読みやすい
- **`scheme`/`authority`/`path-with-query`** を別々にセットするのが Preview 2 流。`http::Request` のような全部入り型ではなく、リソース (handle) を組み立てる
