package dbaas

import "time"

// InstanceStatus is where an instance is in its lifecycle.
type InstanceStatus string

const (
	// StatusPending: created or updated, the operator hasn't finished yet.
	StatusPending  InstanceStatus = "Pending"
	StatusRunning  InstanceStatus = "Running"
	StatusFailed   InstanceStatus = "Failed"
	StatusDeleting InstanceStatus = "Deleting"
	StatusDeleted  InstanceStatus = "Deleted"
)

// Instance is a PostgreSQL instance.
type Instance struct {
	// ID identifies the instance in every other call. It's derived from
	// the idempotency key it was created with.
	ID string `json:"id"`
	// Name is the system-generated resource name (currently equal to ID).
	Name string `json:"name"`
	// Database is the database name chosen at creation time.
	Database       string         `json:"database"`
	Version        int            `json:"version"`
	Storage        string         `json:"storage"`
	Instances      int            `json:"instances"`
	ReadyInstances int            `json:"readyInstances"`
	Status         InstanceStatus `json:"status"`
	CreatedAt      time.Time      `json:"createdAt"`
	// ExternalPort is 0 until the instance is reachable from outside the
	// cluster.
	ExternalPort int `json:"externalPort"`
}

// InstanceCredentials is how to connect to an instance's database.
type InstanceCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	// URI is a ready-to-use postgresql:// connection string.
	URI string `json:"uri"`
}

// LogLine is one line of an instance's or the audit trail's logs.
type LogLine struct {
	// Timestamp is Loki's timestamp: Unix nanoseconds, as a string.
	Timestamp string `json:"timestamp"`
	Line      string `json:"line"`
}

// QueryResult is the outcome of one SQL statement run with
// InstancesService.Query.
type QueryResult struct {
	// Columns is empty for statements that don't return rows (INSERT,
	// UPDATE, DDL); RowCount then is the number of rows affected.
	Columns []string `json:"columns"`
	// Rows holds JSON values: numbers decode as float64.
	Rows     [][]any `json:"rows"`
	RowCount int     `json:"rowCount"`
	// Truncated is true when the result had more rows than the server
	// returns.
	Truncated  bool  `json:"truncated"`
	DurationMs int64 `json:"durationMs"`
}
