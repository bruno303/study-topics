MAKEFLAGS += --silent

include .env
export

RESTIC_REPOSITORY = /backups/repo
RESTIC_PASSWORD ?= supersecret

$(if $(HOST_ROOT),,$(error HOST_ROOT variable is missing))

list-snapshots:
	docker run --rm \
		-e RESTIC_REPOSITORY=$(RESTIC_REPOSITORY) \
		-e RESTIC_PASSWORD=$(RESTIC_PASSWORD) \
		-v $(HOST_ROOT)/backup/backups:/backups \
		restic/restic snapshots

browse-files:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: browse-files SNAPSHOT_ID=foo" && exit 1)
	docker run --rm \
		-e RESTIC_REPOSITORY=$(RESTIC_REPOSITORY) \
		-e RESTIC_PASSWORD=$(RESTIC_PASSWORD) \
		-v $(HOST_ROOT)/backup/backups:/backups \
		restic/restic ls $(SNAPSHOT_ID)

restore:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: restore SNAPSHOT_ID=foo" && exit 1)
	docker run --rm \
		-e RESTIC_REPOSITORY=$(RESTIC_REPOSITORY) \
		-e RESTIC_PASSWORD=$(RESTIC_PASSWORD) \
		-v $(HOST_ROOT)/backup/backups:/backups \
		-v $(HOST_ROOT)/backup/restore:/restore \
		restic/restic restore $(SNAPSHOT_ID) --target /restore

restore-path:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: restore-path SNAPSHOT_ID=foo" && exit 1)
	@test -n "$(PATH)" || (echo "❌ PATH variable is missing. Use: restore-path PATH=foo" && exit 1)
	docker run --rm \
		-e RESTIC_REPOSITORY=$(RESTIC_REPOSITORY) \
		-e RESTIC_PASSWORD=$(RESTIC_PASSWORD) \
		-v $(HOST_ROOT)/backup/backups:/backups \
		-v $(HOST_ROOT)/backup/restore:/restore \
		restic/restic restore $(SNAPSHOT_ID) \
			--path $(PATH) \
			--target /restore

upload-backup:
	sudo chown -R 1000:1000 ./backups
	rclone copy ./backups googledrive:self-hosted-backups --progress

backup:
	docker compose run --rm backup

download-backup:
	rclone copy googledrive:self-hosted-backups ./backups --progress
