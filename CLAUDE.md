# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Purpose

WebAssembly のユースケースを試すためのサンドボックス。各ディレクトリが独立したサンプルで、それぞれが特定のシナリオ（言語境界、プラグイン機構、ブラウザ実行、サーバサイド WASI、軽量サンドボックスなど）を検証する。

## Repository Layout

- 各サンプルはトップレベルのディレクトリに 1 つずつ配置する（例: `go-host-rust-guest/`, `wasi-plugin/`, `browser-image-filter/`）
- ディレクトリ名はホスト言語・ゲスト言語・題材を表現する短いケバブケース
- 各サンプル直下に独自の `README.md` を置き、目的・実行コマンド・前提ランタイム（wasmtime / wazero / browser など）を記述する
- 共通コードを抽出するのは、3 つ以上のサンプルで重複が出てから

## Host Runtime Policy

- **新規サンプルのデフォルトホストは wasmtime-go** (`github.com/bytecodealliance/wasmtime-go/v44+`) を使う
  - 理由: WASI Preview 1 完全対応 + Preview 2 (`wasi:sockets`, `wasi-http`) への将来パスを確保するため
  - wazero は Preview 2 に対応する気配がなく、socket 系・http 系を試したい用途では行き止まり
- ただし **wazero は CGO 不要 / バイナリ軽量 / クロスコンパイル容易** という別軸の利点があるため、`wazero-*` サンプルは比較資料として残す
- 同じ題材を両ホストで動かして対比できる場合（例: `wazero-cli` ↔ `wasmtime-cli`）は両方置いて良い
- TinyGo は使わない方針（標準 Go ツールチェインのみ）。`//go:wasmexport` の戻り値 1 個制限などはこの方針から派生する制約

## Working on a Sample

- 新しいサンプルを始めるときは、まず空ディレクトリを作って `README.md` を書き、その下で個別に依存管理（`go.mod` / `Cargo.toml` / `package.json` など）を初期化する
- ビルド・実行コマンドはサンプルごとに完結させる（リポジトリ全体のタスクランナーは置かない）
- 採用ランタイム（wasmtime-go / wazero / ブラウザなど）を README に明記する
- ゲスト側 Go コードは reactor モード (`GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared`) で書くのが既定

## Conventions

- ホスト側 / ゲスト側のコードはサブディレクトリで分け、どちらが WASM にコンパイルされるか名前で分かるようにする（例: `host/`, `guest/`）
- `.wasm` 成果物はコミットしない（ビルド手順を README に書く）
