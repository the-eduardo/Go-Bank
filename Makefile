postgres:
	docker run --name gobank_postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it gobank_postgres createdb --username=root --owner=root gobank_db

dropdb:
	docker exec -it gobank_postgres dropdb gobank_db

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go github.com/the-eduardo/Go-Bank/db/sqlc Store
# mockery is deprecated, use mock instead
mockery:
	mockery --config=.mockery.yaml

# For server building only
dockerbuild:
	docker run --name gobank-main --network bank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@gobank_postgres:5432/gobank_db?sslmode=disable" gobank:latest

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock mockery dockerbuild
