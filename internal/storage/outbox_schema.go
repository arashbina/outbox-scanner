package storage

var schema = `
CREATE TABLE IF NOT EXISTS registry (
    id text PRIMARY KEY UNIQUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at timestamp, 
    schema text
);
    
CREATE TABLE IF NOT EXISTS outbox (
    id uuid PRIMARY KEY UNIQUE,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at timestamp,
    publisher_id text,
    schema_id text NOT NULL,
    entity text,
    entity_id uuid,
    action text,
    payload text,
    consistency_id SERIAL,
    FOREIGN KEY (schema_id) REFERENCES registry(id)
);
`
