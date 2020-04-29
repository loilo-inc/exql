updb:
	docker-compose -f docker-compose.db.yml up -d
down:
	docker-compose -f docker-compose.db.yml down
fmt:
	go fmt github.com/loilo-inc/exql/...
test:
	go test -race -cover -coverprofile=coverage.out -covermode=atomic -count 1
README.md: template/README.md example/*.go
	go run tool/main.go
