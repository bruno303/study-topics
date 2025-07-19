#!/bin/bash

set -a
. .env
set +a

echo "Creating volume folders on $VOLUMES_DIR"

mkdir -p "$VOLUMES_DIR/gitea/{data,config}"
mkdir -p "$VOLUMES_DIR/postgres/data"

echo "✅ Done"
