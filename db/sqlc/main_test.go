package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/the-eduardo/Go-Bank/util"
	"os"
	"testing"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal().Msgf("cannot load config: %v", err)
	}

	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %v", err)
	}
	defer testDB.Close()
	testQueries = New(testDB)

	os.Exit(m.Run())
}
