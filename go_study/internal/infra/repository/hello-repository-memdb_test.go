package repository

import (
	"context"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

func TestHelloMemDbRepository_SaveThenListAll_ReturnsSavedEntity(t *testing.T) {
	db := database.NewMemDbRepository[models.HelloData]()
	repo := NewHelloMemDbRepository(db)

	entity := &models.HelloData{Id: "id-1", Name: "Bruno", Age: 29}

	saved, err := repo.Save(context.Background(), entity)
	if err != nil {
		t.Fatalf("expected save without error, got %v", err)
	}
	if saved == nil {
		t.Fatal("expected saved entity, got nil")
	}
	if *saved != *entity {
		t.Fatalf("expected saved entity %+v, got %+v", *entity, *saved)
	}

	list, err := repo.ListAll(context.Background())
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

func TestHelloMemDbRepository_ListAll_WhenEmpty_ReturnsEmptySlice(t *testing.T) {
	db := database.NewMemDbRepository[models.HelloData]()
	repo := NewHelloMemDbRepository(db)

	list, err := repo.ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected list without error, got %v", err)
	}
	if list == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d items", len(list))
	}
}
