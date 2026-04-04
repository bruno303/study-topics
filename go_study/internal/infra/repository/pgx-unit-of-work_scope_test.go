package repository

import (
	"context"
	"errors"
	"testing"
)

func TestPgxUnitOfWork_JoinedScope_CommitRollback_ReturnInvalidScopeTransitionErr(t *testing.T) {
	uow := newPgxUnitOfWorkWithTxRef(&PgxUnitOfWorkConfig{}, &transactionRef{}, joinedScope)
	ctx := context.Background()

	if err := uow.Commit(ctx); !errors.Is(err, InvalidScopeTransitionErr) {
		t.Fatalf("expected InvalidScopeTransitionErr on commit for joined scope, got: %v", err)
	}

	if err := uow.Rollback(ctx); !errors.Is(err, InvalidScopeTransitionErr) {
		t.Fatalf("expected InvalidScopeTransitionErr on rollback for joined scope, got: %v", err)
	}
}
