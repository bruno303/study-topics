GO_TOOLS_INSTALL=xargs -n1 -I {} go install {} < "tools/tools.txt"
JS_TOOLS_INSTALL=xargs -n1 -I {} npm install -g {} < "tools/tools-js.txt"
API_OUTPUT=./tmp/api

init:
	-rm .env 2> /dev/null
	cp .env.example .env

run-cli:
	go run cmd/cli/*.go

run-api:
	go run cmd/api/*.go

debug-api:
	go build -gcflags="all=-N -l" -o $(API_OUTPUT) cmd/api/*.go
	$(API_OUTPUT)

run-api-live:
	air -c .air.toml

deps:
	$(GO_TOOLS_INSTALL)
	$(JS_TOOLS_INSTALL)
	go mod vendor
	go mod tidy -e -v

docker-up-infra:
	docker compose up -d

docker-down:
	docker compose down

docker-up-app: docker-up-infra
	docker compose -f docker-compose.api.yml up --build --force-recreate

test:
	go test -timeout 30s ./...

test-watch:
	nodemon --watch ./ -e go,yml,yaml,json --delay 2 --exec "make test"

mocks:
	go generate -v ./...
