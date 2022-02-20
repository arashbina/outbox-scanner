package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/linkedin/goavro/v2"
	"log"
	"time"
)

type EventType int

const (
	SchemaEventType EventType = iota
	OutboxEventType
)

type SchemaRequest struct {
	ID string
}

type SchemaResponse struct {
	ID          string
	CreatedAt   *time.Time `db:"created_at"`
	PublishedAt *time.Time `db:"published_at"`
	Schema      string
}

func (s SchemaResponse) GetEvent() EventType {
	return SchemaEventType
}

func (s SchemaResponse) GetID() string {
	return s.ID
}

func (s SchemaResponse) Verify() error {
	return nil
}

type OutboxRecord struct {
	ID            uuid.UUID
	CreatedAt     *time.Time `db:"created_at"`
	PublishedAt   *time.Time `db:"published_at"`
	PublisherID   string     `db:"publisher_id"`
	SchemaID      string     `db:"schema_id"`
	Entity        string
	EntityID      string `db:"entity_id"`
	Action        string
	Payload       []byte
	ConsistencyID int    `db:"consistency_id"`
	Schema        string `json:"-"`
}

func (o OutboxRecord) GetEvent() EventType {
	return OutboxEventType
}

func (o OutboxRecord) GetID() string {
	return o.ID.String()
}

func (o OutboxRecord) Verify() error {

	codec, err := goavro.NewCodec(o.Schema)
	if err != nil {
		return err
	}

	var format map[string]interface{}
	err = json.Unmarshal([]byte(codec.CanonicalSchema()), &format)
	if err != nil {
		log.Println("could not get avro object")
		return err
	}
	log.Println(format)
	return nil
}

type Event interface {
	OutboxRecord | SchemaResponse
	GetEvent() EventType
	GetID() string
	Verify() error
}
