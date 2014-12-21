migrator
========

Package migrator provides basic SQL migration capabilities.

This package is only tested with PostgreSQL using the lib/pq driver. I
don't expect this to work with any other setup. This package exists mostly
for myself but if it helps you then that is cool too. The public API of
this package is pretty minimal so maybe one day I'll get around to making
it a bit more flexible and actually writing some tests.


Usage
-----

I create a file per migration version and they look something like this.

```go
func Up_20140630T023811Z(tx *sql.Tx) error {
  _, err := tx.Exec(`CREATE EXTENSION hstore;`)
  return err
}

func Down_20140630T023811Z(tx *sql.Tx) error {
  _, err := tx.Exec(`DROP EXTENSION hstore;`)
  return err
}

func init() {
  migrator.Register("20140630T023811Z", "enable_extensions",
    Up_20140630T023811Z, Down_20140630T023811Z)
}
```

To migrate up to the latest version...

```go
migrator.Migrate(db, "")
```

To migrate up or down to a specific migration...

```go
migrator.Migrate(db, "20140630T023811Z")
```

To view the current status of migrations...

```go
migrator.Status(db)
```

Copyright (c) 2015 by Philip Nelson. See LICENSE for details.
