package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *PostgresSQL) InsertAction(job_id, name string, payload map[string]interface{}) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, execErr := client.Exec(ctx, "INSERT INTO action (job_id, name, payload) VALUES ($1, $2, $3)", job_id, name, payload)
		if execErr != nil {
			return nil, execErr
		}

		return nil, nil
	})

	return err
}

func (p *PostgresSQL) DeleteAction(job_id string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, execErr := client.Exec(ctx, "DELETE FROM action WHERE job_id = $1", job_id)
		if execErr != nil {
			return nil, execErr
		}

		return nil, nil
	})

	return err
}
