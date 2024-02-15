postgres:
	docker run --name gobank_postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it gobank_postgres createdb --username=root --owner=root gobank_db

dropdb:
	docker exec -it gobank_postgres dropdb gobank_db

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server
