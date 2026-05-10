// node-jco-component-text-record host
//
// `wasmtime-go-component-text-record` と **同じ Rust component** を、
// 今度は Node + jco 経由で読み込んで同じ 8 cases を回す。
// - Go 版: map[string]interface{} ↔ record
// - Node 版: 普通の JS オブジェクト ↔ record (jco が透過変換)
// - Go 版: []interface{} ↔ list
// - Node 版: 普通の配列 ↔ list
//
// jco が Component Model の値型を JS のネイティブ表現にマッピングする
// ので、Go 側で必要だった「型タグ付き map」が要らないのが対比のポイント。
import * as demo from "./dist/demo.js";

const cases = [
  // text patterns
  { name: "greet",  fn: () => demo.greet("world"),                 want: "Hello, world!" },
  { name: "greet",  fn: () => demo.greet("ダーリン"),               want: "Hello, ダーリン!" },
  { name: "upper",  fn: () => demo.upper("hello world"),           want: "HELLO WORLD" },
  { name: "words",  fn: () => demo.words("  the quick   brown fox  "), want: ["the", "quick", "brown", "fox"] },

  // struct patterns
  {
    name: "translate",
    fn: () => demo.translate({ x: 1, y: 2 }, 10, 20),
    want: { x: 11, y: 22 },
  },
  {
    name: "distance",
    fn: () => demo.distance({ x: 0, y: 0 }, { x: 3, y: 4 }),
    want: 5,
  },
  {
    name: "format-person",
    fn: () => demo.formatPerson({ name: "Alice", age: 30, tags: ["engineer", "wasm-fan"] }),
    want: "Alice (age 30) [engineer, wasm-fan]",
  },
  {
    name: "make-person",
    fn: () => demo.makePerson("Bob", 42, ["reviewer", "skeptic"]),
    want: { name: "Bob", age: 42, tags: ["reviewer", "skeptic"] },
  },
];

function deepEqual(a, b) {
  if (typeof a === "number" && typeof b === "number") {
    return Math.abs(a - b) < 1e-12 * Math.max(Math.abs(a), 1);
  }
  if (Array.isArray(a) && Array.isArray(b)) {
    return a.length === b.length && a.every((v, i) => deepEqual(v, b[i]));
  }
  if (a && b && typeof a === "object" && typeof b === "object") {
    const ak = Object.keys(a).sort();
    const bk = Object.keys(b).sort();
    if (ak.length !== bk.length) return false;
    return ak.every((k, i) => k === bk[i] && deepEqual(a[k], b[k]));
  }
  return a === b;
}

function summary(v) {
  const s = JSON.stringify(v);
  return s.length > 80 ? s.slice(0, 77) + "..." : s;
}

let failed = 0;
for (const c of cases) {
  try {
    const got = c.fn();
    if (!deepEqual(got, c.want)) {
      console.log(`  ✗ ${c.name}\n      got:  ${summary(got)}\n      want: ${summary(c.want)}`);
      failed++;
      continue;
    }
    console.log(`  ✓ ${c.name} = ${summary(got)}`);
  } catch (err) {
    console.log(`  ✗ ${c.name}: ${err}`);
    failed++;
  }
}

if (failed > 0) {
  console.error(`\n${failed} case(s) failed`);
  process.exit(1);
}
console.log(`\nAll ${cases.length} cases passed.`);
