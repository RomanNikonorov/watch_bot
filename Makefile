GO ?= go
GOCACHE := $(CURDIR)/.gocache
GOTMPDIR := $(CURDIR)/.tmp
BIN_DIR := $(CURDIR)/.bin
APP := $(BIN_DIR)/watch_bot

.PHONY: build test vet race integration-test migrate run clean

build:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)" "$(BIN_DIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) build -o "$(APP)" .

test:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) test ./...

vet:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) vet ./...

race:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) test -race ./...

integration-test:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) test -tags=integration ./dao

migrate:
	@test -n "$(CONNECTION_STR)" || (echo "CONNECTION_STR is required" && exit 1)
	@for migration in migrations/*.sql; do \
		echo "Applying $$migration"; \
		psql "$(CONNECTION_STR)" -v ON_ERROR_STOP=1 -f "$$migration"; \
	done

run:
	mkdir -p "$(GOCACHE)" "$(GOTMPDIR)"
	GOCACHE="$(GOCACHE)" GOTMPDIR="$(GOTMPDIR)" $(GO) run .

clean:
	rm -rf "$(GOCACHE)" "$(GOTMPDIR)" "$(BIN_DIR)"
