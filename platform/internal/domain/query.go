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
