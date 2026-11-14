package domain

// QueryResult is the outcome of running a single SQL statement against an
// instance's database.
type QueryResult struct {
	// Columns is empty for statements that don't return rows (INSERT,
	// UPDATE, DELETE, DDL) - RowCount then comes from the command tag
	// (e.g. rows affected) instead of len(Rows).
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`

	RowCount int `json:"rowCount"`
	// Truncated is true when more rows existed than the server-side cap
	// (see infra/pgquery) - only Rows up to that cap are included.
	Truncated bool `json:"truncated"`

	DurationMs int64 `json:"durationMs"`
}

// QueryError is a statement the database rejected: a syntax error, a missing
// table, a constraint violation. The fields come from PostgreSQL and are safe
// to return to the caller, who owns the database and wrote the statement.
// Message and Detail can contain row values, so don't log them.
type QueryError struct {
	// SQLState is the PostgreSQL error code, e.g. 42P01 (undefined table).
	SQLState string
	Message  string
	Detail   string
	Hint     string
	// Position is the 1-based character offset of the error in the
	// statement, or 0 when PostgreSQL doesn't report one.
	Position int
}

func (e *QueryError) Error() string {
	return e.Message
}
