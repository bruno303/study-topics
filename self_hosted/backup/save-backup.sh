#!/bin/bash
set -euo pipefail

set -a
. .env
set +a

DATE=$(date +%Y-%m-%d)
ARCHIVE_NAME="./tmp/backup-$DATE.tar.gz"
BACKUP_CONTEXT_DIR=$(pwd)

echo "Temp file: $ARCHIVE_NAME"
echo "Restic repo: $RESTIC_REPOSITORY"
echo "Backup destination: $BACKUP_DEST_DIR"

# Create tar of current repo
tar -czf "$ARCHIVE_NAME" -C "$RESTIC_REPOSITORY" .

# Copy to flash drive
rclone copy "$ARCHIVE_NAME" "$BACKUP_DEST_DIR" --progress

# Keep only the last backup on the flash drive
cd "$BACKUP_DEST_DIR"
ls -1t backup-*.tar.gz | tail -n +2 | xargs -r rm --

# Clean temp archive
cd "$BACKUP_CONTEXT_DIR"
rm "$ARCHIVE_NAME"
