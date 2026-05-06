.PHONY: all build guest host run clean

all: build

build: guest host

guest:
	cd guest && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../guest.wasm .

host: guest.wasm
	go build -o wasmtime-cli .

guest.wasm:
	$(MAKE) guest

run: build
	./wasmtime-cli add 3 5
	./wasmtime-cli mul 6 7

clean:
	rm -f wasmtime-cli guest.wasm
