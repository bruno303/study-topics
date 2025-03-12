#!/bin/bash

trap "echo 'Script terminated by signal'; exit" SIGINT SIGTERM

while [ true ]; do
  sleep 1
  echo "running..."
done
