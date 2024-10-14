package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/schedule-job/schedule-job-server/internal/errorset"
)

type jobInfo struct {
	JobID       string    `json:"job_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Members     []string  `json:"members"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (p *PostgresSQL) InsertJob(name, description, author string, members []string) (string, error) {
	var job_id string
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		errQuery := client.QueryRow(ctx, "INSERT INTO job (name, description, author, members) VALUES ($1, $2, $3, $4) RETURNING job_id", name, description, author, members).Scan(&job_id)
		if errQuery != nil {
			return nil, errQuery
		}

		return nil, nil
	})

	if err != nil {
		return "", errorset.ErrSQL
	}

	return job_id, nil
}

func (p *PostgresSQL) UpdateJob(job_id, name, description, author string, members []string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, errQuery := client.Exec(ctx, "INSERT INTO job (job_id, name, description, author, members) VALUES ($1, $2, $3, $4, $5)", job_id, name, description, author, members)
		if errQuery != nil {
			return nil, errQuery
		}

		return nil, nil
	})

	if err != nil {
		return errorset.ErrSQL
	}

	return nil
}

func (p *PostgresSQL) DeleteJob(job_id string) error {
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		_, errExec := client.Exec(ctx, "DELETE FROM job WHERE job_id = $1", job_id)
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

func (p *PostgresSQL) SelectJob(job_id string) (*jobInfo, error) {
	job := jobInfo{JobID: job_id}
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		errQuery := client.QueryRow(ctx, "SELECT name, description, author, members, created_at FROM job WHERE job_id = $1 ORDER BY created_at desc", job_id).Scan(
			&job.Name,
			&job.Description,
			&job.Author,
			&job.Members,
			&job.CreatedAt,
		)

		if errQuery != nil {
			return nil, errQuery
		}

		return nil, nil
	})

	if err != nil {
		return nil, errorset.ErrSQL
	}

	return &job, nil
}

func (p *PostgresSQL) SelectJobs(user, lastId string, limit int) (*[]jobInfo, error) {
	var lastCreated = time.Date(1990, 1, 1, 1, 0, 0, 0, time.UTC)
	if lastId != "" {
		job, err := p.SelectJob(lastId)
		if err == nil {
			lastCreated = job.CreatedAt
		}
	}

	jobs := []jobInfo{}
	_, err := p.usePostgresSQL(func(client *pgx.Conn, ctx context.Context) (result interface{}, err error) {
		rows, errQuery := client.Query(ctx, "SELECT t1.job_id, t1.name, t1.description, t1.author, t1.members, t1.created_at FROM job t1 INNER JOIN ( SELECT job_id, MAX(created_at) AS latest_date FROM job GROUP BY job_id ) t2 ON t1.job_id = t2.job_id AND t1.created_at = t2.latest_date WHERE created_at > $2 AND ($1 = t1.author OR $1 = ANY(t1.members)) ORDER BY created_at LIMIT $3", user, lastCreated, limit)

		if errQuery != nil {
			return nil, errQuery
		}

		for rows.Next() {
			job := jobInfo{}
			errScan := rows.Scan(
				&job.JobID,
				&job.Name,
				&job.Description,
				&job.Author,
				&job.Members,
				&job.CreatedAt,
			)
			if errScan != nil {
				continue
			}
			jobs = append(jobs, job)
		}

		return nil, nil
	})

	if err != nil {
		return nil, errorset.ErrSQL
	}

	return &jobs, nil
}
