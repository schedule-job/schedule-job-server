package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/schedule-job/schedule-job-server/internal/errorset"
)

type actionInfo struct {
	JobId   string                 `json:"jobId"`
	Name    string                 `json:"name"`
	Payload map[string]interface{} `json:"payload"`
}

func (p *PostgresSQL) InsertAction(job_id, name string, payload map[string]interface{}) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, errExec := client.Exec(ctx, "INSERT INTO action (job_id, name, payload) VALUES ($1, $2, $3)", job_id, name, payload)
		if errExec != nil {
			return nil, errExec
		}

		return nil, nil
	})
	if err != nil {
		return errorset.ErrSQL
	}
	return nil
}

func (p *PostgresSQL) UpdateAction(job_id, name string, payload map[string]interface{}) error {
	return p.InsertAction(job_id, name, payload)
}

func (p *PostgresSQL) DeleteAction(job_id string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, errExec := client.Exec(ctx, "DELETE FROM action WHERE job_id = $1", job_id)
		if errExec != nil {
			return nil, errExec
		}
		return nil, nil
	})
	if err != nil {
		return errorset.ErrSQL
	}
	return nil
}

func (p *PostgresSQL) SelectAction(job_id string) (*actionInfo, error) {
	info := actionInfo{}
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		errQuery := client.QueryRow(ctx, "SELECT job_id, name, payload FROM action WHERE job_id = $1 ORDER BY created_at desc", job_id).Scan(
			&info.JobId,
			&info.Name,
			&info.Payload,
		)

		if errQuery != nil {
			return nil, errQuery
		}

		return nil, nil
	})

	if err != nil {
		return nil, errorset.ErrSQL
	}

	return &info, nil
}
