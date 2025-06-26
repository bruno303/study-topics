#!/bin/bash
set -euo pipefail

set -a
. .env
set +a

echo "📦 Starting backup"
# Init repo if needed
if ! restic snapshots > /dev/null 2>&1; then
  echo "🔐 Initializing restic repository"
  restic init
fi

# Backup each subfolder of /data
for SERVICE_DIR in "$DATA_DIR"/*; do
  [ -d "$SERVICE_DIR" ] || continue
  SERVICE=$(basename "$SERVICE_DIR")
  TAG="$SERVICE"
  echo "📂 Backing up $SERVICE_DIR (tag: $TAG)"
  restic backup "$SERVICE_DIR" --tag "$TAG"

  # Cleanup old snapshots
  echo "🧹 Applying retention policy: $RETENTION_ARGS (tag: $TAG)"
  restic forget --tag "$TAG" $RETENTION_ARGS
done

echo "✅ Backup complete"
