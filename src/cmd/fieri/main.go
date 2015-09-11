package main

import (
	"github.com/opsee/fieri/consumer"
	"github.com/opsee/fieri/store"
	"log"
	"os"
	"os/signal"
	"strings"
)

func main() {
	logger := log.New(os.Stdout, "[fieri] ", log.Lshortfile|log.LstdFlags)

	pgConnection := os.Getenv("POSTGRES_CONN")
	if pgConnection == "" {
		logger.Fatal("You have to give me a postgres connection by setting the POSTGRES_CONN env var")
	}

	db, err := store.NewPostgres(pgConnection)
	if err != nil {
		logger.Fatal("Error initializing postgres:", err)
	}

	lookupdHosts := os.Getenv("LOOKUPD_HOSTS")
	if lookupdHosts == "" {
		logger.Fatal("You'll need to give me a nsqlookupd connection(s) by setting the LOOKUPD_HOSTS env var (comma-separated)")
	}

	lookupds := strings.Split(lookupdHosts, ",")
	nsq, err := consumer.NewNsq(lookupds, db, logger)
	if err != nil {
		logger.Fatal("Error initializing nsq consumer:", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	<-interrupt

	if err = nsq.Stop(); err != nil {
		logger.Fatal(err)
	}
}
