BUN ?= bun
BIN := bin/server

.PHONY: all web server test vet check clean

all: web server

web:
	cd web && $(BUN) install --frozen-lockfile && $(BUN) run build

server:
	go build -o $(BIN) ./cmd/server

check:
	cd web && $(BUN) run check

vet:
	@test -z "$$(gofmt -l .)" || (echo "not gofmt'd:"; gofmt -l .; exit 1)
	go vet ./...

test: vet
	go test ./...

clean:
	rm -rf $(BIN)
