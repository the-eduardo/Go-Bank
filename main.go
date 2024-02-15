package main

import (
	"GoBank/api"
	db "GoBank/db/sqlc"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

const (
	dbSource      = "postgresql://root:secret@localhost:5432/gobank_db"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
