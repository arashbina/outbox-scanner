package main

import (
	"github.com/arashbina/outbox-scanner/internal/scanner"
	"github.com/arashbina/outbox-scanner/internal/storage"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const ScanInterval = 2 //seconds
const ServiceID = "service_id"
const OutboxTopic = "OUTBOX_TEST"

func main() {

	userPostgres := os.Getenv("POSTGRES_USER")
	passPostgres := os.Getenv("POSTGRES_PASS")

	db, err := storage.New(storage.Config{
		User: userPostgres,
		Pass: passPostgres,
	})
	if err != nil {
		panic(err)
	}

	userKafka := os.Getenv("KAFKA_USER")
	passKafka := os.Getenv("KAFKA_SECRET")

	cfg := scanner.Config{
		ServiceID:   ServiceID,
		OutboxTopic: OutboxTopic,
		KafaUser:    userKafka,
		KafkaSecret: passKafka,
	}
	scn, err := scanner.New(db, cfg)
	if err != nil {
		panic(err)
	}

	scn.Start(ScanInterval)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-shutdown:
		if sig == syscall.SIGSTOP {
			log.Fatalln("sig stop shutdown")
		}
		scn.Shutdown()
	}
}
