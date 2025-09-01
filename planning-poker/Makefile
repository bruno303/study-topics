FRONT_DIR=frontend/planning-poker-front

.PHONY: init
init:
	go mod tidy
	go mod download
	cd $(FRONT_DIR) && npm i

.PHONY: run
run:
	go run ./cmd/api

.PHONY: run-frontend
run-frontend:
	cd $(FRONT_DIR) && npm run dev

.PHONY: build
build:
	go build -o bin/api ./cmd/api

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin

.PHONY: fmt
fmt:
	go fmt ./...

docker-build-backend:
	docker build --no-cache -t planning-poker-backend .

docker-run-backend: docker-build-backend
	docker run --rm --network=bridge -p 8080:8080 planning-poker-backend

docker-build-frontend:
	cd $(FRONT_DIR) && docker build -f Dockerfile --no-cache -t planning-poker-front .

docker-run-frontend: docker-build-frontend
	cd $(FRONT_DIR) && docker run --rm --network=bridge -p 3000:3000 planning-poker-front .
