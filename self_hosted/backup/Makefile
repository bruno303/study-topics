MAKEFLAGS += --silent

include .env
export

.PHONY: list-snapshots
list-snapshots:
	restic snapshots

.PHONY: browse-files
browse-files:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: browse-files SNAPSHOT_ID=foo" && exit 1)
	restic ls $(SNAPSHOT_ID)

.PHONY: restore
restore:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: restore SNAPSHOT_ID=foo" && exit 1)
	restic restore $(SNAPSHOT_ID) --target ./restore

.PHONY: restore-path
restore-path:
	@test -n "$(SNAPSHOT_ID)" || (echo "❌ SNAPSHOT_ID variable is missing. Use: restore-path SNAPSHOT_ID=foo" && exit 1)
	@test -n "$(PATH)" || (echo "❌ PATH variable is missing. Use: restore-path PATH=foo" && exit 1)
	restic restore $(SNAPSHOT_ID) --path $(PATH) --target ./restore

.PHONY: save-backup
save-backup:
	./save-backup.sh

.PHONY: backup
backup:
	./backup.sh

.PHONY: full-backup
full-backup: backup save-backup

.PHONY: verify-backup
verify-backup:
	cd backup-verifier && ./run-verifier.sh
