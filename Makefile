updb:
	docker-compose -f docker-compose.db.yml up -d
down:
	docker-compose -f docker-compose.db.yml down
fmt:
	go fmt github.com/loilo-inc/exql/...
test:
	go test -race -cover -coverprofile=coverage.out -covermode=atomic \
	github.com/loilo-inc/exql/... -count 1