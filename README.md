# outbox-scanner

Outbox Scanner is a scanner for the data streaming outbox patter.

The scanner performs the following tasks:

- Creates necessary Schema and Outbox tables if they don't exist
- Scans the Schema table periodically and sends out new schemas
- Responds to schema requests by consumers
- Scans the outbox table for new events
- Publishes new events and marks them as published
- Verifies that the messages conform to the referencing schema

