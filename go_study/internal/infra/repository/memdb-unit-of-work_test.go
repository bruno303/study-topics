package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
)

func TestMemDbUnitOfWork_BeginCommitRollback_AreNoOps(t *testing.T) {
	uow := NewMemDbUnitOfWork(nil)
	ctx := context.Background()

	if err := uow.Begin(ctx); err != nil {
		t.Fatalf("expected begin without error, got %v", err)
	}
	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("expected commit without error, got %v", err)
	}
	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("expected rollback without error, got %v", err)
	}
}

func TestMemDbUnitOfWork_HelloRepository_ReturnsRepository(t *testing.T) {
	uow := NewMemDbUnitOfWork(nil)

	if repo := uow.HelloRepository(); repo == nil {
		t.Fatal("expected hello repository to be initialized")
	}
}

func TestMemDbUnitOfWorkFactory_Create_UsesSharedStorageAcrossUnitOfWorks(t *testing.T) {
	factory := NewMemDbUnitOfWorkFactory()

	uow1, ok := factory.Create().(*MemDbUnitOfWork)
	if !ok {
		t.Fatalf("expected *MemDbUnitOfWork, got %T", factory.Create())
	}
	uow2, ok := factory.Create().(*MemDbUnitOfWork)
	if !ok {
		t.Fatalf("expected *MemDbUnitOfWork, got %T", factory.Create())
	}

	id := fmt.Sprintf("shared-%d", time.Now().UnixNano())
	entity := &models.HelloData{Id: id, Name: "Shared", Age: 10}

	if _, err := uow1.HelloRepository().Save(context.Background(), entity); err != nil {
		t.Fatalf("expected save without error, got %v", err)
	}

	list, err := uow2.HelloRepository().ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected list without error, got %v", err)
	}

	found := false
	for _, item := range list {
		if item.Id == id && item.Name == entity.Name && item.Age == entity.Age {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected second unit of work to see entity with id %q", id)
	}
}
