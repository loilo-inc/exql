MOCKGEN := go run go.uber.org/mock/mockgen@v0.6.0
up: down
	docker compose up -d
down:
	docker compose down
fmt:
	go fmt github.com/loilo-inc/exql/...
.PHONY: test
test:
	go test ./... -race -cover -coverprofile=coverage.out -covermode=atomic -count 1
README.md: template/README.md tool/**/*.go example/*.go
	go run tool/rdmegen/main.go
.PHONY: mocks
mocks:
	rm -rf mocks/
	$(MOCKGEN) -source iface/iface.go -destination ./mocks/mock_iface/iface.go -package mock_iface
	$(MOCKGEN) -source query/query.go -destination ./mocks/mock_query/query.go -package mock_query
