package dbsuite

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestToAlphanumeric_OutputsExpectedName(t *testing.T) {
	t.Parallel()
	// Arrange
	in := "Hello I'm so happy! 😊 I'm writing this test to ensure no weird DB names come through! That'd be bad 🚨"

	// Act
	actual := toSafeDBName(in)

	// Assert
	assert.Equal(t, "appyimwritingthistesttoensurenoweirddbnamescomethroughthatdbebad", actual)
}

func TestEnsureDatabase_CreatesDatabaseAsExpected(t *testing.T) {
	t.Parallel()

	container, err := postgrescontainer.Run(t.Context(), "postgres:alpine", postgrescontainer.BasicWaitStrategies())
	require.NoError(t, err)

	t.Cleanup(func() { _ = container.Terminate(t.Context()) })

	dbNames := []string{"foo", "bar"}

	for _, name := range dbNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			connectionString := container.MustConnectionString(t.Context(), "sslmode=disable")

			// Act
			db, cancel, err := ensureDatabase(t.Context(), t.Name(), connectionString)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, cancel)

			defer func() { _ = cancel() }()

			row := db.QueryRowContext(t.Context(), "select current_database()")
			require.NoError(t, row.Err())

			var result string

			err = row.Scan(&result)
			require.NoError(t, err)

			assert.Equal(t, toSafeDBName(t.Name()), result)
		})
	}
}

func TestEnsureDatabase_CancelDeletesAndClosesDatabase(t *testing.T) {
	t.Parallel()

	container, err := postgrescontainer.Run(t.Context(), "postgres:alpine", postgrescontainer.BasicWaitStrategies())
	require.NoError(t, err)

	t.Cleanup(func() { _ = container.Terminate(t.Context()) })

	dbNames := []string{"foo", "bar"}

	for _, name := range dbNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			suiteConnectionString := container.MustConnectionString(t.Context(), "sslmode=disable")

			db, cancel, err := ensureDatabase(t.Context(), t.Name(), suiteConnectionString)
			require.NoError(t, err)
			require.NotNil(t, cancel)

			// Act
			err = cancel()

			// Assert
			require.NoError(t, err)

			// Ensure closed
			row := db.QueryRowContext(t.Context(), "select current_database()")
			require.ErrorContains(t, row.Err(), "database is closed")

			suiteDB, err := sql.Open("postgres", suiteConnectionString)
			require.NoError(t, err)

			dbName := toSafeDBName(t.Name())

			err = suiteDB.QueryRowContext(t.Context(), "SELECT datname FROM pg_database WHERE datname = $1", dbName).Scan(new(""))
			require.ErrorIs(t, err, sql.ErrNoRows)
		})
	}
}
