package dbsuite

import (
	"database/sql"
	"sync"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	// postgresMutex is used to ensure there's no concurrency issues when starting
	// or shutting down the singlePostgres instance
	postgresMutex sync.Mutex

	// postgresUsageCounter keeps track of how many suites are currently using the singleton. It must
	// only be altered under the mutex.
	postgresUsageCounter int

	// singlePostgres is the single postgres test suite
	singlePostgres *postgrescontainer.PostgresContainer
)

// SinglePostgres is an alternative of the Postgres suite that is intended for codebases that feature
// large amounts of db suite tests. Instead of spinning up a container for each of those suites, SinglePostgres uses
// a single global suite that is instrumented through a mutex.
type SinglePostgres struct {
	suite.Suite

	// ContainerImage can be set before a suite is run. Defaults to postgres:alpine. This value should not differ
	// between implementations, as only the first
	ContainerImage string

	// DB is the connection to a test-specific database. After the test the database is dropped.
	DB *sql.DB

	// cancel must be executed in TearDownTest to drop the created database and close the connection
	cancel func() error

	// suiteConnectionString is used to determine the test-specific connection string
	suiteConnectionString string

	// cancelSuite will cancel the entire suite and terminate the container
	cancelSuite func() error
}

func (s *SinglePostgres) SetupSuite() {
	t := s.T()

	postgresMutex.Lock()
	defer postgresMutex.Unlock()

	// Ensure we that a new suite was requested
	postgresUsageCounter++

	if s.ContainerImage == "" {
		s.ContainerImage = "postgres:alpine"
	}

	var err error

	// If no instance is in use, spin up a container
	if singlePostgres == nil {
		singlePostgres, err = postgrescontainer.Run(t.Context(), s.ContainerImage, postgrescontainer.BasicWaitStrategies())
		require.NoError(t, err)
	}

	s.suiteConnectionString = singlePostgres.MustConnectionString(t.Context(), "sslmode=disable")

	s.cancelSuite = func() error {
		return singlePostgres.Terminate(t.Context())
	}
}

// SetupTest will create a new database from the test name and set the cancel function to drop it later
func (s *SinglePostgres) SetupTest() {
	t := s.T()

	db, cancel, err := ensureDatabase(t.Context(), t.Name(), s.suiteConnectionString)
	require.NoError(t, err)

	s.DB = db
	s.cancel = cancel
}

// TearDownTest will close the connection and call the cancel function
func (s *SinglePostgres) TearDownTest() {
	t := s.T()

	// DB is closed from the cancel function in
	err := s.cancel()
	require.NoError(t, err)
}

func (s *SinglePostgres) TearDownSuite() {
	t := s.T()

	postgresMutex.Lock()
	defer postgresMutex.Unlock()

	// First, see how many instances are currently using the singleton and remove ourselves from
	// that number. This ensures all other instances running the teardown are aware that we're exiting.
	postgresUsageCounter--

	// If there are still other goroutines, skip any shutdown attempts and leave that to the others.
	if postgresUsageCounter > 0 {
		return
	}

	// Commence shutdown sequence
	err := s.cancelSuite()
	require.NoError(t, err)

	// Reset container
	singlePostgres = nil
}
