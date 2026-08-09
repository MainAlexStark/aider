BINARY=aider

.PHONY: build install uninstall clean run

build:
	go build -o $(BINARY) ./cmd/agent

install: build
	sudo install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	sudo rm -f /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)

run:
	go run ./cmd/agent