package dbsuite

import (
	"database/sql"
	"strings"
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

	assert.Equal(t, "testpostgrestestdatabaseiscreated", result)
}

func (p *PostgresTestSuite) TestDatabaseIsDeletedAfterwards() {
	t := p.T()
	// Arrange
	dbString := strings.ReplaceAll(p.suiteConnectionString, "/postgres?", "/testpostgrestestdatabaseiscreated?")

	db, err := sql.Open("postgres", dbString)
	require.NoError(t, err)

	// Act
	row := db.QueryRowContext(t.Context(), "select current_database()")

	// Assert
	require.ErrorContains(t, row.Err(), "does not exist")
}

func TestPostgres(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PostgresTestSuite))
}
