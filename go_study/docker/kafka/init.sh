#!/bin/sh

# blocks until kafka is reachable
kafka-topics --bootstrap-server kafka:29092 --list
echo 'Creating kafka topics'
kafka-topics --bootstrap-server kafka:29092 --create --if-not-exists --topic "go-study.hello" --replication-factor 1 --partitions 10

echo 'Successfully created the following topics:'
kafka-topics --bootstrap-server kafka:29092 --list
