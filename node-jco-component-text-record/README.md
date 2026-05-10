# node-jco-component-text-record

[`wasmtime-go-component-text-record`](../wasmtime-go-component-text-record) と **同一の Rust component** を、ホストだけ Node + `jco` に差し替えて同じ 8 cases を回すサンプル。Component Model の **ポータビリティ** (1 つの component が異なるホスト言語で動く) を実機で示す。

## 対比表

| | Go 版 (wasmtime-go) | Node 版 (jco) |
|---|---|---|
| ホスト言語 | Go | JavaScript (Node 20+) |
| ランタイム | libwasmtime (CGO) | jco transpile → V8 |
| `record` の Go/JS 表現 | `map[string]interface{}` | プレーン JS オブジェクト `{ x: 1, y: 2 }` |
| `list<T>` の表現 | `[]interface{}` | プレーン JS 配列 `[1, 2, 3]` |
| 関数名の Go/JS マッピング | WIT そのまま (`format-person`) | キャメルケース化 (`formatPerson`) |
| 起動コスト | `Instantiate` ~ms | jco の ESM ロード + `instantiate` ~数十 ms |
| バイナリ配布 | 単一 Go バイナリ (要 libwasmtime native) | Node + `dist/` 一式 |

## 構成

```
node-jco-component-text-record/
├── guest/                # Rust component (wasmtime-go-component-text-record と同一の WIT/実装、独立ビルド)
│   ├── wit/world.wit
│   ├── Cargo.toml
│   └── src/lib.rs
├── host/
│   ├── package.json      # @bytecodealliance/preview2-shim, jco
│   └── main.js           # transpile された ESM を import → 8 cases 実行
└── Makefile
```

## ビルド & 実行

```sh
make run
```

期待出力:

```
  ✓ greet = "Hello, world!"
  ✓ greet = "Hello, ダーリン!"
  ✓ upper = "HELLO WORLD"
  ✓ words = ["the","quick","brown","fox"]
  ✓ translate = {"x":11,"y":22}
  ✓ distance = 5
  ✓ format-person = "Alice (age 30) [engineer, wasm-fan]"
  ✓ make-person = {"name":"Bob","age":42,"tags":["reviewer","skeptic"]}

All 8 cases passed.
```

## なぜ jco は楽なのか

jco は WIT 由来の component を **JS native value にネイティブ変換** する transpile 出力を生成するので:

- `record` → ふつうの `{}` (フィールド名そのまま)
- `list<T>` → ふつうの `[]`
- `string` → JS string
- 戻り値の deep-equal が `JSON.stringify` レベルで自然

Go (wasmtime-go) 側は generic な `interface{}` 経由でやり取りするから、Go のリッチな型システムをコンポーネント型に当てるのは呼び手の責任になる (`map[string]interface{}` を組み立てる手間)。**jco が WIT 型をホスト言語に「最適に」マッピングできる JS の動的さの旨味** が出ている。

## なぜ wasi imports の link が要らないのか

Go 版では `DefineUnknownImportsAsTraps` で wasi:io/error を trap 化してたが、jco は **`@bytecodealliance/preview2-shim` で Preview 2 imports をデフォルトで全部繋ぐ**ので、何もしなくても通る。同じ component の対比として、ホスト側の WASI 実装の差分が見える。
