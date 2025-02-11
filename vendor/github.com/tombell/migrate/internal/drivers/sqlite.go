package drivers

type SQLite struct {
	PostgresSQL
}

func (s *SQLite) Name() string {
	return "sqlite"
}
