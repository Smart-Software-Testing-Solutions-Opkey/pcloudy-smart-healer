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
	HealingQueue     HealingQueue
}

// because the impls will all be interface
// we don't need to add pointer to T
type StoreImpl[T any] func(*sqlx.Tx) T

type UnitOfWorkFactory struct {
	db                      *sqlx.DB
	pageStoreFactory        StoreImpl[PageStore]
	locatorStoreFactory     StoreImpl[LocatorStore]
	descriptionQueueFactory StoreImpl[DescriptionQueue]
	healingQueueFactory     StoreImpl[HealingQueue]
}

type FactoryParams struct {
	PageStoreFactory        StoreImpl[PageStore]
	LocatorStoreFactory     StoreImpl[LocatorStore]
	DescriptionQueueFactory StoreImpl[DescriptionQueue]
	HealingQueueFactory     StoreImpl[HealingQueue]
}

func NewUnitOfWorkFactory(db *sqlx.DB, p FactoryParams) UnitOfWorkFactory {
	return UnitOfWorkFactory{
		db:                      db,
		pageStoreFactory:        p.PageStoreFactory,
		locatorStoreFactory:     p.LocatorStoreFactory,
		descriptionQueueFactory: p.DescriptionQueueFactory,
		healingQueueFactory:     p.HealingQueueFactory,
	}
}

func (f *UnitOfWorkFactory) NewUnitOfWork(ctx context.Context) (*UnitOfWork, error) {
	tx, err := f.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create unit of work: %w", err)
	}

	return &UnitOfWork{
		tx:               tx,
		Pages:            f.pageStoreFactory(tx),
		Locators:         f.locatorStoreFactory(tx),
		DescriptionQueue: f.descriptionQueueFactory(tx),
		HealingQueue:     f.healingQueueFactory(tx),
	}, nil
}

func (u *UnitOfWork) Commit() error {
	return u.tx.Commit()
}

func (u *UnitOfWork) Rollback() error {
	return u.tx.Rollback()
}
