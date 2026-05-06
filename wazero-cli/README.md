# wazero-cli

最小構成の wazero 検証サンプル。Go で書いた関数を WASM (reactor モード) にビルドし、ホスト側 Go CLI から wazero でロードして呼び出す。

## 構成

```
wazero-cli/
├── guest/          # WASM にコンパイルされる側
│   ├── main.go     # //go:wasmexport で add / mul を公開
│   └── go.mod
├── main.go         # ホスト CLI（wazero で guest.wasm をロード）
├── go.mod
└── Makefile
```

## 前提

- Go 1.24+ （`//go:wasmexport` のため）
- TinyGo は使わない（標準ツールチェインのみ）

## ビルド & 実行

```sh
make build
./wazero-cli add 3 5    # => 8
./wazero-cli mul 6 7    # => 42
```

## ポイント

- ゲストは `-buildmode=c-shared` で **reactor モード** の wasm を生成
  - 普通の Command モードだと `_start` 終了後にモジュールが死んで exported 関数を呼べない
  - そもそも `//go:wasmexport` は c-shared 必須
- ホストは `wazero.NewModuleConfig().WithStartFunctions("_initialize")` で初期化関数を呼ぶ
  - wazero のデフォルト start function は `_start` だけなので、reactor wasm では明示が要る
- `guest.wasm` は `//go:embed` でホストバイナリに同梱
