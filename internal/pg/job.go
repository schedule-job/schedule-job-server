package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *PostgresSQL) InsertJob(name, description, author string, members []string) (string, error) {
	var id string
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		queryErr := client.QueryRow(ctx, "INSERT INTO job (name, description, author, members) VALUES ($1, $2, $3, $4) RETURNING id", name, description, author, members).Scan(&id)
		if queryErr != nil {
			return nil, queryErr
		}

		return nil, nil
	})

	return id, err
}

func (p *PostgresSQL) DeleteJob(job_id string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, execErr := client.Exec(ctx, "DELETE FROM job WHERE job_id = $1", job_id)
		if execErr != nil {
			return nil, execErr
		}

		return nil, nil
	})

	return err
}
