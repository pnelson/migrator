package migrator

import (
	"database/sql"
	"time"
)

// A version is an applied migration.
type version struct {
	id        int64
	version   string
	name      string
	createdAt time.Time
}

// queryVersionsNew creates the versions table if not already created.
var queryVersionsNew = `
CREATE TABLE IF NOT EXISTS versions (
  id         BIGSERIAL PRIMARY KEY,
  version    TEXT NOT NULL,
  name       TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// queryVersionsAll selects the applied migrations by ascending version.
var queryVersionsAll = `
SELECT id, version, name, created_at
  FROM versions
  ORDER BY version ASC;
`

// queryVersionsLast selects the version timestamp most recently applied.
var queryVersionsLast = `
SELECT version
  FROM versions
  ORDER BY version DESC
  LIMIT 1;
`

// queryVersionsInsert inserts a new version.
var queryVersionsInsert = `
INSERT INTO versions (version, name)
  VALUES ($1, $2);
`

// queryVersionsDelete deletes the version by timestamp.
var queryVersionsDelete = `
DELETE FROM versions
  WHERE version = $1;
`

// versions returns a slice of versions applied.
func versions(db *sql.DB) ([]*version, error) {
	var rv []*version
	rows, err := db.Query(queryVersionsAll)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		v := new(version)
		err := rows.Scan(&v.id, &v.version, &v.name, &v.createdAt)
		if err != nil {
			return nil, err
		}

		rv = append(rv, v)
	}

	err = rows.Err()
	if err != nil {
		return rv, err
	}

	return rv, nil
}

// currentVersion returns the version timestamp most recently applied.
func currentVersion(db *sql.DB) (string, error) {
	var v string

	err := db.QueryRow(queryVersionsLast).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}

	return v, err
}
