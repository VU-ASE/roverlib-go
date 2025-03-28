# Find more information at ase.vu.nl/docs/framework/glossary/makefiles
.PHONY: build start clean test

BUILD_DIR=bin/
BINARY_NAME=roverlib

lint:
	@echo "Lint check..."
	@golangci-lint run

build: lint
	@echo "You cannot build a library :("

clean:
	@echo "Cleaning all targets for ${BINARY_NAME}"
	rm -rf $(BUILD_DIR)

test: lint
	go test ./src -v -count=1 -timeout 0
