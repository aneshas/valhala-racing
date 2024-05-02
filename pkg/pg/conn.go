package pg

import (
	"context"
	"database/sql"
	"github.com/aneshas/tx/v2/sqltx"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Conn returns the context executor for the transaction
func Conn(ctx context.Context, db *sql.DB) boil.ContextExecutor {
	if tx, ok := sqltx.From(ctx); ok {
		return tx
	}

	return db
}
