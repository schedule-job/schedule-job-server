package pg

func New(dsn string) *PostgresSQL {
	return &PostgresSQL{Dsn: dsn}
}
