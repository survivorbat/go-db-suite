package dbsuite

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// maxDBNameLength is compatible with Postgres' naming schema
const maxDBNameLength = 64

// toSafeDBName removes special characters and ensures the name has a max length
func toSafeDBName(in string) string {
	output := make([]rune, 0, len(in))

	for _, char := range strings.ToLower(in) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			output = append(output, char)
		}
	}

	if len(output) <= maxDBNameLength {
		return string(output)
	}

	return string(output[len(output)-maxDBNameLength:])
}

// dropTimeout is used to ensure the deletion of a database has a timeout, as we're not
// using the test context when running the drop command.
const dropTimeout = 5 * time.Second

// ensureDatabase creates a database for the given (test) name. It uses the suite connection
// to create it and returns a new connection object with a cancel func. This method is exposed to allow manual
// creation of databases for more complex test setups. The connectionURL is required to edit the DB name into it.
func ensureDatabase(ctx context.Context, rootDB *sql.DB, connectionURL *url.URL, name string) (*sql.DB, func() error, error) {
	dbName := toSafeDBName(name)

	_, err := rootDB.ExecContext(ctx, "CREATE DATABASE "+dbName+";")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database %q: %w", dbName, err)
	}

	connectionURL.Path = "/" + dbName

	db, err := sql.Open("postgres", connectionURL.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open new database connection: %w", err)
	}

	//nolint:contextcheck // Intentionally not using the same context
	cancel := func() error {
		// Ensure both connections are closed after cancellation
		defer func() { _ = db.Close() }()

		// Not using the test context as that may be cancelled at this point
		ctx, cancel := context.WithTimeout(context.Background(), dropTimeout)
		defer cancel()

		_, err := rootDB.ExecContext(ctx, "DROP DATABASE "+dbName+" WITH (FORCE)")

		return err
	}

	return db, cancel, nil
}
