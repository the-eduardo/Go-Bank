DB_URL=postgresql://root:secret@localhost:5432/gobank_db?sslmode=disable

postgres:
	docker run --name gobank_postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it gobank_postgres createdb --username=root --owner=root gobank_db

dropdb:
	docker exec -it gobank_postgres dropdb gobank_db

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

dbdocs:
	dbdocs build .\doc\db.dbml

dbschema:
	dbml2sql --postgres -o doc/gobankschema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go github.com/the-eduardo/Go-Bank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/the-eduardo/Go-Bank/worker TaskDistributor
# mockery is deprecated, use mock instead
mockery:
	mockery --config=.mockery.yaml

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    --openapiv2_out=doc/swagger --openapiv2_opt=allow_merge,merge_file_name=gobank \
	proto/*.proto

evans:
	evans --port 9090 -r repl
# For server building only
dockerbuild:
	docker run --name gobank-main --network bank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@gobank_postgres:5432/gobank_db?sslmode=disable" gobank:latest

redis:
	docker run --name gobank_redis --network bank-network -p 6379:6379 -d redis:7.4-rc2-alpine

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock mockery dockerbuild dbdocs dbschema proto evans redis new_migration
