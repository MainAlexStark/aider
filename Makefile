BINARY=aider
INSTALL_DIR=/usr/local/bin

.PHONY: build install uninstall clean

build:
	go build -o $(BINARY) ./cmd/agent

install: build
	sudo install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

uninstall:
	sudo rm -f $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)