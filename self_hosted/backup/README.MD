# Backup

This project aims to guarantee that all important data from the homelab is backed up.

## Generating backup

The backup is done by restic that is executed inside a container. The container executed the [backup.sh](./backup.sh) script.
This script will check all folders, if they are defined in `VOLUME` env var, all its volumes will be backed up, keeping only the last 3 backups of each volume.

## Uploading

The upload of the backup is done by [rclone](https://rclone.org/) that will upload the backup to google drive.

## Scheduling

For now I rely on crontab to run it once per week.

Crontab configuration example:

```txt
0 3 * * 0 /path/to/backup-script.sh >> /path/to/backup-log 2>&1
```
