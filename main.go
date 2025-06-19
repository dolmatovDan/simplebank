package main

import (
	"context"
	"log"

	"github.com/dolmatovDan/simplebank/api"
	db "github.com/dolmatovDan/simplebank/db/sqlc"
	"github.com/dolmatovDan/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *db.Queries
var conn *pgxpool.Pool

func main() {
	var err error

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
		return
	}

	conn, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
		return
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
		return
	}
}
