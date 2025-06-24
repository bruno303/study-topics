#!/bin/bash
set -euo pipefail

echo "📦 Starting backup"

# Load env vars
export RESTIC_PASSWORD
export RESTIC_REPOSITORY=$RESTIC_REPO

# Init repo if needed
if ! restic snapshots > /dev/null 2>&1; then
  echo "🔐 Initializing restic repository"
  restic init
fi

# Backup each subfolder of /data
for SERVICE_DIR in /data/*; do
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
