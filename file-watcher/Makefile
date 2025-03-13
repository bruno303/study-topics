OUTPUT_FILE=bin/file-watcher

build:
	@go build -o $(OUTPUT_FILE) cmd/*.go

run: build
	@$(OUTPUT_FILE) --files "example/*.sh" --delay 3 --signal SIGINT --command "echo ""file changed!""" --command "./example/test.sh"

deps:
	@go mod vendor
	@go mod tidy

.PHONY: build run deps
