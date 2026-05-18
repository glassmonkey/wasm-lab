// jco-wasi-http-testkit — listen 側 (HTTP サーバ実装) が wasm component の
// パターン。Go の httptest との対比は逆方向で、こちらは **production と
// 同じ deploy 経路** をテストに使う:
//
//   wasi:http/proxy 実装の Rust component を、jco transpile + preview2-shim
//   の HTTPServer で wrap して 127.0.0.1 でリッスンさせる。
//   テスト本体は ふつうに fetch() で叩いて response を assert する。
//
// 大事な性質:
//   - guest 側のコードはテスト用に何も変えない (wasmtime serve / Spin に
//     deploy するのと完全に同一の component が走る)
//   - テストはネットワーク経路 (loopback) を本当に通る = "real path"
//   - fetch() は async なので main thread が固まらず、shim の synckit
//     worker と素直に協調する (前案のデッドロック問題は解消)
//   - 1 つの component を多数の URL/method/body 組合せで網羅できる

import { test, after } from "node:test";
import assert from "node:assert/strict";
import { HTTPServer } from "@bytecodealliance/preview2-shim/http";
import { incomingHandler } from "./dist/handler.js";

// テスト群で共有する 1 つの HTTPServer。component 側はステートレスなので
// 個別 test 間で共有しても安全。
const server = new HTTPServer(incomingHandler);
server.listen(0, "127.0.0.1");
const addr = server.address();
const baseUrl = `http://${addr.address}:${addr.port}`;

after(() => server.stop());

// ---- テストケース ---------------------------------------------------

test("GET /hello → 200 'Hello, world!'", async () => {
  const res = await fetch(`${baseUrl}/hello`);
  assert.equal(res.status, 200);
  assert.equal(await res.text(), "Hello, world!");
  // 注: content-type header は現状 shim 経由で抜け落ちる (preview2-shim
  // の OutgoingResponse → Node http.ServerResponse 変換に既知の制限がある)。
  // ヘッダ系の網羅は別 issue として残す。
});

test("GET /empty → 200 空ボディ", async () => {
  const res = await fetch(`${baseUrl}/empty`);
  assert.equal(res.status, 200);
  assert.equal(await res.text(), "");
});

test("GET /echo-path/foo/bar → 200 (path をボディに echo)", async () => {
  const res = await fetch(`${baseUrl}/echo-path/foo/bar`);
  assert.equal(res.status, 200);
  assert.equal(await res.text(), "/echo-path/foo/bar");
});

test("GET /teapot → 418", async () => {
  const res = await fetch(`${baseUrl}/teapot`);
  assert.equal(res.status, 418);
  assert.equal(await res.text(), "I'm a teapot");
});

test("GET /unknown → 404", async () => {
  const res = await fetch(`${baseUrl}/unknown`);
  assert.equal(res.status, 404);
  assert.equal(await res.text(), "not found");
});

test("並行リクエスト: wasm handler が正しく interleave される", async () => {
  // 並行性 (host runtime の責務) を実証: 10 リクエスト同時投げて全部正解で
  // 戻ってくる。guest 側はステートレスかつ単一スレッドだが、host (Node の
  // event loop + shim の synckit) が複数 invocation をシリアル化して
  // 順次回す。
  const targets = ["/hello", "/empty", "/echo-path/x", "/teapot", "/unknown"];
  const results = await Promise.all(
    Array.from({ length: 10 }, (_, i) =>
      fetch(`${baseUrl}${targets[i % targets.length]}`).then(async (r) => ({
        status: r.status,
        body: await r.text(),
      })),
    ),
  );
  // 期待値を再計算
  const expected = Array.from({ length: 10 }, (_, i) => {
    const path = targets[i % targets.length];
    if (path === "/hello") return { status: 200, body: "Hello, world!" };
    if (path === "/empty") return { status: 200, body: "" };
    if (path.startsWith("/echo-path")) return { status: 200, body: path };
    if (path === "/teapot") return { status: 418, body: "I'm a teapot" };
    return { status: 404, body: "not found" };
  });
  assert.deepEqual(results, expected);
});
