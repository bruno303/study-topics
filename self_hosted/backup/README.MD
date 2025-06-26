# Backup

This project aims to guarantee that all important data from the homelab is backed up.

## Generating backup

The backup is done by [restic](https://github.com/restic/restic). All the necessary scripts are documented on [Makefile](./Makefile).
This script will save a backup of all subfolders of the data folder defined in `DATA_DIR`, keeping only the last 3 backups of each volume.

## Uploading

The upload of the backup is done by [rclone](https://rclone.org/). The destination can be flash drives, folders or even cloud providers.

## Scheduling

For now I rely on crontab to run it once per week.

Crontab configuration example:

```txt
0 3 * * 0 /path/to/backup-script.sh >> /path/to/backup-log 2>&1
```
