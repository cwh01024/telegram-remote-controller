.PHONY: build test run clean install

BINARY_NAME=telegram-remote-controller

build:
	go build -o $(BINARY_NAME) ./cmd/bot

test:
	go test ./... -v

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

install: build
	./scripts/install.sh

uninstall:
	./scripts/uninstall.sh
