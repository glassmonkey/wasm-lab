.PHONY: all probe-future probe-error probe-stream clean

all: probe-future probe-error probe-stream

probe-future:
	@echo "=== Probe 1: future<u32> ==="
	-cd probes/future-simple && cargo component build --release --target wasm32-wasip2 2>&1 | tail -5
	@echo

probe-error:
	@echo "=== Probe 2: error-context ==="
	-cd probes/error-context && cargo component build --release --target wasm32-wasip2 2>&1 | tail -8
	@echo

probe-stream:
	@echo "=== Probe 3: stream<u8> ==="
	-cd probes/stream-bytes && cargo component build --release --target wasm32-wasip2 2>&1 | tail -5
	@echo

clean:
	rm -rf probes/*/target probes/*/Cargo.lock probes/*/src/bindings.rs
