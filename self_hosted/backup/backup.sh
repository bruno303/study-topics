#!/bin/bash
set -euo pipefail

DEBUG=0

function logDebug () {
  if [ $DEBUG -eq 1 ]; then
    echo "$(date +"%Y-%m-%d %H:%M:%S") $1"
  fi
}

set -a
. .env
set +a

logDebug "📦 Starting backup"
# Init repo if needed
if ! restic snapshots > /dev/null 2>&1; then
  logDebug "🔐 Initializing restic repository"
  restic init > /dev/null 2>&1
  logDebug "🔐 Restic repository initialized with success"
fi

# Backup each subfolder of /data
for SERVICE_DIR in "$DATA_DIR"/*; do
  [ -d "$SERVICE_DIR" ] || continue
  SERVICE=$(basename "$SERVICE_DIR")
  TAG="$SERVICE"
  logDebug "📂 Backing up $SERVICE_DIR (tag: $TAG)"
  restic backup "$SERVICE_DIR" --tag "$TAG" > /dev/null 2>&1

  # Cleanup old snapshots
  logDebug "🧹 Applying retention policy: $RETENTION_ARGS (tag: $TAG)"
  restic forget --tag "$TAG" $RETENTION_ARGS > /dev/null 2>&1
done

logDebug "✅ Backup complete"
