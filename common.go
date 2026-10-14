package dbsuite

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
)

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

// EnsureDatabase creates a database for the given (test) name. It uses the suite connection string
// to create it and returns a new connection object with a cancel func. This method is exposed to allow manual
// creation of databases for more complex test setups.
func EnsureDatabase(ctx context.Context, name string, connectionString string) (*sql.DB, func() error, error) {
	dbName := toSafeDBName(name)

	suiteDB, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open postgres to suite database: %w", err)
	}

	_, err = suiteDB.ExecContext(ctx, "CREATE DATABASE "+dbName+";")
	if err != nil {
		_ = suiteDB.Close() // Can't be deferred
		return nil, nil, fmt.Errorf("failed to create database %q: %w", dbName, err)
	}

	dbConnectionURL, err := url.Parse(connectionString)
	if err != nil {
		_ = suiteDB.Close() // Can't be deferred
		return nil, nil, fmt.Errorf("failed to parse connection URL: %w", err)
	}

	dbConnectionURL.Path = "/" + dbName

	db, err := sql.Open("postgres", dbConnectionURL.String())
	if err != nil {
		_ = suiteDB.Close() // Can't be deferred
		return nil, nil, fmt.Errorf("failed to open new database connection: %w", err)
	}

	//nolint:contextcheck // Intentionally not using the same context
	cancel := func() error {
		// Ensure both connections are closed after cancellation
		defer func() { _ = suiteDB.Close() }()
		defer func() { _ = db.Close() }()

		// Not using the test context as that may be cancelled at this point
		ctx, cancel := context.WithTimeout(context.Background(), dropTimeout)
		defer cancel()

		_, err := suiteDB.ExecContext(ctx, "DROP DATABASE "+dbName+" WITH (FORCE)")

		return err
	}

	return db, cancel, nil
}
