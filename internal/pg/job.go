package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type jobInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Members     []string  `json:"members"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (p *PostgresSQL) InsertJob(name, description, author string, members []string) (string, error) {
	var job_id string
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		queryErr := client.QueryRow(ctx, "INSERT INTO job (name, description, author, members) VALUES ($1, $2, $3, $4) RETURNING job_id", name, description, author, members).Scan(&job_id)
		if queryErr != nil {
			return nil, queryErr
		}

		return nil, nil
	})

	return job_id, err
}

func (p *PostgresSQL) UpdateJob(job_id, name, description, author string, members []string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, queryErr := client.Exec(ctx, "INSERT INTO job (job_id, name, description, author, members) VALUES ($1, $2, $3, $4, $5)", job_id, name, description, author, members)
		if queryErr != nil {
			return nil, queryErr
		}

		return nil, nil
	})
	return err
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

func (p *PostgresSQL) SelectJob(job_id string) (*jobInfo, error) {
	job := jobInfo{}
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		queryErr := client.QueryRow(ctx, "SELECT name, description, author, members, created_at FROM job WHERE job_id = $1 ORDER BY created_at desc", job_id).Scan(
			&job.Name,
			&job.Description,
			&job.Author,
			&job.Members,
			&job.CreatedAt,
		)

		if queryErr != nil {
			return nil, queryErr
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return &job, nil
}
