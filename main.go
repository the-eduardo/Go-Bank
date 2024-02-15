package main

import (
	"GoBank/api"
	db "GoBank/db/sqlc"
	"GoBank/util"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func main() {
	// Load the configuration
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
