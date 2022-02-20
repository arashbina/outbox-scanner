# Outbox Scanner

Outbox Scanner is a scanner for the data streaming outbox pattern. Scanner deals
with two separate tables:
1. Schema Table
2. Outbox Table

Any new record in the Schema table will count as a new schema and will be published
by the scanner automatically. Scanner can also respond to requests for a schema
by id through Kafka. The requested schema is published back on the Kafka with appropriate
topic. 

Any records on the Outbox table must have a corresponding schema_id in the Schema table.
This is constrained by a foreign key. Scanner scans for new events stored in the 
outbox table and publishes them to Kafka and updates the records with a `published_at` date.

The scanner currently performs the following tasks:

- Creates necessary Schema and Outbox tables if they don't exist
- Scans the Schema table periodically and sends out new schemas
- Responds to schema requests by consumers
- Scans the outbox table for new events
- Publishes new events and marks them as published

The following functionality is planned for the scanner:
- to verify new schemas have the correct compatibility with older ones
- to verify that the messages conform to the referencing schema
- a leader/follower functionality using Kubernetes API or a locking mechanism using a database to ensure messages are only published once (for performance considerations)



