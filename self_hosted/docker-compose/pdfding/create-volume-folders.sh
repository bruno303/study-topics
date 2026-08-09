#!/bin/bash

set -a
. .env
set +a

echo "Creating volume folders on $VOLUMES_DIR"

mkdir -p $VOLUMES_DIR/{sqlite_data,media}

echo "✅ Done"
