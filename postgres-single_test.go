package dbsuite

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SinglePostgresTestSuite struct {
	SinglePostgres
}

func (p *SinglePostgresTestSuite) TestDatabaseIsCreated() {
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

func (p *SinglePostgresTestSuite) TestDatabaseIsDeletedAfterwards() {
	t := p.T()
	// Arrange
	dbName := toSafeDBName(t.Name())

	// This must be run AFTER TearDown, hence the weird construction. t.Context is also cancelled at this point
	t.Cleanup(func() {
		db, err := sql.Open("postgres", p.suiteConnectionString)
		require.NoError(t, err)

		// Act
		err = db.QueryRowContext(context.Background(), "SELECT datname FROM pg_database WHERE datname = $1", dbName).Scan(new(""))

		// Assert
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

const concurrency = 10

func TestSinglePostgres(t *testing.T) {
	t.Parallel()

	for index := range concurrency {
		// Run multiple single postgresses in parallel
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			suite.Run(t, new(SinglePostgresTestSuite))
		})
	}
}
