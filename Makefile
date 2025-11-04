.PHONY: all vet lint test

all: vet lint test

vet:
	@go vet ./...
	@echo "✓ vet"

lint:
	@golangci-lint run ./...
	@echo "✓ lint"

test:
	@go test -race ./...
	@echo "✓ test"
