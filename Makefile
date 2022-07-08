
.PHONY: build
build:
	go build -o bin/schemagen ./cmd/tfschemagen

.PHONY: generate
generate:
	go run cmd/tfschemagen/main.go example
