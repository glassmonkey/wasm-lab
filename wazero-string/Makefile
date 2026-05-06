.PHONY: all build guest host run clean

all: build

build: guest host

guest:
	cd guest && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../guest.wasm .

host: guest.wasm
	go build -o wazero-string .

guest.wasm:
	$(MAKE) guest

run: build
	./wazero-string greet world
	./wazero-string greet ダーリン

clean:
	rm -f wazero-string guest.wasm
