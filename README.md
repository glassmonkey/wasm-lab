# wasm-lab

WebAssembly のユースケースを試すサンドボックス。各ディレクトリが独立したサンプルで、**ホストランタイム × WASI バージョン × 言語境界** の組み合わせを段階的に検証していく。

## サンプル一覧

| ディレクトリ | ホスト | ゲスト | テーマ |
|---|---|---|---|
| [`wazero-cli/`](./wazero-cli) | wazero (Pure Go) | Go (`//go:wasmexport` reactor) | wazero でホスト⇄ゲスト関数呼び出しの最小例 (`add`/`mul`) |
| [`wazero-string/`](./wazero-string) | wazero | Go reactor | linear memory 越しに文字列を渡す。`alloc`/`free` + packed i64 戻り値 |
| [`wasmtime-cli/`](./wasmtime-cli) | wasmtime-go (CGO) | Go reactor | `wazero-cli` と同題材を wasmtime-go に移植して比較 |
| [`wasmtime-component-probe/`](./wasmtime-component-probe) | wasmtime-go | – | wasmtime-go v44 が **Component Model 非対応** であることの実証 |
| [`node-jco-html/`](./node-jco-html) | Node.js + jco | Rust component | **WASI Preview 2 + wasi-http** で example.com を fetch & HTML パース |

## 学習の流れ

1. **`wazero-cli`** で wazero 上の最小 Host⇄Guest 関数呼び出しを把握
2. **`wazero-string`** で linear memory 経由のデータ授受、ABI 設計を学ぶ
3. **`wasmtime-cli`** で別ホスト (wasmtime-go) との API 差分を体感
4. **`wasmtime-component-probe`** で Component Model に踏み込もうとして wasmtime-go の壁にぶつかる
5. **`node-jco-html`** で Node.js + jco を選択し、Component + wasi-http を実機で動かす

## 各サンプルの方針

- **TinyGo は使わない**（標準 Go ツールチェインのみ）
- ゲストが Go の場合は **reactor モード** (`GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared`) が既定
- 新規サンプルのデフォルトホストは **wasmtime-go**（wazero との比較対象として残す）
- WASI Preview 2 / Component Model の検証には **Node + jco** を使う
- `.wasm` 成果物・`node_modules` 等のビルド成果物はコミットしない

詳細は [`CLAUDE.md`](./CLAUDE.md) を参照。
