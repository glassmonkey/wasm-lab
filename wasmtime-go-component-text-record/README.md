# wasmtime-go-component-text-record

`wasmtime-go-component-check` の発展版。**テキスト (string) ベース** と **構造体 (record) ベース** の WIT パターンを 1 サンプルにまとめて、`glassmonkey/wasmtime-go` の `develop` ブランチに新しく載せた `string + list + record` の marshaling が実機で動くことを WIT 経由で検証する。

## 検証対象

| WIT 関数 | 入力 | 出力 | 何を試すか |
|---|---|---|---|
| `greet(name: string) -> string` | string | string | 文字列の双方向 |
| `upper(s: string) -> string` | string | string | UTF-8 ケース変換 |
| `words(s: string) -> list<string>` | string | list<string> | 戻りで list を受け取れるか |
| `translate(p: point, dx: f64, dy: f64) -> point` | record + 2 f64 | record | 引数で record を渡し戻りで record を受ける |
| `distance(a: point, b: point) -> f64` | record × 2 | f64 | record を複数渡してプリミティブ戻り |
| `format-person(p: person) -> string` | record (string + u32 + list<string>) | string | **ネスト**: record の中の list と多型フィールド |
| `make-person(name, age, tags) -> person` | string + u32 + list<string> | record | スカラ引数から record 構築 |

WIT:

```wit
record point   { x: f64, y: f64 }
record person  { name: string, age: u32, tags: list<string> }
```

## Go 側の表現

| WIT 型 | Go 型 |
|---|---|
| `string` | `string` |
| `list<T>` | `[]interface{}` (要素は再帰的にマーシャリング) |
| `record { ... }` | `map[string]interface{}` (キー = フィールド名) |
| primitives | `int32`/`uint32`/`float64`/`bool` 等 (既存どおり) |

## ビルド & 実行

```sh
make run
```

期待出力:

```
  ✓ greet = Hello, world!
  ✓ greet = Hello, ダーリン!
  ✓ upper = HELLO WORLD
  ✓ words = [the quick brown fox]
  ✓ translate = map[x:11 y:22]
  ✓ distance = 5
  ✓ format-person = Alice (age 30) [engineer, wasm-fan]
  ✓ make-person = map[age:42 name:Bob tags:[reviewer skeptic]]

All 8 cases passed.
```

## 依存関係 (sibling fork)

`go.mod` の `replace` で `../../wasmtime-go` (`glassmonkey/wasmtime-go` の develop チェックアウト) を直接参照する。詳しくは [`../wasmtime-go-component-check/README.md`](../wasmtime-go-component-check/README.md) と同じ説明が当てはまる。

## なぜ `DefineUnknownImportsAsTraps` を呼ぶか

cargo-component が出力する component は Rust ランタイム経由で `wasi:io/error@0.2.6` を import する (パニック処理用 stub)。本サンプルの関数群はパニックしない設計なので実際には呼ばれないが、未解決のままだと `Instantiate` が import 解決失敗で fail する。

develop には `add_wasip2` をラップする `DefineWasi` がまだ無い (Phase 2c 候補) ので、暫定的に **未解決 import を trap に落とす** ことで instantiate を通している。`DefineWasi` が入った時点で trap 化を取り除く想定。
