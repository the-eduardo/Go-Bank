postgres:
	docker run --name postgreas -p 54321:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres
createdb:
	docker exec -it postgres1 createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres1 dropdb simple_bank
migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
sqlc:
	sqlc generate
sqlcdocker:
	docker run --rm -v ${pwd}:/src -w /src kjconroy/sqlc generate
test:
	go test -v -cover ./...
server:
	go run main.go
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/the-eduardo/Go-Bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc sqlcdocker test server mock

