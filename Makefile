BINDIR ?= ./bin
INSTALL_DIR ?= /usr/bin

BUILDNR ?= $(shell git describe --tags --always --dirty 2>/dev/null || git rev-parse --short HEAD)

GO := go
PKG := ./cmd/mon

.PHONY: build install run test vet fmt clean

build:
	@mkdir -p "$(BINDIR)"
	$(GO) build -o "$(BINDIR)/mon-$(BUILDNR)" $(PKG)
	@cp -f "$(BINDIR)/mon-$(BUILDNR)" "$(BINDIR)/mon"

run: build
	"$(BINDIR)/mon"

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

clean:
	@rm -rf "$(BINDIR)"

install: build
	@sudo mkdir -p "$(INSTALL_DIR)"; \
	echo "Installing $(BINDIR)/mon to $(INSTALL_DIR)/mon"; \
	sudo cp -f "$(BINDIR)/mon" "$(INSTALL_DIR)/mon"; \
	sudo chmod 0755 "$(INSTALL_DIR)/mon"
