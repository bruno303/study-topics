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

# Backup each service
for SERVICE in $VOLUMES; do
  SERVICE_DIR="/data/$SERVICE"
  echo "📂 Backing up $SERVICE"

  # For each subfolder inside the service directory
  for SUBDIR in "$SERVICE_DIR"/*; do
    [ -d "$SUBDIR" ] || continue
    NAME=$(basename "$SUBDIR")
    TAG="${SERVICE}-${NAME}"
    echo "  🔸 Backing up $SUBDIR (tag: $TAG)"
    restic backup "$SUBDIR" --tag "$TAG"

    # Cleanup old snapshots
    echo "🧹 Applying retention policy: $RETENTION_ARGS (tag: $TAG)"
    restic forget --tag "$TAG" $RETENTION_ARGS
  done
done

echo "✅ Backup complete"
