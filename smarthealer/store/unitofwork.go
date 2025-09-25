package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UnitOfWork struct {
	tx               *sqlx.Tx
	Pages            PageStore
	Locators         LocatorStore
	DescriptionQueue DescriptionQueue
}

// because the impls will all be interface
// we don't need to add pointer to T
type StoreImpl[T any] func(*sqlx.Tx) T

type UnitOfWorkFactory struct {
	Db                      *sqlx.DB
	PageStoreFactory        StoreImpl[PageStore]
	LocatorStoreFactory     StoreImpl[LocatorStore]
	DescriptionQueueFactory StoreImpl[DescriptionQueue]
}

func (f *UnitOfWorkFactory) NewUnitOfWork(ctx context.Context) (*UnitOfWork, error) {
	tx, err := f.Db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create unit of work: %w", err)
	}

	return &UnitOfWork{
		tx:               tx,
		Pages:            f.PageStoreFactory(tx),
		Locators:         f.LocatorStoreFactory(tx),
		DescriptionQueue: f.DescriptionQueueFactory(tx),
	}, nil
}

func (u *UnitOfWork) Commit() error {
	return u.tx.Commit()
}

func (u *UnitOfWork) Rollback() error {
	return u.tx.Rollback()
}
