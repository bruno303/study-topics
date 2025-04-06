OUT_FILE = bin/app

deps:
	go mod tidy
	go mod vendor

build:
	go build -o $(OUT_FILE) ./cmd

run: build
	./$(OUT_FILE)

debug: build
	./$(OUT_FILE) --debug

.PHONY: deps build
