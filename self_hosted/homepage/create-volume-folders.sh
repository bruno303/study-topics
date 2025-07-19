#!/bin/bash

set -a
. .env
set +a

echo "Creating volume folders on $VOLUMES_DIR"

mkdir -p "$VOLUMES_DIR/app/config"
mkdir -p "$VOLUMES_DIR/app/public/icons"

echo "✅ Done"
