FRONT_DIR=frontend/planning-poker-front

include .env
export

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
	docker run --rm -p 8080:8080 --env-file .env --name planning-poker-backend planning-poker-backend

docker-build-frontend:
	cd $(FRONT_DIR) && docker build -f Dockerfile \
		--build-arg NEXT_PUBLIC_BACKEND_URL=http://localhost:8080 \
  	--build-arg NEXT_PUBLIC_BACKEND_HOST=localhost:8080 \
		--no-cache -t planning-poker-frontend .

docker-run-frontend: docker-build-frontend
	cd $(FRONT_DIR) && docker run --rm -p 3000:3000 --name planning-poker-frontend planning-poker-frontend
