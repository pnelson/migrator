// Package migrator provides basic SQL migration capabilities.
package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
)

// A migration is a named pair of migrationFunc.
type migration struct {
	name string
	up   migrationFunc
	down migrationFunc
}

// A migrationFunc is a function that performs operations on a
// SQL transaction and returns an error.
type migrationFunc func(tx *sql.Tx) error

// migrations is a map of migration keyed by version timestamp.
var migrations = make(map[string]*migration)

// Register makes a migration available by the provided name.
// If Register is called twice with the same name or if a
// migrationFunc is nil, it panics.
func Register(version, name string, up, down migrationFunc) {
	if up == nil || down == nil {
		panic("migrator: Register up and down are both required")
	}

	if _, ok := migrations[version]; ok {
		panic("migrator: Register called twice for migrator " + version)
	}

	migrations[version] = &migration{name: name, up: up, down: down}
}

// Migrate performs the database migrations to bring the database
// to the state of the target version timestamp. Use an empty target
// to represent the most recent migration.
func Migrate(db *sql.DB, target string) error {
	vs := sorted()
	if target == "" {
		target = vs[len(vs)-1]
	}

	_, err := db.Exec(queryVersionsNew)
	if err != nil {
		return err
	}

	current, err := currentVersion(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying latest migration version: %v", err)
		return err
	}

	up := true
	if current > target {
		sort.Sort(sort.Reverse(sort.StringSlice(vs)))
		up = false
	}

	for _, v := range vs {
		if !shouldMigrate(v, current, target, up) {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		err = migrate(tx, v, up)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error migrating %q: %v\n", v, err)
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// Status prints the sorted list of migrations and whether or not
// they have been applied to the database.
func Status(db *sql.DB) error {
	vs, err := versions(db)
	if err != nil {
		return err
	}

	for _, v := range sorted() {
		s := " "
		if applied(v, vs) {
			s = "x"
		}
		fmt.Printf("[%s] %s %s\n", s, v, migrations[v].name)
	}

	return nil
}

// applied returns true if version if found in the slice of versions.
func applied(version string, vs []*version) bool {
	for _, v := range vs {
		if v.version == version {
			return true
		}
	}

	return false
}

// shouldMigrate returns whether the migration for version needs to be
// performed based on the current and target version timestamps and
// the direction of migrations.
func shouldMigrate(version, current, target string, up bool) bool {
	if !up {
		return version <= current && version > target
	}

	return version > current && version <= target
}

// sorted returns a slice of version timestamps in ascending order.
func sorted() []string {
	var rv []string
	for k := range migrations {
		rv = append(rv, k)
	}

	sort.Strings(rv)

	return rv
}

// migrate executes the appropriate migrationFunc within the transaction
// and records the migration in the versions table.
func migrate(tx *sql.Tx, version string, up bool) error {
	var err error

	if !up {
		err = migrations[version].down(tx)
		if err != nil {
			return err
		}

		_, err = tx.Exec(queryVersionsDelete, version)
		return err
	}

	err = migrations[version].up(tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(queryVersionsInsert, version, migrations[version].name)
	return err
}

// empty is a nil migratorFunc for the purpose of having an empty state
// to migrate down to.
func empty(tx *sql.Tx) error {
	return nil
}

func init() {
	Register("00010101T000000Z", "nil", empty, empty)
}
