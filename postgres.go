package dbsuite

import (
	"database/sql"

	_ "github.com/lib/pq" // Ensure postgres is available
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

	s.cancelSuite = func() error {
		return container.Terminate(t.Context())
	}
}

// SetupTest will create a new database from the test name and set the cancel function to drop it later
func (s *Postgres) SetupTest() {
	t := s.T()

	db, cancel, err := ensureDatabase(t.Context(), t.Name(), s.suiteConnectionString)
	require.NoError(t, err)

	s.DB = db
	s.cancel = cancel
}

// TearDownTest will close the connection and call the cancel function
func (s *Postgres) TearDownTest() {
	t := s.T()

	// DB is closed from the cancel function in
	err := s.cancel()
	require.NoError(t, err)
}

// TearDownSuite will close the top-level connection and call the suite cancel function
func (s *Postgres) TearDownSuite() {
	t := s.T()

	err := s.cancelSuite()
	require.NoError(t, err)
}
