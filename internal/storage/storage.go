package storage

import (
	"fmt"
	"github.com/arashbina/outbox-scanner/internal/models"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

type Config struct {
	User string
	Pass string
}

func New(cfg Config) (*DB, error) {
	str := fmt.Sprintf("user=%s password=%s dbname=outbox sslmode=disable", cfg.User, cfg.Pass)
	db, err := sqlx.Connect("postgres", str)
	if err != nil {
		return nil, err
	}

	d := DB{DB: db}
	err = d.createOutboxTables()
	return &d, nil
}

func (d *DB) createOutboxTables() error {
	_, err := d.Exec(schema)
	return err
}

func (d *DB) GetSchema(id string) ([]models.SchemaResponse, error) {
	query := "SELECT * FROM public.registry WHERE id=$1"
	var sr []models.SchemaResponse
	err := d.Select(&sr, query, id)
	if err != nil {
		return nil, fmt.Errorf("error selecting: %s", err)
	}

	return sr, nil
}
