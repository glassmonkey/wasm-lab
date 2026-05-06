// jco transpile が出力した ESM (./dist/fetcher.js) を import して、
// guest component の extract(url) を呼ぶだけのホスト。
//
// 依存:
//   - @bytecodealliance/jco               (transpile 用、devDep)
//   - @bytecodealliance/preview2-shim     (jco 出力が import する WASI shim)
//
// jco transpile の出力は内部で preview2-shim を import するので、
// ここで明示的に import しなくても WASI Preview 2 (wasi-http 含む) が
// Node の fetch / fs / clocks にバインドされる。
import { extract } from "./dist/fetcher.js";

const url = process.argv[2] ?? "https://example.com";

try {
  const info = extract(url);
  console.log(JSON.stringify(info, null, 2));
} catch (err) {
  console.error("error:", err);
  process.exit(1);
}
