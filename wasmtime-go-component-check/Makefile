.PHONY: all build guest host run clean tools-check

GUEST_WASM := guest/target/wasm32-wasip2/release/calculator.wasm

all: run

tools-check:
	@command -v cargo >/dev/null  || { echo "cargo not found"; exit 1; }
	@cargo component --version >/dev/null 2>&1 || { echo "cargo-component not installed: cargo install cargo-component"; exit 1; }
	@command -v wkg >/dev/null    || { echo "wkg not found: cargo install wkg"; exit 1; }
	@rustup target list --installed | grep -q wasm32-wasip2 || { echo "wasm32-wasip2 not installed: rustup target add wasm32-wasip2"; exit 1; }
	@test -d ../../wasmtime-go || { echo "expected ../../wasmtime-go (glassmonkey/wasmtime-go fork) — see README.md"; exit 1; }

build: tools-check $(GUEST_WASM) host

guest: tools-check
	cd guest && [ -d wit/deps ] || wkg wit fetch
	cd guest && cargo component build --release --target wasm32-wasip2

$(GUEST_WASM): guest

host: $(GUEST_WASM)
	go build -o check .

run: build
	./check

clean:
	rm -rf check guest/target guest/wit/deps guest/wkg.lock guest/Cargo.lock
