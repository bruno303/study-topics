package repository

import (
	"context"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
)

func TestUnitOfWorkFactoryImpl_Create_ReturnsPgxUnitOfWork(t *testing.T) {
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: nil})

	uow := factory.Create()
	if uow == nil {
		t.Fatal("expected unit of work, got nil")
	}
	if _, ok := uow.(*PgxUnitOfWork); !ok {
		t.Fatalf("expected *PgxUnitOfWork, got %T", uow)
	}
}

func TestMemDbUnitOfWorkFactoryImpl_Create_ReturnsMemDbUnitOfWork(t *testing.T) {
	factory := NewMemDbUnitOfWorkFactory()

	uow := factory.Create()
	if uow == nil {
		t.Fatal("expected unit of work, got nil")
	}
	if _, ok := uow.(*MemDbUnitOfWork); !ok {
		t.Fatalf("expected *MemDbUnitOfWork, got %T", uow)
	}
}

func TestMemDbUnitOfWorkFactoryImpl_Create_WhenStorageIsInjected_UsesSharedBackingStateAcrossFreshUnitOfWorks(t *testing.T) {
	storage := NewMemDbStorage()
	factory := NewMemDbUnitOfWorkFactory(storage)

	uow1, ok := factory.Create().(*MemDbUnitOfWork)
	if !ok {
		t.Fatalf("expected *MemDbUnitOfWork, got %T", factory.Create())
	}
	uow2, ok := factory.Create().(*MemDbUnitOfWork)
	if !ok {
		t.Fatalf("expected *MemDbUnitOfWork, got %T", factory.Create())
	}

	if uow1 == uow2 {
		t.Fatal("expected fresh unit of work instances per create call")
	}

	entity := &models.HelloData{Id: "factory-injected-storage", Name: "Shared", Age: 10}

	if _, err := uow1.HelloRepository().Save(context.Background(), entity); err != nil {
		t.Fatalf("expected save without error, got %v", err)
	}

	list, err := uow2.HelloRepository().ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected list without error, got %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected one entity in list, got %d", len(list))
	}
	if list[0] != *entity {
		t.Fatalf("expected listed entity %+v, got %+v", *entity, list[0])
	}
}
