.PHONY: all build guest transpile install run clean tools-check

GUEST_WASM := guest/target/wasm32-wasip2/release/html_fetcher.wasm

all: build

tools-check:
	@command -v cargo >/dev/null  || { echo "cargo not found"; exit 1; }
	@command -v node >/dev/null   || { echo "node not found"; exit 1; }
	@cargo component --version >/dev/null 2>&1 || { echo "cargo-component not installed: cargo install cargo-component"; exit 1; }
	@command -v wkg >/dev/null    || { echo "wkg not found: cargo install wkg"; exit 1; }
	@rustup target list --installed | grep -q wasm32-wasip2 || { echo "wasm32-wasip2 not installed: rustup target add wasm32-wasip2"; exit 1; }

build: install transpile

guest: tools-check
	cd guest && [ -d wit/deps ] || wkg wit fetch
	cd guest && cargo component build --release --target wasm32-wasip2

transpile: guest
	cd host && rm -rf dist && npx -y -- @bytecodealliance/jco transpile ../$(GUEST_WASM) -o dist --name fetcher

install:
	cd host && npm install

run: build
	cd host && node main.js https://example.com

clean:
	rm -rf guest/target guest/wit/deps guest/wkg.lock host/dist host/node_modules host/package-lock.json
