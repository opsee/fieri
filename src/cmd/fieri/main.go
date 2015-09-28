package main

import (
	kvlog "github.com/go-kit/kit/log"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/fieri/consumer"
	"github.com/opsee/fieri/onboarder"
	"github.com/opsee/fieri/service"
	"github.com/opsee/fieri/store"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func main() {
	kvlogger := kvlog.NewLogfmtLogger(os.Stdout)

	pgConnection := os.Getenv("POSTGRES_CONN")
	if pgConnection == "" {
		log.Fatal("You have to give me a postgres connection by setting the POSTGRES_CONN env var")
	}

	db, err := store.NewPostgres(pgConnection)
	if err != nil {
		log.Fatal("Error initializing postgres:", err)
	}

	lookupdHosts := os.Getenv("LOOKUPD_HOSTS")
	if lookupdHosts == "" {
		log.Fatal("You'll need to give me a nsqlookupd connection(s) by setting the LOOKUPD_HOSTS env var (comma-separated)")
	}

	concurrency, err := strconv.Atoi(os.Getenv("FIERI_CONCURRENCY"))
	if err != nil {
		log.Println("WARNING: FIERI_CONCURRENCY was not set properly, so defaulting to 1")
		concurrency = 1
	}

	bastionDiscoveryTopic := os.Getenv("BASTION_DISCOVERY_TOPIC")
	if bastionDiscoveryTopic == "" {
		log.Fatal("You have to give me a topic to consume by setting the BASTION_DISCOVERY_TOPIC env var")
	}

	lookupds := strings.Split(lookupdHosts, ",")
	nsqConsumer, err := consumer.NewNsq(lookupds, db, kvlogger, concurrency, bastionDiscoveryTopic)
	if err != nil {
		log.Fatal("Error initializing nsq consumer:", err)
	}

	nsqdHost := os.Getenv("NSQD_HOST")
	if nsqdHost == "" {
		log.Fatal("Error connecting to nsqd, you need to set NSQD_HOST")
	}

	producer, err := nsq.NewProducer(nsqdHost, nsq.NewConfig())
	if err != nil {
		log.Fatal("Error connecting to nsqd: ", err)
	}

	onboardTopic := os.Getenv("FIERI_ONBOARDING_TOPIC")
	if onboardTopic == "" {
		log.Fatal("You have to give me a topic to publish onboarding by setting the FIERI_ONBOARDING_TOPIC env var")
	}
	onboarder := onboarder.NewOnboarder(db, producer, kvlogger, onboardTopic)

	addr := os.Getenv("FIERI_HTTP_ADDR")
	if addr == "" {
		log.Fatal("You have to give me a listening address by setting the FIERI_HTTP_ADDR env var")
	}

	service := service.NewService(db, onboarder, kvlogger)
	service.StartHTTP(addr)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	<-interrupt

	log.Fatal(nsqConsumer.Stop())
}
