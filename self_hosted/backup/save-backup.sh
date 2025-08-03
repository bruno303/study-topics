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

DATE=$(date +%Y-%m-%d)
ARCHIVE_NAME="./tmp/backup-$DATE.tar.gz"
BACKUP_CONTEXT_DIR=$(pwd)

logDebug "Temp file: $ARCHIVE_NAME"
logDebug "Restic repo: $RESTIC_REPOSITORY"
logDebug "Backup destination: $BACKUP_DEST_DIR"

# Create tar of current repo
tar -czf "$ARCHIVE_NAME" -C "$RESTIC_REPOSITORY" .

if [ ! -d "$BACKUP_DEST_DIR" ]; then
  echo "$(date +"%Y-%m-%d %H:%M:%S") Backup destination directory not found: $BACKUP_DEST_DIR" 1>&2
  exit 1
fi

# Copy to flash drive
rclone copy "$ARCHIVE_NAME" "$BACKUP_DEST_DIR" --progress > /dev/null 2>&1

# Keep only the last backup on the flash drive
cd "$BACKUP_DEST_DIR"
ls -1t backup-*.tar.gz | tail -n +2 | xargs -r rm --

# Keep only the last 3 backups on tmp folder
cd "$BACKUP_CONTEXT_DIR"
ls -1t ./tmp/backup-*.tar.gz | tail -n +4 | xargs -r rm --
