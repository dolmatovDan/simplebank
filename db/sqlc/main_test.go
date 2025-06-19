package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/dolmatovDan/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	var err error

	config, err := util.LoadConfig("../../")
	if err != nil  {
		log.Fatal("cannot load config: ", err)
		return
	}

	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
		return
	}
	testQueries = New(testDB)

	os.Exit(m.Run())
}
