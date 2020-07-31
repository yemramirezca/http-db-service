package postgres

import "regexp"

const (

  PostgresTableCreationQuery = `
    CREATE TABLE IF NOT EXISTS {name} (
      order_id VARCHAR(64),
      namespace VARCHAR(64),
      total DECIMAL(8,2),
      PRIMARY KEY (order_id, namespace)
    )
`
)

var safeSQLRegex = regexp.MustCompile(`[^a-zA-Z0-9\.\-_]`)

// SanitizeSQLArg returns the input string sanitized for safe use in an SQL query as argument.
func SanitizeSQLArg(s string) string {
	return safeSQLRegex.ReplaceAllString(s, "")
}
