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

func (p *PostgresSQL) DeleteJob(id string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, execErr := client.Exec(ctx, "DELETE FROM job WHERE id = $1", id)
		if execErr != nil {
			return nil, execErr
		}

		return nil, nil
	})

	return err
}

func (p *PostgresSQL) SelectJob(id string) (*jobInfo, error) {
	job := jobInfo{}
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		queryErr := client.QueryRow(ctx, "SELECT name, description, author, members, created_at FROM job WHERE id = $1 ORDER BY created_at", id).Scan(
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
