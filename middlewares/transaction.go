package middlewares

import (
	"database/sql"

	"github.com/labstack/echo/v4"
)

func TransactionMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			trx, err := db.Begin()
			if err != nil {
				panic(err)
			}
			ctx.Set("db_trx", trx)
			defer func() {
				if r := recover(); r != nil {
					trx.Rollback()
					panic(r)
				} else if ctx.Response().Status >= 500 {
					trx.Rollback()
				} else {
					trx.Commit()
				}
			}()
			return next(ctx)
		}
	}
}
