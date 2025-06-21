#! /bin/bash

echo "Starting backup"
sudo chown -R root:root ./backups
make backup


echo "Upload backup"
make upload-backup
