package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PostgresSQL struct {
	Dsn string
}

func (pg *PostgresSQL) usePostgresSQL(callback func(client *pgx.Conn, ctx context.Context) (result interface{}, err error)) (result interface{}, err error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, pg.Dsn)

	if err != nil {
		panic(err)
	}

	defer conn.Close(ctx)

	return callback(conn, ctx)
}
