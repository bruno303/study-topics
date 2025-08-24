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
ARCHIVE_NAME="./tmp/$DATE/backup-$DATE.tar.gz"
ARCHIVE_SPLIT_PREFIX="./tmp/$DATE/backup-$DATE.tar.gz.part-"
BACKUP_CONTEXT_DIR=$(pwd)

logDebug "Temp file: $ARCHIVE_NAME"
logDebug "Restic repo: $RESTIC_REPOSITORY"
logDebug "Backup destination: $BACKUP_DEST_DIR"

# Create tar of current repo
if [ ! -d "./tmp/$DATE" ]; then
  mkdir -p "./tmp/$DATE"
fi

if [ $DEBUG -eq 1 ]; then
  tar -czvf "$ARCHIVE_NAME" -C "$RESTIC_REPOSITORY" .
else
  tar -czf "$ARCHIVE_NAME" -C "$RESTIC_REPOSITORY" .
fi

if [ ! -d "$BACKUP_DEST_DIR" ]; then
  echo "$(date +"%Y-%m-%d %H:%M:%S") Backup destination directory not found: $BACKUP_DEST_DIR" 1>&2
  exit 1
fi

if [ ! -d "$BACKUP_DEST_DIR/$DATE" ]; then
  mkdir -p "$BACKUP_DEST_DIR/$DATE"
fi
BACKUP_DEST_DIR="$BACKUP_DEST_DIR/$DATE"

# Split into 2GB chunks if needed
logDebug "Splitting archive into 2GB parts (FAT32 safe)"
split -b 2000M "$ARCHIVE_NAME" "$ARCHIVE_SPLIT_PREFIX"

# Remove original big file if split was successful
if [ -f "${ARCHIVE_SPLIT_PREFIX}aa" ]; then
  rm "$ARCHIVE_NAME"
fi

# Copy to backup destination
logDebug "Copying backup to flash drive"
cp "$ARCHIVE_SPLIT_PREFIX"* "$BACKUP_DEST_DIR"

# Keep only the last backup on the flash drive
logDebug "Cleaning up old backups on flash drive"
cd "$BACKUP_DEST_DIR/.."
ls -1dt */ | tail -n +2 | xargs -r rm -rf --

# Keep only the last 3 backups on tmp folder
logDebug "Cleaning up old backups on tmp folder"
cd "$BACKUP_CONTEXT_DIR/tmp"
ls -1dt */ | tail -n +4 | xargs -r rm -rf --

logDebug "✅ Backup saved"
