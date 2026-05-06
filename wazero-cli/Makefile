.PHONY: all build guest host run clean

all: build

build: guest host

guest:
	cd guest && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../guest.wasm .

host: guest.wasm
	go build -o wazero-cli .

guest.wasm:
	$(MAKE) guest

run: build
	./wazero-cli add 3 5
	./wazero-cli mul 6 7

clean:
	rm -f wazero-cli guest.wasm
