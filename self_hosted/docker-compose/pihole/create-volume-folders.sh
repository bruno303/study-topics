#!/bin/bash

set -a
. .env
set +a

echo "Creating volume folders on $VOLUMES_DIR"

mkdir -p "$VOLUMES_DIR/etc-pihole"

echo "✅ Done"
