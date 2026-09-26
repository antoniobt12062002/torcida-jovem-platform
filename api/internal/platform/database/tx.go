package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// TxFrom returns the transaction carried by ctx, if any. Code that must run
// inside a use case's transaction, such as the audit recorder, uses it to fail
// loudly when it is called outside one.
func TxFrom(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok && tx != nil
}

// WithTx runs fn inside a transaction that travels in the context passed to
// fn. It commits when fn returns nil and rolls back on any error or panic (the
// panic is propagated). When ctx already carries a transaction, fn joins it:
// one use case is one unit of work.
func WithTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	if _, ok := TxFrom(ctx); ok {
		return fn(ctx)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}
