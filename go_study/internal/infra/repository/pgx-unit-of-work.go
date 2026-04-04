package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	transactionRef struct {
		tx pgx.Tx
	}
	transactionScope uint8
	savepointState   uint8
	PgxUnitOfWork    struct {
		config          *PgxUnitOfWorkConfig
		txRef           *transactionRef
		scope           transactionScope
		helloRepository repository.HelloRepository
	}
	PgxUnitOfWorkConfig struct {
		Pool *pgxpool.Pool
	}
	savepointTx struct {
		pgx.Tx
		savepoint string
		state     savepointState
	}
)

const (
	// ownedScope is a transaction started and finalized by this unit of work.
	ownedScope transactionScope = iota
	// joinedScope reuses a parent transaction and must never finalize it.
	joinedScope
	// nestedScope represents savepoint ownership; only savepoint lifecycle is finalized.
	nestedScope
)

const (
	savepointOpen savepointState = iota
	savepointCommitted
	savepointRolledBack
)

var (
	_ transaction.UnitOfWork = (*PgxUnitOfWork)(nil)
)

var (
	InvalidTransactionTypeErr = errors.New("invalid transaction type")
	InvalidPropagationErr     = errors.New("invalid propagation option")
	InvalidScopeTransitionErr = errors.New("invalid transaction scope transition")
	InvalidSavepointStateErr  = errors.New("invalid nested savepoint state")
	TransactionAlreadyOpenErr = errors.New("transaction already opened")
	TransactionNotOpenedErr   = errors.New("transaction not opened")
)

func NewPgxUnitOfWork(cfg *PgxUnitOfWorkConfig) *PgxUnitOfWork {
	return newPgxUnitOfWorkWithTxRef(cfg, &transactionRef{}, ownedScope)
}

func newPgxUnitOfWorkWithTxRef(cfg *PgxUnitOfWorkConfig, txRef *transactionRef, scope transactionScope) *PgxUnitOfWork {
	return &PgxUnitOfWork{
		config:          cfg,
		txRef:           txRef,
		scope:           scope,
		helloRepository: newHelloPgxRepository(cfg.Pool, txRef),
	}
}

func newSavepointTx(base pgx.Tx, savepoint string) pgx.Tx {
	return &savepointTx{
		Tx:        base,
		savepoint: savepoint,
		state:     savepointOpen,
	}
}

func (tx *savepointTx) Commit(ctx context.Context) error {
	if err := tx.ensureOpenFor("commit"); err != nil {
		return err
	}

	_, err := tx.Tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", tx.savepoint))
	if err != nil {
		return err
	}

	tx.state = savepointCommitted
	return err
}

func (tx *savepointTx) Rollback(ctx context.Context) error {
	if err := tx.ensureOpenFor("rollback"); err != nil {
		return err
	}

	_, err := tx.Tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", tx.savepoint))
	if err != nil {
		return err
	}

	_, err = tx.Tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", tx.savepoint))
	if err != nil {
		return err
	}

	tx.state = savepointRolledBack
	return err
}

func (tx *savepointTx) ensureOpenFor(action string) error {
	if tx.state == savepointOpen {
		return nil
	}

	return fmt.Errorf("%w: cannot %s savepoint %q after %s", InvalidSavepointStateErr, action, tx.savepoint, tx.stateLabel())
}

func (tx *savepointTx) stateLabel() string {
	switch tx.state {
	case savepointCommitted:
		return "commit"
	case savepointRolledBack:
		return "rollback"
	default:
		return "unknown"
	}
}

func isSavepointTx(tx pgx.Tx) bool {
	_, ok := tx.(*savepointTx)
	return ok
}

func (tm *PgxUnitOfWork) HelloRepository() repository.HelloRepository {
	return tm.helloRepository
}

func (tm *PgxUnitOfWork) Begin(ctx context.Context, opts transaction.Opts) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Begin"))
	defer end()

	if tm.txRef.current() != nil {
		if tm.scope != ownedScope {
			err := fmt.Errorf("%w: begin is not allowed for externally scoped transactions", InvalidScopeTransitionErr)
			trace.InjectError(ctx, err)
			return err
		}

		trace.InjectError(ctx, TransactionAlreadyOpenErr)
		return TransactionAlreadyOpenErr
	}

	tx, err := tm.resolveTransaction(ctx, opts)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	tm.txRef.set(tx)
	return nil
}

func (tm *PgxUnitOfWork) Commit(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Commit"))
	defer end()

	if err := tm.ensureCanFinalize("commit"); err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	tx := tm.txRef.current()
	if tx == nil {
		trace.InjectError(ctx, TransactionNotOpenedErr)
		return TransactionNotOpenedErr
	}
	defer tm.clear()

	err := tx.Commit(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}
	return nil
}

func (tm *PgxUnitOfWork) Rollback(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Rollback"))
	defer end()

	if err := tm.ensureCanFinalize("rollback"); err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	tx := tm.txRef.current()
	if tx == nil {
		trace.InjectError(ctx, TransactionNotOpenedErr)
		return TransactionNotOpenedErr
	}
	defer tm.clear()

	err := tx.Rollback(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}
	return nil
}

func (tm *PgxUnitOfWork) clear() {
	tm.txRef.set(nil)
}

func (tm *PgxUnitOfWork) resolveTransaction(ctx context.Context, opts transaction.Opts) (pgx.Tx, error) {
	propagation := opts.EffectivePropagation()

	switch propagation {
	case transaction.PropagationRequiresNew:
		tx, err := tm.beginTransaction(ctx)
		if err != nil {
			return nil, err
		}
		return tx, nil
	case transaction.PropagationJoin:
		tx, err := tm.beginTransaction(ctx)
		if err != nil {
			return nil, err
		}
		return tx, nil
	case transaction.PropagationNested:
		return nil, fmt.Errorf("%w: nested propagation requires an existing parent transaction scope", InvalidScopeTransitionErr)
	default:
		return nil, InvalidPropagationErr
	}
}

func (tm *PgxUnitOfWork) ensureCanFinalize(action string) error {
	if tm.scope != joinedScope {
		return nil
	}

	return fmt.Errorf("%w: cannot %s a joined transaction owned by parent scope", InvalidScopeTransitionErr, action)
}

func (tm *PgxUnitOfWork) beginTransaction(ctx context.Context) (pgx.Tx, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "BeginTransaction"))
	defer end()

	return tm.config.Pool.Begin(ctx)
}

func (r *transactionRef) current() pgx.Tx {
	if r == nil {
		return nil
	}
	return r.tx
}

func (r *transactionRef) set(tx pgx.Tx) {
	if r == nil {
		return
	}
	r.tx = tx
}
