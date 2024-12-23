package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type fileTransaction struct {
	paths map[string]string
}

func (t fileTransaction) Commit(ctx context.Context) error {
	for tmpPath, finalPath := range t.paths {
		err := os.Rename(tmpPath, finalPath)
		if err != nil {
			t.deleteAll()
			return err
		}
	}
	return nil
}

func (t fileTransaction) Rollback(ctx context.Context) error {
	for tmpPath := range t.paths {
		err := os.Remove(tmpPath)
		if err != nil {
			t.deleteAll()
			return err
		}
	}
	return nil
}

func (t fileTransaction) deleteAll() {
	for tmpPath, finalPath := range t.paths {
		_ = os.Remove(tmpPath)
		_ = os.Remove(finalPath)
	}
}

type FileRepository struct{}

func NewFileRepository() FileRepository {
	return FileRepository{}
}

func (r FileRepository) WriteFile(ctx context.Context, path string, content []byte) error {
	tx := GetTransactionOrNil(ctx)
	if tx == nil {
		return os.WriteFile(path, content, 0644)
	}

	tmpPath := fmt.Sprintf("%s/%s", os.TempDir(), uuid.NewString())
	tx.fileTransaction.paths[tmpPath] = path
	return os.WriteFile(tmpPath, content, 0644)
}
