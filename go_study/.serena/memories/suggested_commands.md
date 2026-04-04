Setup: make init; make download; make vendor
Run: make run-cli; make run-api; make debug-api; make run-api-live
Docker/local infra: make docker-up-infra; make docker-down; make docker-up-app
Tests: make test; make test-watch
Focused tests: go test ./internal/application/hello -run '^TestHello$' -count=1 ; go test ./internal/infra/repository -run 'Rollback' -count=1 ; go test ./pkg/utils/array -count=1 ; go test -v ./internal/infra/worker -count=1
Quality checks: gofmt -w <changed-files>; go vet ./...; go test ./...
Codegen/mocks: make mocks (go generate -v ./...)
Git basics (Linux): git status; git diff; git add <files>; git commit -m "msg"; git branch --show-current