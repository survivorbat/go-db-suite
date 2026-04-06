package dbsuite

import (
	"database/sql"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Postgres struct {
	suite.Suite

	// DB is the connection to a test-specific database. After the test the database is dropped.
	DB *sql.DB

	// ContainerImage can be set before a suite is run. Defaults to postgres:alpine.
	ContainerImage string

	// cancel must be executed in TearDownTest to drop the created database and close the connection
	cancel func() error

	// suiteConnectionString is used to determine the test-specific connection string
	suiteConnectionString string

	// suiteDB is the top-level connection used to create and drop databases
	suiteDB *sql.DB

	// cancelSuite will cancel the entire suite and terminate the container
	cancelSuite func() error
}

// SetupSuite will spin up a postgres container and set the struct variables
func (s *Postgres) SetupSuite() {
	t := s.T()

	if s.ContainerImage == "" {
		s.ContainerImage = "postgres:alpine"
	}

	container, err := postgrescontainer.Run(t.Context(), s.ContainerImage, postgrescontainer.BasicWaitStrategies())
	require.NoError(t, err)

	s.suiteConnectionString = container.MustConnectionString(t.Context(), "sslmode=disable")

	s.suiteDB, err = sql.Open("postgres", s.suiteConnectionString)
	require.NoError(t, err)

	s.cancelSuite = func() error {
		return container.Terminate(t.Context())
	}
}

// SetupTest will create a new database from the test name and set the cancel function to drop it later
func (s *Postgres) SetupTest() {
	t := s.T()

	dbName := toSafeDBName(t.Name())

	_, err := s.suiteDB.ExecContext(t.Context(), "CREATE DATABASE "+dbName+";")
	require.NoError(t, err)

	dbConnectionString := strings.ReplaceAll(s.suiteConnectionString, "/postgres?", "/"+dbName+"?")

	s.DB, err = sql.Open("postgres", dbConnectionString)
	require.NoError(t, err)

	s.cancel = func() error {
		_, err := s.suiteDB.ExecContext(t.Context(), "DROP DATABASE "+dbName+" WITH (FORCE)")
		return err
	}
}

// TearDownTest will close the connection and call the cancel function
func (s *Postgres) TearDownTest() {
	t := s.T()

	err := s.DB.Close()
	require.NoError(t, err)

	err = s.cancel()
	require.NoError(t, err)
}

// TearDownSuite will close the top-level connection and call the suite cancel function
func (s *Postgres) TearDownSuite() {
	t := s.T()

	err := s.suiteDB.Close()
	require.NoError(t, err)

	err = s.cancelSuite()
	require.NoError(t, err)
}
