GO ?= go
GOCACHE := $(CURDIR)/.gocache
GOTMPDIR := $(CURDIR)/.tmp
BIN_DIR := $(CURDIR)/.bin
APP := $(BIN_DIR)/watch_bot

.PHONY: build test run clean

build:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)" "$(BIN_DIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) build -o "$(APP)" .

test:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) test ./...

run:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) run .

clean:
	rm -rf "$(GOCACHE)" "$(GOTMPDIR)" "$(BIN_DIR)"
