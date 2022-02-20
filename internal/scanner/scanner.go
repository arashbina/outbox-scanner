package scanner

import (
	"encoding/json"
	"fmt"
	"github.com/arashbina/outbox-scanner/internal/models"
	"github.com/arashbina/outbox-scanner/internal/storage"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"time"
)

var db *storage.DB

type Scanner struct {
	storage  *storage.DB
	config   Config
	consumer *kafka.Consumer
	producer *kafka.Producer
}

type Config struct {
	ServiceID   string
	OutboxTopic string
	KafaUser    string
	KafkaSecret string
}

func New(s *storage.DB, cfg Config) (*Scanner, error) {

	scn := Scanner{
		storage: s,
		config:  cfg,
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "pkc-ymrq7.us-east-2.aws.confluent.cloud:9092",
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     cfg.KafaUser,
		"sasl.password":     cfg.KafkaSecret,
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	scn.consumer = c

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "pkc-ymrq7.us-east-2.aws.confluent.cloud:9092",
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     cfg.KafaUser,
		"sasl.password":     cfg.KafkaSecret,
	})

	if err != nil {
		panic(err)
	}

	scn.producer = p

	return &scn, nil
}

func (s *Scanner) Start(scanInterval time.Duration) {

	scanTicker := time.NewTicker(scanInterval * time.Second)
	go func() {
		for {
			select {
			case <-scanTicker.C:
				_ = scanEvent(s, models.SchemaResponse{})
				_ = scanEvent(s, models.OutboxRecord{})
			}
		}
	}()

	go s.consume()
}

func (s *Scanner) publishSchema(id string) error {

	sr, err := s.storage.GetSchema(id)
	if len(sr) != 1 {
		return fmt.Errorf("could not find the schema")
	}

	ss, err := json.Marshal(sr[0])
	if err != nil {
		return err
	}
	return s.produce(ss)
}

func scanEvent[T models.Event](scanner *Scanner, record T) error {
	tx, err := scanner.storage.DB.Beginx()
	if err != nil {
		return err
	}

	var query, update string
	switch record.GetEvent() {
	case models.OutboxEventType:
		query = "SELECT outbox.*, registry.schema FROM public.outbox INNER JOIN public.registry ON outbox.schema_id=registry.id WHERE outbox.published_at IS NULL;"
		update = "UPDATE public.outbox SET published_at = current_timestamp WHERE id=$1;"
	case models.SchemaEventType:
		query = "SELECT * FROM public.registry WHERE registry.published_at IS NULL;"
		update = "UPDATE public.registry SET published_at = current_timestamp WHERE id=$1;"
	}
	res, err := tx.Queryx(query)
	if err != nil {
		return err
	}

	var rows []T
	for res.Next() {
		var r T
		err = res.StructScan(&r)
		if err != nil {
			return err
		}
		err = r.Verify()
		if err != nil {
			return err
		}
		rows = append(rows, r)
	}
	res.Close()

	for _, row := range rows {
		r, err := json.Marshal(row)
		if err != nil {
			log.Printf("error marshaling event: %sl", err)
			continue
		}
		err = scanner.produce(r)
		if err != nil {
			log.Printf("error publishing event: %sl", err)
			continue
		}
		_, err = tx.Exec(update, row.GetID())
		if err != nil {
			log.Printf("error marking event as published: %s", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Scanner) consume() {

	err := s.consumer.SubscribeTopics([]string{s.config.OutboxTopic, s.config.ServiceID}, nil)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := s.consumer.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			if *msg.TopicPartition.Topic == s.config.ServiceID {
				var v models.SchemaRequest
				err = json.Unmarshal(msg.Value, &v)
				if err != nil {
					fmt.Printf("error unmarshalling the request: %s", err)
					continue
				}
				err = s.publishSchema(v.ID)
				if err != nil {
					fmt.Printf("error publishing event: %s", err)
					continue
				}
			}
		} else {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (s *Scanner) Shutdown() {
	s.producer.Close()
	s.consumer.Close()
}

func (s *Scanner) produce(payload []byte) error {

	go func() {
		for e := range s.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	err := s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.config.OutboxTopic, Partition: kafka.PartitionAny},
		Value:          payload,
	}, nil)

	if err != nil {
		return err
	}

	// Wait for message deliveries before shutting down
	s.producer.Flush(15 * 1000)
	return nil
}
