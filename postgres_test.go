package dbsuite

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PostgresTestSuite struct {
	Postgres
}

func (p *PostgresTestSuite) TestDatabaseIsCreated() {
	t := p.T()
	// Act
	row := p.DB.QueryRowContext(t.Context(), "select current_database()")

	// Assert
	require.NoError(t, row.Err())

	var result string

	err := row.Scan(&result)
	require.NoError(t, err)

	assert.Equal(t, toSafeDBName(t.Name()), result)
}

func (p *PostgresTestSuite) TestDatabaseIsDeletedAfterwards() {
	t := p.T()
	// Arrange
	dbName := toSafeDBName(t.Name())

	// This must be run AFTER TearDown, hence the weird construction. t.Context is also cancelled at this point
	t.Cleanup(func() {
		// Act
		err := p.suiteDB.QueryRowContext(context.Background(), "SELECT datname FROM pg_database WHERE datname = $1", dbName).Scan(new(""))

		// Assert
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestPostgres(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PostgresTestSuite))
}
