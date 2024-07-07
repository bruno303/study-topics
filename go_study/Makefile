GO_TOOLS_INSTALL=xargs -n1 -I {} go install {} < "tools/tools.txt"

init:
	-rm .env 2> /dev/null
	cp .env.example .env

run-cli:
	go run cmd/cli/*.go

run-api:
	go run cmd/api/*.go

run-api-live:
	air -c .air.toml

deps:
	$(GO_TOOLS_INSTALL)
	go mod vendor
	go mod tidy -e -v

docker-build:
	docker build -t go-study:1.0.0 .

docker-up-infra:
	docker compose --profile local up -d

docker-down:
	docker compose --profile local down

docker-up-app:
	docker compose --profile app up --build --force-recreate
