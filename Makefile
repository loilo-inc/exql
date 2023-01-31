up:down
	docker-compose -f docker-compose.db.yml up -d
down:
	docker-compose -f docker-compose.db.yml down
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
