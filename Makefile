up: down
	docker compose up -d
down: compose.yml
	docker compose down
fmt:
	go fmt github.com/loilo-inc/exql/...
test:
	go test ./... -race -cover -coverprofile=coverage.out -covermode=atomic -count 1
README.md: template/README.md tool/**/*.go example/*.go
	go run tool/rdmegen/main.go
.PHONY: mocks
mocks:
	rm -rf mocks/
	go generate ./...
compose.yml: tool/composegen/*
	go run tool/composegen/main.go
